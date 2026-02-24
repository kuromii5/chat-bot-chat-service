package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (pg *Postgres) AreTagsValid(ctx context.Context, tags []string) bool {
	pg.tagMu.RLock()
	defer pg.tagMu.RUnlock()

	for _, tag := range tags {
		if _, ok := pg.tagCache[tag]; !ok {
			return false
		}
	}

	return true
}

func (pg *Postgres) UpdateProfileTags(
	ctx context.Context,
	userID uuid.UUID,
	tags []string,
) (oldTags []string, err error) {
	tx, err := pg.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // no-op if tx already committed
	}()

	if err := tx.SelectContext(ctx, &oldTags, deleteProfileTagsQuery, userID); err != nil {
		return nil, fmt.Errorf("delete profile tags: %w", err)
	}

	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, insertProfileTagsQuery, userID, tag); err != nil {
			return nil, fmt.Errorf("insert profile tag: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return oldTags, nil
}

func (pg *Postgres) GetProfileTags(
	ctx context.Context,
	userID uuid.UUID,
) ([]string, error) {
	var tags []string
	if err := pg.DB.SelectContext(ctx, &tags, getProfileTagsQuery, userID); err != nil {
		return nil, fmt.Errorf("get profile tags: %w", err)
	}
	return tags, nil
}
