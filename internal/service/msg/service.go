package msg

//go:generate mockery

import (
	"context"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type MessageRepo interface {
	SaveWithOutbox(ctx context.Context, msg *domain.Message, eventType domain.EventType, humanID, aiID uuid.UUID) (*domain.Message, error)
	GetLastMessage(ctx context.Context, roomID uuid.UUID) (*domain.Message, error)
}

type RoomRepo interface {
	CreateRoom(ctx context.Context, humanID uuid.UUID) (*domain.Room, error)
	GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error)
}

type Service struct {
	repo     MessageRepo
	roomRepo RoomRepo
}

func NewService(repo MessageRepo, roomRepo RoomRepo) *Service {
	return &Service{repo: repo, roomRepo: roomRepo}
}
