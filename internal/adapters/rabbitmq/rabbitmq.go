package rabbitmq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"

	"github.com/kuromii5/chat-bot-chat-service/config"
)

const MaxRetries = 5

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  config.RabbitMQConfig
}

func New(cfg config.RabbitMQConfig) (*RabbitMQ, error) {
	var conn *amqp.Connection
	var err error
	for i := range MaxRetries {
		conn, err = amqp.Dial(cfg.URL)
		if err == nil {
			break
		}
		logrus.WithError(err).Errorf("RabbitMQ not ready, retrying in 2s... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		cfg.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq exchange declare: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
		config:  cfg,
	}, nil
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return fmt.Errorf("close channel: %w", err)
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return fmt.Errorf("close connection: %w", err)
		}
	}
	return nil
}
