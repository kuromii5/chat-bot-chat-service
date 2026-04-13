package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

// kafkaHeaderCarrier implements propagation.TextMapCarrier for injecting
// trace context into Kafka message headers.
type kafkaHeaderCarrier []kafka.Header

func (c *kafkaHeaderCarrier) Get(_ string) string { return "" }

func (c *kafkaHeaderCarrier) Set(key, value string) {
	*c = append(*c, kafka.Header{Key: key, Value: []byte(value)})
}

func (c *kafkaHeaderCarrier) Keys() []string {
	keys := make([]string, len(*c))
	for i, h := range *c {
		keys[i] = h.Key
	}
	return keys
}

var _ propagation.TextMapCarrier = (*kafkaHeaderCarrier)(nil)

// NotificationEvent is the message published to Kafka for the notification-service to consume.
type NotificationEvent struct {
	ID          uuid.UUID        `json:"id"`
	Type        domain.EventType `json:"type"`
	RecipientID uuid.UUID        `json:"recipient_id"`
	RoomID      uuid.UUID        `json:"room_id"`
	SenderID    uuid.UUID        `json:"sender_id"`
	Text        string           `json:"text"`
	OccurredAt  time.Time        `json:"occurred_at"`
}

func (p *Producer) PublishNotification(ctx context.Context, event NotificationEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal notification event: %w", err)
	}

	var carrier kafkaHeaderCarrier
	otel.GetTextMapPropagator().Inject(ctx, &carrier)

	msg := kafka.Message{
		Topic:   NotificationsTopic,
		Key:     []byte(event.RecipientID.String()),
		Value:   data,
		Headers: carrier,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write kafka message: %w", err)
	}
	return nil
}
