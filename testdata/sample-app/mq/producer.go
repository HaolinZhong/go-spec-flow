package mq

import (
	"context"
	"fmt"
)

type Producer struct {
	topic string
}

func NewProducer(topic string) *Producer {
	return &Producer{topic: topic}
}

func (p *Producer) SendMessage(ctx context.Context, key string, value []byte) error {
	fmt.Printf("sending message to topic=%s key=%s\n", p.topic, key)
	return nil
}
