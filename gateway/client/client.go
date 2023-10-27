package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/yuriykis/microblocknet/common/requests"
)

type Client interface {
	InitTransaction(ctx context.Context, tReq requests.CreateTransactionRequest) error
	GetMyUTXOs(ctx context.Context) error
	GetBlockByHeight(ctx context.Context, height int) (*requests.GetBlockByHeightResponse, error)
	GetUTXOsByAddress(ctx context.Context) error
}

type HTTPClient struct {
	endpoint string
}

func NewHTTPClient(endpoint string) *HTTPClient {
	return &HTTPClient{
		endpoint: endpoint,
	}
}

func (c *HTTPClient) InitTransaction(ctx context.Context, tReq requests.CreateTransactionRequest) error {
	return nil
}

func (c *HTTPClient) GetMyUTXOs(ctx context.Context) error {
	return nil
}

func (c *HTTPClient) GetUTXOsByAddress(ctx context.Context) error {
	return nil
}

func (c *HTTPClient) GetBlockByHeight(ctx context.Context, height int) (*requests.GetBlockByHeightResponse, error) {
	cReq := requests.GetBlockByHeightRequest{
		Height: height,
	}
	b, err := json.Marshal(&cReq)
	if err != nil {
		return nil, err
	}

	endpoint := c.endpoint + "/block"
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
