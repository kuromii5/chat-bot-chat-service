package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

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

const fetchLimit = 100

type Relay struct {
	repo      OutboxRepo
	publisher Publisher
	interval  time.Duration
}

func NewRelay(repo OutboxRepo, publisher Publisher, interval time.Duration) *Relay {
	return &Relay{repo: repo, publisher: publisher, interval: interval}
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
	var payload domain.OutboxPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	switch event.EventType {
	case domain.EventNewQuestion:
		return r.publisher.PublishNewQuestion(ctx, payload.Message)
	case domain.EventFollowUp:
		return r.publisher.PublishFollowUp(ctx, payload.Message.RoomID, payload.Message)
	case domain.EventAIReply:
		return r.publisher.PublishAIReply(ctx, payload.HumanID, payload.Message)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}
