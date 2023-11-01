package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/yuriykis/microblocknet/common/requests"
	"github.com/yuriykis/microblocknet/node/service/client"
	grpcPeer "google.golang.org/grpc/peer"
)

// the api server is used to expose the node's functionality the gateway
type ApiServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type apiServer struct {
	apiListenAddr string
	httpServer    *http.Server
	grpcClient    *client.GRPCClient
	dr            DataRetriever
}

func NewApiServer(dr DataRetriever, grpcListenAddress string, apiListenAddr string) (*apiServer, error) {
	httpServer := &http.Server{
		Addr: apiListenAddr,
	}
	grpcClient, err := client.NewGRPCClient(grpcListenAddress)
	if err != nil {
		return nil, err
	}
	return &apiServer{
		apiListenAddr: apiListenAddr,
		httpServer:    httpServer,
		grpcClient:    grpcClient,
		dr:            dr,
	}, nil
}

func (s *apiServer) Start(ctx context.Context) error {
	s.httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/block":
			makeHTTPHandlerFunc(handleGetBlockByHeight(s.dr))(w, r)
		case "/utxo":
			makeHTTPHandlerFunc(handleGetUTXOsByAddress(s.dr))(w, r)
		case "/transaction":
			makeHTTPHandlerFunc(handleNewTransaction(s.dr, s.grpcClient))(w, r)
		case "/height":
			makeHTTPHandlerFunc(handleGetCurrentHeight(s.dr))(w, r)
		case "/healthcheck":
			makeHTTPHandlerFunc(handleHealthCheck())(w, r)
		default:
			writeJSON(
				w,
				http.StatusNotFound,
				map[string]string{"error": "not found"},
			)
		}
	})
	fmt.Printf("API server listening on %s\n", s.apiListenAddr)

	return s.httpServer.ListenAndServe()
}

func (s *apiServer) Stop(ctx context.Context) error {
	return s.httpServer.Close()
}

type HTTPFunc func(w http.ResponseWriter, r *http.Request) error

type APIError struct {
	Code int
	Err  error
}

func (e APIError) Error() string {
	return e.Err.Error()
}

func writeJSON(rw http.ResponseWriter, status int, v any) error {
	rw.WriteHeader(status)
	rw.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(rw).Encode(v)
}

func makeHTTPHandlerFunc(fn HTTPFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			fmt.Println(err)
			if apiErr, ok := err.(APIError); ok {
				writeJSON(w, apiErr.Code, apiErr.Err.Error())
				return
			}
			writeJSON(
				w,
				http.StatusInternalServerError,
				map[string]string{"error": "internal server error"},
			)
		}
	}
}

func handleGetBlockByHeight(dr DataRetriever) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		req := requests.GetBlockByHeightRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("failed to decode request body: %w", err),
			}
		}
		block, err := dr.GetBlockByHeight(context.Background(), req.Height)
		if err != nil {
			return APIError{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("failed to get block by height: %w", err),
			}
		}
		return writeJSON(w, http.StatusOK, requests.GetBlockByHeightResponse{
			Block: block,
		})

	}
}

func handleGetUTXOsByAddress(dr DataRetriever) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		req := requests.GetUTXOsByAddressRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fmt.Println(err)
			return APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("failed to decode request body: %w", err),
			}
		}
		utxos, err := dr.GetUTXOsByAddress(context.Background(), req.Address)
		if err != nil {
			fmt.Println(err)
			return APIError{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("failed to get utxos by address: %w", err),
			}
		}
		return writeJSON(w, http.StatusOK, requests.GetUTXOsByAddressResponse{
			UTXOs: utxos,
		})
	}
}

func handleNewTransaction(dr DataRetriever, c *client.GRPCClient) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		req := requests.NewTransactionRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fmt.Println(err)
			return APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("failed to decode request body: %w", err),
			}
		}
		ctx := grpcPeer.NewContext(context.Background(), &grpcPeer.Peer{
			Addr: &net.IPAddr{
				IP: net.ParseIP(""),
			},
		})
		tx, err := c.NewTransaction(ctx, req.Transaction)
		if err != nil {
			fmt.Println(err)
			return APIError{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("failed to get utxos by address: %w", err),
			}
		}

		return writeJSON(w, http.StatusOK, requests.NewTransactionResponse{
			Transaction: tx,
		})
	}
}

func handleGetCurrentHeight(dr DataRetriever) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		height := dr.Chain().headers.Height()
		return writeJSON(w, http.StatusOK, requests.GetCurrentHeightResponse{
			Height: height,
		})
	}
}

func handleHealthCheck() HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		return writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
