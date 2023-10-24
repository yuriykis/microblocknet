package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/yuriykis/microblocknet/node/proto"
	"github.com/yuriykis/microblocknet/node/service/types"
)

type HTTPClient struct {
	Endpoint string
}

func NewHTTPClient(endpoint string) *HTTPClient {
	return &HTTPClient{
		Endpoint: endpoint,
	}
}

func (c *HTTPClient) GetBlockByHeight(ctx context.Context, height int) (*proto.Block, error) {
	cReq := types.GetBlockByHeightRequest{
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
	var cResp types.GetBlockByHeightResponse
	if err := json.NewDecoder(resp.Body).Decode(&cResp); err != nil {
		return nil, err
	}
	return cResp.Block, nil
}

func (c *HTTPClient) GetUTXOsByAddress(address []byte) ([]*proto.UTXO, error) {
	// TODO: implement
	return nil, nil
}

func (c *HTTPClient) PeersAddrs() []string {
	// TODO: implement
	return nil
}
