package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func (pg *postgres) UpdateProfileTags(
	ctx context.Context,
	userID uuid.UUID,
	tags []string,
) error {
	tx, err := pg.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // no-op if tx already committed
	}()

	var oldTags []string
	if err := tx.SelectContext(ctx, &oldTags, deleteProfileTagsQuery, userID); err != nil {
		return fmt.Errorf("delete profile tags: %w", err)
	}

	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, insertProfileTagsQuery, userID, tag); err != nil {
			return fmt.Errorf("insert profile tag: %w", err)
		}
	}

	payload, err := json.Marshal(domain.TagSyncPayload{
		UserID:  userID,
		Tags:    tags,
		OldTags: oldTags,
	})
	if err != nil {
		return fmt.Errorf("marshal tag sync payload: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		saveOutboxEventQuery,
		"profile_tags",
		userID,
		domain.EventTagsSync,
		payload,
	); err != nil {
		return fmt.Errorf("save outbox event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (pg *postgres) GetProfileTags(
	ctx context.Context,
	userID uuid.UUID,
) (tags []string, err error) {
	if err = pg.DB.SelectContext(ctx, &tags, getProfileTagsQuery, userID); err != nil {
		return nil, fmt.Errorf("get profile tags: %w", err)
	}
	return tags, nil
}
