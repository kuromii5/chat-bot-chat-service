package service

import (
	"context"
	"errors"
	"slices"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/pkg/validator"
	"github.com/sirupsen/logrus"
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

	lastMsgs, err := s.messageRepo.GetLastMessages(ctx, "global", domain.HumanSequentialMessageLimit)
	if err != nil {
		return nil, err
	}

	switch req.Role {
	case domain.Human:
		if err := domain.ValidateHumanMsg(lastMsgs); err != nil {
			return nil, err
		}
	case domain.AI:
		if err := domain.ValidateAIMsg(lastMsgs); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown role: access denied")
	}

	slices.Sort(req.Tags)
	req.Tags = slices.Compact(req.Tags)
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

	if req.Role == domain.Human {
		go func() {
			if err := s.notifier.PublishNewQuestion(ctx, msg); err != nil {
				logrus.WithError(err).Error("failed to publish new question")
			}
		}()
	}

	return msg, nil
}
