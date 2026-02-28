package msg

//go:generate mockery

import (
	"context"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type MessageRepo interface {
	Save(ctx context.Context, msg *domain.Message) (*domain.Message, error)
	GetLastMessages(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.Message, error)
}

type RoomRepo interface {
	CreateRoom(ctx context.Context, humanID uuid.UUID) (*domain.Room, error)
	GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error)
}

type Notifier interface {
	PublishNewQuestion(ctx context.Context, msg *domain.Message) error
	PublishAIReply(ctx context.Context, humanID uuid.UUID, msg *domain.Message) error
	PublishFollowUp(ctx context.Context, roomID uuid.UUID, msg *domain.Message) error
}

type Service struct {
	repo     MessageRepo
	roomRepo RoomRepo
	notifier Notifier
}

func NewService(repo MessageRepo, roomRepo RoomRepo, notifier Notifier) *Service {
	return &Service{repo: repo, roomRepo: roomRepo, notifier: notifier}
}
