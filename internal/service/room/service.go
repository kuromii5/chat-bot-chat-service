package room

//go:generate mockery

import (
	"context"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type RoomRepo interface {
	ClaimRoom(ctx context.Context, roomID uuid.UUID, aiID uuid.UUID) error
	CloseRoom(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error
	GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error)
}

type Service struct {
	repo RoomRepo
}

func NewService(repo RoomRepo) *Service {
	return &Service{repo: repo}
}
