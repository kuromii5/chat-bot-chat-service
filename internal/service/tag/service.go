package tag

//go:generate mockery

import (
	"context"

	"github.com/google/uuid"
)

type TagRepo interface {
	UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error
	GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type TagCache interface {
	AreTagsValid(ctx context.Context, tags []string) bool
}

type Service struct {
	repo  TagRepo
	cache TagCache
}

func NewService(repo TagRepo, cache TagCache) *Service {
	return &Service{repo: repo, cache: cache}
}
