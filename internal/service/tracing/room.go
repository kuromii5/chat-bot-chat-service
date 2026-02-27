package tracing

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type roomInner interface {
	ClaimRoom(ctx context.Context, roomID uuid.UUID, aiID uuid.UUID) error
	CloseRoom(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error
}

// RoomService wraps the room service and adds OTel spans.
// It satisfies handlers/http/room.Service via duck typing.
type RoomService struct {
	inner roomInner
}

func NewRoomService(inner roomInner) *RoomService {
	return &RoomService{inner: inner}
}

func (s *RoomService) ClaimRoom(ctx context.Context, roomID uuid.UUID, aiID uuid.UUID) error {
	ctx, span := otel.Tracer("service/room").Start(ctx, "room.ClaimRoom")
	defer span.End()
	span.SetAttributes(
		attribute.String("room.id", roomID.String()),
		attribute.String("ai.id", aiID.String()),
	)
	err := s.inner.ClaimRoom(ctx, roomID, aiID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (s *RoomService) CloseRoom(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error {
	ctx, span := otel.Tracer("service/room").Start(ctx, "room.CloseRoom")
	defer span.End()
	span.SetAttributes(
		attribute.String("room.id", roomID.String()),
		attribute.String("user.id", userID.String()),
	)
	err := s.inner.CloseRoom(ctx, roomID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
