package tag

import (
	"context"

	"github.com/google/uuid"
)

type TagRepo interface {
	UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) (oldTags []string, err error)
	GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type TagCache interface {
	AreTagsValid(ctx context.Context, tags []string) bool
}

type Notifier interface {
	SyncAIQueue(ctx context.Context, userID uuid.UUID, tags, oldTags []string) error
}

type Service struct {
	repo     TagRepo
	cache    TagCache
	notifier Notifier
}

func NewService(repo TagRepo, cache TagCache, notifier Notifier) *Service {
	return &Service{repo: repo, cache: cache, notifier: notifier}
}
