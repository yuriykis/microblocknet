package main

import (
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/yuriykis/microblocknet/common/messages"
	"github.com/yuriykis/microblocknet/gateway/network"
	"go.uber.org/zap"
)

type KafkaConsumer struct {
	consumer *kafka.Consumer
	logger   *zap.SugaredLogger
	network  network.Networker
	quitCh   chan struct{}
}

func NewKafkaConsumer(topics []string, logger *zap.SugaredLogger) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}
	c.SubscribeTopics(topics, nil)
	return &KafkaConsumer{
		consumer: c,
		logger:   logger,
	}, nil
}

func (c *KafkaConsumer) Start(network network.Networker) {
	c.network = network
	go c.readMessageLoop()
}

func (c *KafkaConsumer) Stop() {
	close(c.quitCh)
	c.consumer.Close()
}

func (c *KafkaConsumer) readMessageLoop() {
	c.logger.Info("starting kafka consumer")
loop:
	for {
		select {
		case <-c.quitCh:
			return
		default:
			msg, err := c.consumer.ReadMessage(-1)
			if err != nil {
				c.logger.Errorf("Consumer error: %v (%v)\n", err, msg)
				continue loop
			}
			var request messages.RegisterNodeMessage
			if err := json.Unmarshal(msg.Value, &request); err != nil {
				c.logger.Errorf("failed to unmarshal message: %v", err)
				continue loop
			}
			c.logger.Infof("received message: %v", request)
			if err := c.network.NewPeer(request.Address); err != nil {
				c.logger.Errorf("failed to register node: %v", err)
				continue loop
			}
			c.logger.Infof("node %s registered", request.Address)
		}

	}
}
