package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yuriykis/microblocknet/common/requests"
	"github.com/yuriykis/microblocknet/node/client"
	"github.com/yuriykis/microblocknet/node/service"
	grpcPeer "google.golang.org/grpc/peer"
)

func StartApiTrasport(s *ApiNodeServer) error {
	s.httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/block":
			makeHTTPHandlerFunc(handleGetBlockByHeight(s.node))(w, r)
		case "/utxo":
			makeHTTPHandlerFunc(handleGetUTXOsByAddress(s.node))(w, r)
		case "/transaction":
			makeHTTPHandlerFunc(handleNewTransaction(s.node, s.grpcClient))(w, r)
		case "/height":
			makeHTTPHandlerFunc(handleGetCurrentHeight(s.node))(w, r)
		case "/healthcheck":
			makeHTTPHandlerFunc(handleHealthCheck(s.node))(w, r)
		case "/metrics":
			promhttp.Handler().ServeHTTP(w, r)
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

func StopApiTransport(s *ApiNodeServer) error {
	return s.httpServer.Close()
}

type ApiNodeServer struct {
	apiListenAddr string
	httpServer    *http.Server
	grpcClient    *client.GRPCClient
	node          service.Api
}

func NewApiServer(
	grpcListenAddress string,
	apiListenAddr string,
	node service.Api,
) (*ApiNodeServer, error) {
	httpServer := &http.Server{
		Addr: apiListenAddr,
	}
	grpcClient, err := client.NewGRPCClient(grpcListenAddress)
	if err != nil {
		return nil, err
	}
	return &ApiNodeServer{
		apiListenAddr: apiListenAddr,
		httpServer:    httpServer,
		grpcClient:    grpcClient,
		node:          node,
	}, nil
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

func handleGetBlockByHeight(node service.Api) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		req := requests.GetBlockByHeightRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("failed to decode request body: %w", err),
			}
		}
		block, err := node.Chain().GetBlockByHeight(req.Height)
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

func handleGetUTXOsByAddress(node service.Api) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		req := requests.GetUTXOsByAddressRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fmt.Println(err)
			return APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("failed to decode request body: %w", err),
			}
		}

		// TODO: ctx generated by copilot, check if this can be useful
		ctx := grpcPeer.NewContext(context.Background(), &grpcPeer.Peer{
			Addr: &net.IPAddr{
				IP: net.ParseIP(""),
			},
		})

		utxos, err := node.Chain().Store().UTXOStore(ctx).GetByAddress(ctx, req.Address)
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

func handleNewTransaction(node service.Api, c *client.GRPCClient) HTTPFunc {
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

func handleGetCurrentHeight(node service.Api) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		height := node.Chain().Height()
		return writeJSON(w, http.StatusOK, requests.GetCurrentHeightResponse{
			Height: height,
		})
	}
}

func handleHealthCheck(node service.Api) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		node.Gate().SetConnected(true)
		return writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
