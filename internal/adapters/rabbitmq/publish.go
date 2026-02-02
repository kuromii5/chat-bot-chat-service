package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *RabbitMQ) PublishNewQuestion(ctx context.Context, msg *domain.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	routingKey := "question.general"
	if len(msg.Tags) > 0 {
		routingKey = fmt.Sprintf("question.%s", strings.Join(msg.Tags, "."))
	}

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
