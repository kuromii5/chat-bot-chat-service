package msg

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

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
	ctx, span := otel.Tracer("service/msg").Start(ctx, "msg.SendMessage")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.id", req.UserID.String()),
		attribute.String("user.role", string(req.Role)),
		attribute.StringSlice("message.tags", req.Tags),
	)

	if err := validator.Validate(req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("validate: %w", err)
	}

	lastMsgs, err := s.repo.GetLastMessages(ctx, "global", domain.HumanSequentialMessageLimit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("get last messages: %w", err)
	}

	switch req.Role {
	case domain.Human:
		if err := domain.ValidateHumanMsg(lastMsgs); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("validate human msg: %w", err)
		}
	case domain.AI:
		if err := domain.ValidateAIMsg(lastMsgs); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("validate AI msg: %w", err)
		}
	default:
		err := errors.New("unknown role: access denied")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("save message: %w", err)
	}

	if req.Role == domain.Human {
		go func() {
			publishCtx := context.WithoutCancel(ctx)
			if err := s.notifier.PublishNewQuestion(publishCtx, saved); err != nil {
				logrus.WithError(err).Error("failed to publish new question")
			}
		}()
	}

	return saved, nil
}
