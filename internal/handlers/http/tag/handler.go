package tag

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error
	GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}
