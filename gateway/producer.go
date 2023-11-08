package main

import (
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/yuriykis/microblocknet/common/requests"
)

type RequestProducer interface {
	ProduceRequest(request any) error
}

type KafkaRequestProducer struct {
	producer *kafka.Producer
}

func NewKafkaRequestProducer() (RequestProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		return nil, err
	}
	return &KafkaRequestProducer{
		producer: p,
	}, nil
}

// request should a type from request package in common module
func (p *KafkaRequestProducer) ProduceRequest(request any) error {
	var topic string
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	switch request.(type) {
	case requests.InitTransactionRequest:
		topic = "init_transaction"
	case requests.NewTransactionRequest:
		topic = "new_transaction"
	case requests.RegisterNodeRequest:
		topic = "register_node"
	default:
		return fmt.Errorf("unknown request type")
	}
	return p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: requestBytes,
	}, nil)
}
