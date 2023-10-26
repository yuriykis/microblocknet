package main

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"

	"github.com/yuriykis/microblocknet/common/requests"
	apiclient "github.com/yuriykis/microblocknet/node/service/api_client"
)

func main() {
	listenAddr := flag.String("listen-addr", ":6000", "The address to listen on for incoming HTTP requests")
	flag.Parse()

	// nodesAddrs := []string{"node1:3000", "node2:3001", "node3:3002"}

	h := NewApiHandler()
	http.HandleFunc("/block", MakeAPIFunc(h.handleGetBlockByHeight))
	http.HandleFunc("/utxo", MakeAPIFunc(h.handleGetUTXOsByAddress))

	http.ListenAndServe(*listenAddr, nil)
}

type apiHandler struct {
	client apiclient.Client
}

func NewApiHandler() *apiHandler {
	client := apiclient.NewHTTPClient("http://localhost:4001") // hardcoded for now
	return &apiHandler{
		client: client,
	}
}

func (h *apiHandler) handleGetBlockByHeight(w http.ResponseWriter, r *http.Request) error {
	var req requests.GetBlockByHeightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	res, err := h.client.GetBlockByHeight(context.Background(), req.Height)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, res)
}

func (h *apiHandler) handleGetUTXOsByAddress(w http.ResponseWriter, r *http.Request) error {
	var req requests.GetUTXOsByAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	res, err := h.client.GetUTXOsByAddress(context.Background(), req.Address)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, res.UTXOs)
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func MakeAPIFunc(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			writeJSON(
				w,
				http.StatusInternalServerError,
				map[string]string{"error": err.Error()},
			)
		}
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}
