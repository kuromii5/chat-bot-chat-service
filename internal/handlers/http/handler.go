package http

import (
	"context"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	msg "github.com/kuromii5/chat-bot-chat-service/internal/service/msg"
)

type MessageSvc interface {
	SendMessage(ctx context.Context, req msg.CreateMessageReq) (*domain.Message, error)
}

type MessageHandler struct {
	svc MessageSvc
}

func NewMessageHandler(svc MessageSvc) *MessageHandler {
	return &MessageHandler{svc: svc}
}

type TagSvc interface {
	UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error
	GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type TagHandler struct {
	svc TagSvc
}

func NewTagHandler(svc TagSvc) *TagHandler {
	return &TagHandler{svc: svc}
}
