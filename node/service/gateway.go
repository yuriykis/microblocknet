package service

import (
	"sync"
	"time"

	"github.com/yuriykis/microblocknet/common/messages"
	"go.uber.org/zap"
)

const (
	gatewayRegisterInterval = 10 * time.Second
	keepConnectedInterval   = 5 * time.Second
)

type gatewayClient struct {
	Endpoint  string
	connected bool
	timer     *time.Timer
	logger    *zap.SugaredLogger
	mp        MessageProducer
	mu        *sync.Mutex
}

func NewGatewayClient(endpoint string, logger *zap.SugaredLogger) *gatewayClient {
	mp, err := NewKafkaMessageProducer()
	if err != nil {
		logger.Errorf("failed to create message producer: %v", err)
	}
	return &gatewayClient{
		Endpoint:  endpoint,
		logger:    logger,
		connected: false,
		mp:        mp,
		mu:        &sync.Mutex{},
	}
}

// SetConnected sets the connected flag and starts a timer to reset it
func (c *gatewayClient) SetConnected(connected bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if connected {
		if c.timer != nil {
			c.timer.Stop()
		}
		c.timer = time.AfterFunc(keepConnectedInterval, func() {
			c.mu.Lock()
			c.connected = false
			c.mu.Unlock()
		})
	} else {
		if c.timer != nil {
			c.timer.Stop()
			c.timer = nil
		}
	}
	c.connected = connected
}

func (c *gatewayClient) RegisterMe(addr string) error {
	rMsg := messages.RegisterNodeMessage{
		Address: addr,
	}
	if err := c.mp.ProduceMessage(rMsg); err != nil {
		return err
	}
	return nil
}

func (c *gatewayClient) registerGatewayLoop(quitCh chan struct{}, myAddr string) {
ping:
	for {
		select {
		case <-quitCh:
			return
		case <-time.After(gatewayRegisterInterval):
			if !c.connected {
				err := c.RegisterMe(myAddr)
				if err != nil {
					c.logger.Errorf("failed to register with gateway: %v", err)
					continue ping
				}
			}
			c.logger.Infof("successfully pinged gateway, address: %s", c.Endpoint)
		}
	}
}
