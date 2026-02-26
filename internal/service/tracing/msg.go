package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	msgservice "github.com/kuromii5/chat-bot-chat-service/internal/service/msg"
)

type msgInner interface {
	SendMessage(ctx context.Context, req msgservice.CreateMessageReq) (*domain.Message, error)
}

// MsgService wraps the msg service and adds an OTel span around SendMessage.
// It satisfies handlers/http/msg.Service via duck typing.
type MsgService struct {
	inner msgInner
}

func NewMsgService(inner msgInner) *MsgService {
	return &MsgService{inner: inner}
}

func (s *MsgService) SendMessage(ctx context.Context, req msgservice.CreateMessageReq) (*domain.Message, error) {
	ctx, span := otel.Tracer("service/msg").Start(ctx, "msg.SendMessage")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.id", req.UserID.String()),
		attribute.String("user.role", string(req.Role)),
		attribute.StringSlice("message.tags", req.Tags),
	)

	result, err := s.inner.SendMessage(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}
