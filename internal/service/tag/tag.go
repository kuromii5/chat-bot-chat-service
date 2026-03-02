package tag

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func (s *Service) UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error {
	slices.Sort(tags)
	tags = slices.Compact(tags)

	if !s.cache.AreTagsValid(ctx, tags) {
		return domain.ErrInvalidTags
	}

	if err := s.repo.UpdateProfileTags(ctx, userID, tags); err != nil {
		return fmt.Errorf("update profile tags: %w", err)
	}

	return nil
}

func (s *Service) GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error) {
	tags, err := s.repo.GetProfileTags(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get profile tags: %w", err)
	}
	return tags, nil
}
