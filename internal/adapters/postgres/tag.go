package postgres

import (
	"context"

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

func (pg *Postgres) UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) (oldTags []string, err error) {
	tx, err := pg.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := tx.SelectContext(ctx, &oldTags, deleteProfileTagsQuery, userID); err != nil {
		return nil, err
	}

	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, insertProfileTagsQuery, userID, tag); err != nil {
			return nil, err
		}
	}

	return oldTags, tx.Commit()
}

func (pg *Postgres) GetProfileTags(ctx context.Context, userID uuid.UUID) (tags []string, err error) {
	return tags, pg.DB.SelectContext(ctx, &tags, getProfileTagsQuery, userID)
}
