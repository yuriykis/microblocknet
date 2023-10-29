package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/yuriykis/microblocknet/common/requests"
)

type Client interface {
	InitTransaction(
		ctx context.Context,
		tReq requests.InitTransactionRequest,
	) (*requests.InitTransactionResponse, error)
	GetMyUTXOs(ctx context.Context) error
	GetBlockByHeight(ctx context.Context, height int) (*requests.GetBlockByHeightResponse, error)
	GetUTXOsByAddress(ctx context.Context) error
	NewTransaction(ctx context.Context, tReq requests.NewTransactionRequest) (*requests.NewTransactionResponse, error)
}

type HTTPClient struct {
	endpoint string
}

func NewHTTPClient(endpoint string) *HTTPClient {
	return &HTTPClient{
		endpoint: endpoint,
	}
}

func (c *HTTPClient) InitTransaction(
	ctx context.Context,
	tReq requests.InitTransactionRequest,
) (*requests.InitTransactionResponse, error) {
	endpoint := c.endpoint + "/transaction/init"
	b, err := json.Marshal(&tReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var cResp requests.InitTransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		return nil, err
	}

	return &cResp, nil
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

func (c *HTTPClient) NewTransaction(
	ctx context.Context,
	tReq requests.NewTransactionRequest,
) (*requests.NewTransactionResponse, error) {
	endpoint := c.endpoint + "/transaction"
	b, err := json.Marshal(&tReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var cResp requests.NewTransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		return nil, err
	}

	return &cResp, nil
}
