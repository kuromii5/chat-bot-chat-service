package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

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

	msg := kafka.Message{
		Topic: NotificationsTopic,
		Key:   []byte(event.RecipientID.String()),
		Value: data,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write kafka message: %w", err)
	}
	return nil
}
