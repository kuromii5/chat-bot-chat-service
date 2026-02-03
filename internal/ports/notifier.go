package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type MessageNotifier interface {
	PublishNewQuestion(ctx context.Context, msg *domain.Message) error
	SyncAIQueue(ctx context.Context, userID uuid.UUID, tags, oldTags []string) error
}
