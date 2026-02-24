package msg

import (
	"context"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type MessageRepo interface {
	Save(ctx context.Context, msg *domain.Message) (*domain.Message, error)
	GetLastMessages(ctx context.Context, roomID string, limit int) ([]*domain.Message, error)
}

type Notifier interface {
	PublishNewQuestion(ctx context.Context, msg *domain.Message) error
}

type Service struct {
	repo     MessageRepo
	notifier Notifier
}

func NewService(repo MessageRepo, notifier Notifier) *Service {
	return &Service{repo: repo, notifier: notifier}
}
