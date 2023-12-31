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

func (c *HTTPClient) String() string {
	return c.Endpoint
}

func (c *HTTPClient) GetBlockByHeight(ctx context.Context, height int) (requests.GetBlockByHeightResponse, error) {
	cReq := requests.GetBlockByHeightRequest{
		Height: height,
	}
	var cResp requests.GetBlockByHeightResponse
	b, err := json.Marshal(&cReq)
	if err != nil {
		return cResp, err
	}
	endpoint := c.Endpoint + "/block"
	req, err := http.NewRequest("GET", endpoint, bytes.NewBuffer(b))
	if err != nil {
		return cResp, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return cResp, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		return cResp, err
	}
	return cResp, nil
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

func (c *HTTPClient) NewTransaction(
	ctx context.Context,
	tReq requests.NewTransactionRequest,
) (requests.NewTransactionResponse, error) {
	res := requests.NewTransactionResponse{}
	b, err := json.Marshal(&tReq)
	if err != nil {
		return res, err
	}
	endpoint := c.Endpoint + "/transaction"
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	if err != nil {
		return res, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	var cResp requests.NewTransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		return res, err
	}
	return cResp, nil
}

func (c *HTTPClient) Height(ctx context.Context) (requests.GetCurrentHeightResponse, error) {
	res := requests.GetCurrentHeightResponse{}
	endpoint := c.Endpoint + "/height"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return res, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return res, err
	}
	return res, nil
}

func (c *HTTPClient) Healthcheck(ctx context.Context) (requests.HealthcheckResponse, error) {
	res := requests.HealthcheckResponse{}
	endpoint := c.Endpoint + "/healthcheck"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return res, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return res, err
	}
	return res, nil
}
