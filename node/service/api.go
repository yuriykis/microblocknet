package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yuriykis/microblocknet/node/service/types"
)

// the api server is used to expose the node's functionality the gateway
type ApiServer interface {
	Start() error
	Stop() error
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

func (s *apiServer) Start() error {
	s.httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/block":
			makeHTTPHandlerFunc(handleGetBlockByHeight(s.svc))(w, r)
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

func (s *apiServer) Stop() error {
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
		req := types.GetBlockByHeightRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("failed to decode request body: %w", err),
			}
		}
		block, err := svc.GetBlockByHeight(req.Height)
		if err != nil {
			return APIError{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("failed to get block by height: %w", err),
			}
		}
		return writeJSON(w, http.StatusOK, types.GetBlockByHeightResponse{
			Block: block,
		})

	}
}
