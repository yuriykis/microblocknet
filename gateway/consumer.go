package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/yuriykis/microblocknet/common/messages"
)

type KafkaConsumer struct {
	consumer *kafka.Consumer
	service  *service
	quitCh   chan struct{}
}

func NewKafkaConsumer(topics []string) (*KafkaConsumer, error) {
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
	}, nil
}

func (c *KafkaConsumer) Start(service *service) {
	c.service = service
	go c.readMessageLoop()
}

func (c *KafkaConsumer) Stop() {
	close(c.quitCh)
	c.consumer.Close()
}

func (c *KafkaConsumer) readMessageLoop() {
loop:
	for {
		select {
		case <-c.quitCh:
			return
		default:
			msg, err := c.consumer.ReadMessage(-1)
			if err != nil {
				fmt.Printf("Consumer error: %v (%v)\n", err, msg)
				continue loop
			}
			var request messages.RegisterNodeMessage
			if err := json.Unmarshal(msg.Value, &request); err != nil {
				fmt.Printf("failed to unmarshal request: %v", err)
				continue loop
			}
			fmt.Printf("Received message: %v\n", request)
			if err := c.service.NewNode(context.TODO(), request.Address); err != nil {
				fmt.Printf("failed to register node: %v", err)
				continue loop
			}
		}

	}
}
