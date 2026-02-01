package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type MessageRepository interface {
	Save(ctx context.Context, msg *domain.Message) error
	GetLastMessages(ctx context.Context, roomID string, limit int) ([]*domain.Message, error)
	GetUserRole(ctx context.Context, userID uuid.UUID) (string, error)
}
