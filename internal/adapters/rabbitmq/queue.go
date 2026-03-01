package rabbitmq

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

func (r *RabbitMQ) BindRoomToAI(ctx context.Context, roomID uuid.UUID, aiID uuid.UUID) error {
	queueName := fmt.Sprintf("ai_queue_%s", aiID.String())
	routingKey := fmt.Sprintf("room.%s", roomID.String())

	if err := r.channel.QueueBind(queueName, routingKey, r.config.Exchange, false, nil); err != nil {
		return fmt.Errorf("queue bind room: %w", err)
	}
	return nil
}

func (r *RabbitMQ) SyncAIQueue(
	ctx context.Context,
	aiID uuid.UUID,
	tags, oldTags []string,
) error {
	queueName := fmt.Sprintf("ai_queue_%s", aiID.String())

	if _, err := r.channel.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return fmt.Errorf("queue declare: %w", err)
	}

	for _, tag := range oldTags {
		routingKey := fmt.Sprintf("question.#.%s.#", tag)
		if err := r.channel.QueueUnbind(queueName, routingKey, r.config.Exchange, nil); err != nil {
			return fmt.Errorf("queue unbind: %w", err)
		}
	}

	for _, tag := range tags {
		routingKey := fmt.Sprintf("question.#.%s.#", tag)
		if err := r.channel.QueueBind(
			queueName,
			routingKey,
			r.config.Exchange,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("queue bind: %w", err)
		}
	}

	return nil
}

func (r *RabbitMQ) ListenReplies(
	ctx context.Context,
	userID uuid.UUID,
	handler func(ctx context.Context, body []byte) error,
) error {
	queueName := fmt.Sprintf("human_queue_%s", userID.String())
	routingKey := fmt.Sprintf("reply.%s", userID.String())

	if _, err := r.channel.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return fmt.Errorf("queue declare: %w", err)
	}
	if err := r.channel.QueueBind(queueName, routingKey, r.config.Exchange, false, nil); err != nil {
		return fmt.Errorf("queue bind: %w", err)
	}

	msgs, err := r.channel.Consume(queueName, "", false, false, false, false, nil)
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
				msgCtx, span := otel.Tracer("rabbitmq").Start(msgCtx, "rabbitmq.deliver_reply")
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
