package tag

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func (s *Service) UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error {
	ctx, span := otel.Tracer("service/tag").Start(ctx, "tag.UpdateProfileTags")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.StringSlice("tags", tags),
	)

	slices.Sort(tags)
	tags = slices.Compact(tags)

	if !s.cache.AreTagsValid(ctx, tags) {
		span.RecordError(domain.ErrInvalidTags)
		span.SetStatus(codes.Error, domain.ErrInvalidTags.Error())
		return domain.ErrInvalidTags
	}

	oldTags, err := s.repo.UpdateProfileTags(ctx, userID, tags)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("update profile tags: %w", err)
	}

	if err := s.notifier.SyncAIQueue(ctx, userID, tags, oldTags); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("sync AI queue: %w", err)
	}

	return nil
}

func (s *Service) GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error) {
	ctx, span := otel.Tracer("service/tag").Start(ctx, "tag.GetProfileTags")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", userID.String()))

	tags, err := s.repo.GetProfileTags(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("get profile tags: %w", err)
	}
	return tags, nil
}
