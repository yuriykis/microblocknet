package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/yuriykis/microblocknet/common/requests"
	"go.uber.org/zap"
)

const gatewayPingInterval = 2 * time.Second

type gatewayClient struct {
	Endpoint  string
	connected bool
	logger    *zap.SugaredLogger
}

func NewGatewayClient(endpoint string, logger *zap.SugaredLogger) *gatewayClient {
	return &gatewayClient{
		Endpoint:  endpoint,
		logger:    logger,
		connected: false,
	}
}

func (c *gatewayClient) Healthcheck(ctx context.Context) bool {
	endpoint := c.Endpoint + "/healthcheck"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		c.logger.Errorf("failed to create request: %v", err)
		return false
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		c.logger.Errorf("failed to send request: %v", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("gateway returned status code %d", resp.StatusCode)
		return false
	}

	return true
}

func (c *gatewayClient) RegisterMe(ctx context.Context, addr string) (requests.RegisterNodeResponse, error) {
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

func (c *gatewayClient) pingGatewayLoop(quitCh chan struct{}, myAddr string) {
ping:
	for {
		select {
		case <-quitCh:
			return
		case <-time.After(gatewayPingInterval):
			ok := c.Healthcheck(context.Background())
			if !ok {
				c.connected = false
				c.logger.Errorf("failed to ping gateway")
				continue ping
			}
			if !c.connected {
				_, err := c.RegisterMe(context.Background(), myAddr)
				if err != nil {
					c.logger.Errorf("failed to register with gateway: %v", err)
					continue ping
				}
				c.connected = true
			}
			c.logger.Infof("successfully pinged gateway")
		}
	}
}
