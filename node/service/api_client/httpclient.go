package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/yuriykis/microblocknet/common/requests"
)

type HTTPClient struct {
	Endpoint string
}

func NewHTTPClient(endpoint string) *HTTPClient {
	return &HTTPClient{
		Endpoint: endpoint,
	}
}

func (c *HTTPClient) GetBlockByHeight(ctx context.Context, height int) (*requests.GetBlockByHeightResponse, error) {
	cReq := requests.GetBlockByHeightRequest{
		Height: height,
	}
	b, err := json.Marshal(&cReq)
	if err != nil {
		return nil, err
	}
	endpoint := c.Endpoint + "/block"
	req, err := http.NewRequest("GET", endpoint, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var cResp requests.GetBlockByHeightResponse
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		return nil, err
	}
	return &cResp, nil
}

func (c *HTTPClient) GetUTXOsByAddress(
	ctx context.Context,
	address []byte,
) (*requests.GetUTXOsByAddressResponse, error) {
	cReq := requests.GetUTXOsByAddressRequest{
		Address: address,
	}
	b, err := json.Marshal(&cReq)
	if err != nil {
		return nil, err
	}
	endpoint := c.Endpoint + "/utxo"
	req, err := http.NewRequest("GET", endpoint, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var cResp requests.GetUTXOsByAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		return nil, err
	}
	return &cResp, nil
}

func (c *HTTPClient) PeersAddrs(ctx context.Context) []string {
	// TODO: implement
	return nil
}
