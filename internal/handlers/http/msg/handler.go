package msg

import (
	"context"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	msgservice "github.com/kuromii5/chat-bot-chat-service/internal/service/msg"
)

type Service interface {
	SendMessage(ctx context.Context, req msgservice.CreateMessageReq) (*domain.Message, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}
