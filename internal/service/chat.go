package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/adapters/postgres/message"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/pkg/validator"
)

type CreateMessageReq struct {
	UserID  uuid.UUID
	Content string
}

func (s *Service) SendMessage(ctx context.Context, req CreateMessageReq) (*domain.Message, error) {
	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	role, err := s.messageRepo.GetUserRole(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	lastMsgs, err := s.messageRepo.GetLastMessages(ctx, "global", 3)
	if err != nil {
		return nil, err
	}

	switch role {
	case "Human":
		if err := message.ValidateHumanMsg(lastMsgs); err != nil {
			return nil, err
		}
	case "AI":
		if err := message.ValidateAIMsg(lastMsgs); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown role access denied")
	}

	msg := &domain.Message{
		ID:         uuid.New(),
		SenderID:   req.UserID,
		SenderRole: role,
		RoomID:     "global",
		Content:    req.Content,
		CreatedAt:  time.Now(),
	}

	if err := s.messageRepo.Save(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}
