package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type MessageRepository interface {
	Save(ctx context.Context, msg *domain.Message) (*domain.Message, error)
	GetLastMessages(ctx context.Context, roomID string, limit int) ([]*domain.Message, error)
}

type TagRepository interface {
	AreTagsValid(ctx context.Context, tags []string) bool
	UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) (oldTags []string, err error)
	GetProfileTags(ctx context.Context, userID uuid.UUID) (tags []string, err error)
}
