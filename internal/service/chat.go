package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/adapters/postgres/message"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/pkg/validator"
)

type CreateMessageReq struct {
	UserID  uuid.UUID
	Content string
	Role    domain.Role
	Tags    []string
}

func (s *Service) SendMessage(ctx context.Context, req CreateMessageReq) (*domain.Message, error) {
	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	lastMsgs, err := s.messageRepo.GetLastMessages(ctx, "global", 3)
	if err != nil {
		return nil, err
	}

	switch req.Role {
	case domain.Human:
		if err := message.ValidateHumanMsg(lastMsgs); err != nil {
			return nil, err
		}
	case domain.AI:
		if err := message.ValidateAIMsg(lastMsgs); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown role: access denied")
	}

	msg, err := s.messageRepo.Save(ctx, &domain.Message{
		SenderID:   req.UserID,
		SenderRole: req.Role,
		RoomID:     "global",
		Content:    req.Content,
		Tags:       req.Tags,
	})
	if err != nil {
		return nil, err
	}

	// TODO: publish message to rabbitmq

	return msg, nil
}
