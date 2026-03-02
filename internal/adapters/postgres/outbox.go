package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func (pg *postgres) FetchPending(ctx context.Context, limit int) (events []*domain.OutboxEvent, err error) {
	if err = pg.DB.SelectContext(ctx, &events, fetchPendingQuery, limit); err != nil {
		return nil, fmt.Errorf("fetch pending: %w", err)
	}
	return events, nil
}

func (pg *postgres) MarkPublished(ctx context.Context, id uuid.UUID) error {
	if _, err := pg.DB.ExecContext(ctx, markPublishedQuery, id); err != nil {
		return fmt.Errorf("mark published: %w", err)
	}
	return nil
}

func (pg *postgres) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	if _, err := pg.DB.ExecContext(ctx, markFailedQuery, id, errMsg); err != nil {
		return fmt.Errorf("mark failed: %w", err)
	}
	return nil
}
