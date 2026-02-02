package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/internal/service"
)

type ChatService interface {
	SendMessage(ctx context.Context, req service.CreateMessageReq) (*domain.Message, error)
	UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error
	GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type Handler struct {
	service ChatService
}

func NewHandler(service ChatService) *Handler {
	return &Handler{
		service: service,
	}
}
