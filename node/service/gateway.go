package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/yuriykis/microblocknet/common/requests"
)

type GatewayClient struct {
	Endpoint string
}

func NewGatewayClient(endpoint string) *GatewayClient {
	return &GatewayClient{
		Endpoint: endpoint,
	}
}

func (c *GatewayClient) RegisterMe(ctx context.Context, addr string) (requests.RegisterNodeResponse, error) {
	rRes := requests.RegisterNodeResponse{}
	rReq := requests.RegisterNodeRequest{
		Address: addr,
	}
	b, err := json.Marshal(&rReq)
	if err != nil {
		return rRes, err
	}
	endpoint := c.Endpoint + "/node/register"
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	if err != nil {
		return rRes, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return rRes, err
	}
	defer resp.Body.Close()
	var cResp requests.RegisterNodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		return rRes, err
	}
	return cResp, nil
}
