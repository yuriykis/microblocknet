package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

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
		height := r.URL.Query().Get("height")
		if height == "" {
			return errors.New("height is required")
		}
		h, err := strconv.Atoi(height)
		if err != nil {
			return err
		}
		b, err := svc.GetBlockByHeight(h)
		if err != nil {
			return err
		}
		return json.NewEncoder(w).Encode(b)
	}
}
