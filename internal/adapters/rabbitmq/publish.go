package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

// amqpHeaderCarrier adapts amqp.Table to propagation.TextMapCarrier.
type amqpHeaderCarrier amqp.Table

func (c amqpHeaderCarrier) Get(key string) string {
	v, _ := (amqp.Table)(c)[key].(string)
	return v
}

func (c amqpHeaderCarrier) Set(key, val string) { (amqp.Table)(c)[key] = val }

func (c amqpHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

func (r *RabbitMQ) PublishAIReply(ctx context.Context, humanID uuid.UUID, msg *domain.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	headers := amqp.Table{}
	otel.GetTextMapPropagator().Inject(ctx, amqpHeaderCarrier(headers))

	routingKey := fmt.Sprintf("reply.%s", humanID.String())
	if err := r.channel.PublishWithContext(ctx,
		r.config.Exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Headers:      headers,
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	return nil
}

func (r *RabbitMQ) PublishNewQuestion(ctx context.Context, msg *domain.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	headers := amqp.Table{}
	otel.GetTextMapPropagator().Inject(ctx, amqpHeaderCarrier(headers))

	routingKey := fmt.Sprintf("question.%s", strings.Join(msg.Tags, "."))
	if err := r.channel.PublishWithContext(ctx,
		r.config.Exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Headers:      headers,
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	return nil
}

func (r *RabbitMQ) Listen(
	ctx context.Context,
	userID uuid.UUID,
	handler func(ctx context.Context, body []byte) error,
) error {
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

				msgCtx := otel.GetTextMapPropagator().Extract(ctx, amqpHeaderCarrier(d.Headers))
				msgCtx, span := otel.Tracer("rabbitmq").Start(msgCtx, "rabbitmq.deliver")

				if err := handler(msgCtx, d.Body); err == nil {
					span.End()
					if ackErr := d.Ack(false); ackErr != nil {
						logrus.WithError(ackErr).Error("failed to ack message")
					}
				} else {
					span.RecordError(err)
					span.End()
					if nackErr := d.Nack(false, true); nackErr != nil {
						logrus.WithError(nackErr).Error("failed to nack message")
					}
				}
			}
		}
	}()

	return nil
}
