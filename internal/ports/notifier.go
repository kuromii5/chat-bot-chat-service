package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type MessageHandler func(ctx context.Context, body []byte) error
type MessageNotifier interface {
	PublishNewQuestion(ctx context.Context, msg *domain.Message) error
	SyncAIQueue(ctx context.Context, userID uuid.UUID, tags, oldTags []string) error
	Listen(ctx context.Context, userID uuid.UUID, handler MessageHandler) error
}
