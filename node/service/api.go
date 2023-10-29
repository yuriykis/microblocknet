package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yuriykis/microblocknet/common/requests"
)

// the api server is used to expose the node's functionality the gateway
type ApiServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type apiServer struct {
	apiListenAddr string
	httpServer    *http.Server
	svc           Service
}

func NewApiServer(svc Service, apiListenAddr string) *apiServer {
	httpServer := &http.Server{
		Addr: apiListenAddr,
	}
	return &apiServer{
		apiListenAddr: apiListenAddr,
		httpServer:    httpServer,
		svc:           svc,
	}
}

func (s *apiServer) Start(ctx context.Context) error {
	s.httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/block":
			makeHTTPHandlerFunc(handleGetBlockByHeight(s.svc))(w, r)
		case "/utxo":
			makeHTTPHandlerFunc(handleGetUTXOsByAddress(s.svc))(w, r)
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

func handleGetBlockByHeight(svc Service) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		req := requests.GetBlockByHeightRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("failed to decode request body: %w", err),
			}
		}
		block, err := svc.GetBlockByHeight(context.Background(), req.Height)
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

func handleGetUTXOsByAddress(svc Service) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		req := requests.GetUTXOsByAddressRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fmt.Println(err)
			return APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("failed to decode request body: %w", err),
			}
		}
		utxos, err := svc.GetUTXOsByAddress(context.Background(), req.Address)
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
