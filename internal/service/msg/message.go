package msg

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

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
		return nil, fmt.Errorf("validate: %w", err)
	}

	lastMsgs, err := s.repo.GetLastMessages(ctx, "global", domain.HumanSequentialMessageLimit)
	if err != nil {
		return nil, fmt.Errorf("get last messages: %w", err)
	}

	switch req.Role {
	case domain.Human:
		if err := domain.ValidateHumanMsg(lastMsgs); err != nil {
			return nil, fmt.Errorf("validate human msg: %w", err)
		}
	case domain.AI:
		if err := domain.ValidateAIMsg(lastMsgs); err != nil {
			return nil, fmt.Errorf("validate AI msg: %w", err)
		}
	default:
		return nil, errors.New("unknown role: access denied")
	}

	slices.Sort(req.Tags)
	req.Tags = slices.Compact(req.Tags)
	saved, err := s.repo.Save(ctx, &domain.Message{
		SenderID:   req.UserID,
		SenderRole: req.Role,
		RoomID:     "global",
		Content:    req.Content,
		Tags:       req.Tags,
	})
	if err != nil {
		return nil, fmt.Errorf("save message: %w", err)
	}

	if req.Role == domain.Human {
		go func() {
			if err := s.notifier.PublishNewQuestion(ctx, saved); err != nil {
				logrus.WithError(err).Error("failed to publish new question")
			}
		}()
	}

	return saved, nil
}
