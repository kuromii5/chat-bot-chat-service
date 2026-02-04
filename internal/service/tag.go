package service

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func (s *Service) UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error {
	slices.Sort(tags)
	tags = slices.Compact(tags)
	if !s.tagRepo.AreTagsValid(ctx, tags) {
		return domain.ErrInvalidTags
	}

	oldTags, err := s.tagRepo.UpdateProfileTags(ctx, userID, tags)
	if err != nil {
		return err
	}

	if err := s.notifier.SyncAIQueue(ctx, userID, tags, oldTags); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.tagRepo.GetProfileTags(ctx, userID)
}
