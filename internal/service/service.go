package service

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
	UpdateProfileTags(
		ctx context.Context,
		userID uuid.UUID,
		tags []string,
	) (oldTags []string, err error)
	GetProfileTags(ctx context.Context, userID uuid.UUID) (tags []string, err error)
}

type TagCache interface {
	AreTagsValid(ctx context.Context, tags []string) bool
}

// MessageHandler is the callback type for consuming messages from the notifier.
type MessageHandler func(ctx context.Context, body []byte) error

type MessageNotifier interface {
	PublishNewQuestion(ctx context.Context, msg *domain.Message) error
	SyncAIQueue(ctx context.Context, userID uuid.UUID, tags, oldTags []string) error
	Listen(ctx context.Context, userID uuid.UUID, handler func(ctx context.Context, body []byte) error) error
}

type Service struct {
	messageRepo MessageRepository
	tagRepo     TagRepository
	tagCache    TagCache
	notifier    MessageNotifier
}

func NewService(
	messageRepo MessageRepository,
	tagRepo TagRepository,
	tagCache TagCache,
	notifier MessageNotifier,
) *Service {
	return &Service{
		messageRepo: messageRepo,
		tagRepo:     tagRepo,
		tagCache:    tagCache,
		notifier:    notifier,
	}
}
