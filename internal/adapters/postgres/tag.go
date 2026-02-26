package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (pg *postgres) UpdateProfileTags(
	ctx context.Context,
	userID uuid.UUID,
	tags []string,
) (oldTags []string, err error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.UpdateProfileTags")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.table", "core.profile_tags"),
		attribute.String("user.id", userID.String()),
	)

	tx, err := pg.DB.BeginTxx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // no-op if tx already committed
	}()

	if err := tx.SelectContext(ctx, &oldTags, deleteProfileTagsQuery, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("delete profile tags: %w", err)
	}

	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, insertProfileTagsQuery, userID, tag); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("insert profile tag: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("commit: %w", err)
	}
	return oldTags, nil
}

func (pg *postgres) GetProfileTags(
	ctx context.Context,
	userID uuid.UUID,
) ([]string, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.GetProfileTags")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "core.profile_tags"),
		attribute.String("user.id", userID.String()),
	)

	var tags []string
	if err := pg.DB.SelectContext(ctx, &tags, getProfileTagsQuery, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("get profile tags: %w", err)
	}
	return tags, nil
}
