package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	kafkaadapter "github.com/kuromii5/chat-bot-chat-service/internal/adapters/kafka"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type OutboxRepo interface {
	FetchPending(ctx context.Context, limit int) ([]*domain.OutboxEvent, error)
	MarkPublished(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}

type Publisher interface {
	PublishNewQuestion(ctx context.Context, msg *domain.Message) error
	PublishFollowUp(ctx context.Context, roomID uuid.UUID, msg *domain.Message) error
	PublishAIReply(ctx context.Context, humanID uuid.UUID, msg *domain.Message) error
}

type QueueSyncer interface {
	SyncAIQueue(ctx context.Context, aiID uuid.UUID, tags, oldTags []string) error
}

type Binder interface {
	BindRoomToAI(ctx context.Context, roomID uuid.UUID, aiID uuid.UUID) error
}

type KafkaNotifier interface {
	PublishNotification(ctx context.Context, event kafkaadapter.NotificationEvent) error
}

const fetchLimit = 100

type Relay struct {
	repo      OutboxRepo
	publisher Publisher
	syncer    QueueSyncer
	binder    Binder
	kafka     KafkaNotifier
	interval  time.Duration
}

func NewRelay(repo OutboxRepo, publisher Publisher, syncer QueueSyncer, binder Binder, kafka KafkaNotifier, interval time.Duration) *Relay {
	return &Relay{repo: repo, publisher: publisher, syncer: syncer, binder: binder, kafka: kafka, interval: interval}
}

func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.process(ctx)
		}
	}
}

func (r *Relay) process(ctx context.Context) {
	events, err := r.repo.FetchPending(ctx, fetchLimit)
	if err != nil {
		logrus.WithError(err).Error("outbox: fetch pending failed")
		return
	}

	for _, event := range events {
		if err := r.dispatch(ctx, event); err != nil {
			logrus.WithError(err).WithField("event_id", event.ID).Error("outbox: dispatch failed")
			if markErr := r.repo.MarkFailed(ctx, event.ID, err.Error()); markErr != nil {
				logrus.WithError(markErr).WithField("event_id", event.ID).Error("outbox: mark failed error")
			}
			continue
		}
		if err := r.repo.MarkPublished(ctx, event.ID); err != nil {
			logrus.WithError(err).WithField("event_id", event.ID).Error("outbox: mark published error")
		}
	}
}

func (r *Relay) dispatch(ctx context.Context, event *domain.OutboxEvent) error {
	switch event.EventType {
	case domain.EventNewQuestion, domain.EventFollowUp, domain.EventAIReply:
		return r.dispatchMessage(ctx, event)
	case domain.EventTagsSync:
		return r.dispatchTagSync(ctx, event)
	case domain.EventRoomClaimed:
		return r.dispatchRoomClaimed(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}

func (r *Relay) dispatchMessage(ctx context.Context, event *domain.OutboxEvent) error {
	var payload domain.MessagePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	switch event.EventType {
	case domain.EventNewQuestion:
		if err := r.publisher.PublishNewQuestion(ctx, payload.Message); err != nil {
			return fmt.Errorf("PublishNewQuestion: %w", err)
		}
		return nil
	case domain.EventFollowUp:
		if err := r.publisher.PublishFollowUp(ctx, payload.Message.RoomID, payload.Message); err != nil {
			return fmt.Errorf("PublishFollowUp: %w", err)
		}
		return nil
	case domain.EventAIReply:
		if err := r.publisher.PublishAIReply(ctx, payload.HumanID, payload.Message); err != nil {
			return fmt.Errorf("PublishAIReply: %w", err)
		}
		r.notifyKafka(ctx, kafkaadapter.NotificationEvent{
			ID:          event.ID,
			Type:        event.EventType,
			RecipientID: payload.HumanID,
			RoomID:      payload.Message.RoomID,
			SenderID:    payload.Message.SenderID,
			Text:        payload.Message.Content,
			OccurredAt:  payload.Message.CreatedAt,
		})
		return nil
	default:
		return fmt.Errorf("unknown message event type: %s", event.EventType)
	}
}

func (r *Relay) dispatchTagSync(ctx context.Context, event *domain.OutboxEvent) error {
	var payload domain.TagSyncPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal tag sync payload: %w", err)
	}

	if err := r.syncer.SyncAIQueue(ctx, payload.UserID, payload.Tags, payload.OldTags); err != nil {
		return fmt.Errorf("SyncAIQueue: %w", err)
	}
	return nil
}

func (r *Relay) dispatchRoomClaimed(ctx context.Context, event *domain.OutboxEvent) error {
	var payload domain.RoomClaimedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal room claimed payload: %w", err)
	}

	if err := r.binder.BindRoomToAI(ctx, payload.RoomID, payload.AiID); err != nil {
		return fmt.Errorf("BindRoomToAI: %w", err)
	}
	r.notifyKafka(ctx, kafkaadapter.NotificationEvent{
		ID:          event.ID,
		Type:        event.EventType,
		RecipientID: payload.HumanID,
		RoomID:      payload.RoomID,
		SenderID:    payload.AiID,
		OccurredAt:  event.CreatedAt,
	})
	return nil
}

func (r *Relay) notifyKafka(ctx context.Context, event kafkaadapter.NotificationEvent) {
	if err := r.kafka.PublishNotification(ctx, event); err != nil {
		logrus.WithError(err).WithField("event_id", event.ID).Warn("outbox: kafka notification failed (best-effort)")
	}
}
