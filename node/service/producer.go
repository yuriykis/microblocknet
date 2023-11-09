package service

import (
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/yuriykis/microblocknet/common/messages"
)

type MessageProducer interface {
	ProduceMessage(request any) error
}

type KafkaMessageProducer struct {
	producer *kafka.Producer
}

func NewKafkaMessageProducer() (MessageProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		return nil, err
	}
	return &KafkaMessageProducer{
		producer: p,
	}, nil
}

// message should a type from message package in common module
func (p *KafkaMessageProducer) ProduceMessage(msg any) error {
	var topic string
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	switch msg.(type) {
	case messages.RegisterNodeMessage:
		topic = "register_node"
	default:
		return fmt.Errorf("unknown msg type")
	}
	return p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: msgBytes,
	}, nil)
}
