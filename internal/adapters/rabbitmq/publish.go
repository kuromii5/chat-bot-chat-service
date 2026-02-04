package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/internal/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *RabbitMQ) PublishNewQuestion(ctx context.Context, msg *domain.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	routingKey := fmt.Sprintf("question.%s", strings.Join(msg.Tags, "."))
	return r.channel.PublishWithContext(ctx,
		r.config.Exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (r *RabbitMQ) Listen(ctx context.Context, userID uuid.UUID, handler ports.MessageHandler) error {
	queueName := fmt.Sprintf("ai_queue_%s", userID.String())

	msgs, err := r.channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}

				if err := handler(ctx, d.Body); err == nil {
					d.Ack(false)
				} else {
					d.Nack(false, true)
				}
			}
		}
	}()

	return nil
}
