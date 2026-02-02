package service

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func (s *Service) UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error {
	if !s.tagRepo.AreTagsValid(ctx, tags) {
		return domain.ErrInvalidTags
	}
	slices.Sort(tags)
	tags = slices.Compact(tags)

	if err := s.tagRepo.UpdateProfileTags(ctx, userID, tags); err != nil {
		return err
	}

	if err := s.notifier.SetupAIQueue(ctx, userID, tags); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.tagRepo.GetProfileTags(ctx, userID)
}
