package kafka

import (
	"github.com/segmentio/kafka-go"
)

const NotificationsTopic = "notifications"

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
	}
	return &Producer{writer: w}
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
