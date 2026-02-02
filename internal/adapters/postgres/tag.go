package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (pg *Postgres) AreTagsValid(ctx context.Context, tags []string) bool {
	if len(tags) == 0 {
		return true
	}

	pg.tagMu.RLock()
	defer pg.tagMu.RUnlock()

	for _, tag := range tags {
		if _, ok := pg.tagCache[tag]; !ok {
			return false
		}
	}

	return true
}

func (pg *Postgres) UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error {
	tx, err := pg.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, deleteProfileTagsQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to clear old tags: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, insertProfileTagsQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, tag := range tags {
		if _, err := stmt.ExecContext(ctx, userID, tag); err != nil {
			return fmt.Errorf("failed to insert tag %s: %w", tag, err)
		}
	}

	return tx.Commit()
}

func (pg *Postgres) GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var tags []string
	return tags, pg.DB.SelectContext(ctx, &tags, getProfileTagsQuery, userID)
}
