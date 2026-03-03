package tracing

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

// broker is the union of all broker interfaces consumed by the outbox relay.
type broker interface {
	PublishNewQuestion(ctx context.Context, msg *domain.Message) error
	PublishFollowUp(ctx context.Context, roomID uuid.UUID, msg *domain.Message) error
	PublishAIReply(ctx context.Context, humanID uuid.UUID, msg *domain.Message) error
	SyncAIQueue(ctx context.Context, aiID uuid.UUID, tags, oldTags []string) error
	BindRoomToAI(ctx context.Context, roomID uuid.UUID, aiID uuid.UUID) error
}

const brokerTracer = "rabbitmq"

// Broker wraps any broker and adds an OTel span to every call.
type Broker struct {
	inner broker
}

func NewBroker(inner broker) *Broker {
	return &Broker{inner: inner}
}

func (b *Broker) PublishNewQuestion(ctx context.Context, msg *domain.Message) error {
	ctx, span := otel.Tracer(brokerTracer).Start(ctx, "rabbitmq.PublishNewQuestion")
	defer span.End()
	span.SetAttributes(
		attribute.String("messaging.operation", "publish"),
		attribute.String("messaging.routing_key", "question.*"),
		attribute.String("message.id", msg.ID.String()),
	)

	err := b.inner.PublishNewQuestion(ctx, msg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (b *Broker) PublishFollowUp(ctx context.Context, roomID uuid.UUID, msg *domain.Message) error {
	ctx, span := otel.Tracer(brokerTracer).Start(ctx, "rabbitmq.PublishFollowUp")
	defer span.End()
	span.SetAttributes(
		attribute.String("messaging.operation", "publish"),
		attribute.String("messaging.routing_key", "room."+roomID.String()),
		attribute.String("room.id", roomID.String()),
		attribute.String("message.id", msg.ID.String()),
	)

	err := b.inner.PublishFollowUp(ctx, roomID, msg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (b *Broker) PublishAIReply(ctx context.Context, humanID uuid.UUID, msg *domain.Message) error {
	ctx, span := otel.Tracer(brokerTracer).Start(ctx, "rabbitmq.PublishAIReply")
	defer span.End()
	span.SetAttributes(
		attribute.String("messaging.operation", "publish"),
		attribute.String("messaging.routing_key", "reply."+humanID.String()),
		attribute.String("human.id", humanID.String()),
		attribute.String("message.id", msg.ID.String()),
	)

	err := b.inner.PublishAIReply(ctx, humanID, msg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (b *Broker) SyncAIQueue(ctx context.Context, aiID uuid.UUID, tags, oldTags []string) error {
	ctx, span := otel.Tracer(brokerTracer).Start(ctx, "rabbitmq.SyncAIQueue")
	defer span.End()
	span.SetAttributes(
		attribute.String("messaging.operation", "queue_sync"),
		attribute.String("ai.id", aiID.String()),
		attribute.Int("tags.new", len(tags)),
		attribute.Int("tags.old", len(oldTags)),
	)

	err := b.inner.SyncAIQueue(ctx, aiID, tags, oldTags)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (b *Broker) BindRoomToAI(ctx context.Context, roomID uuid.UUID, aiID uuid.UUID) error {
	ctx, span := otel.Tracer(brokerTracer).Start(ctx, "rabbitmq.BindRoomToAI")
	defer span.End()
	span.SetAttributes(
		attribute.String("messaging.operation", "queue_bind"),
		attribute.String("room.id", roomID.String()),
		attribute.String("ai.id", aiID.String()),
	)

	err := b.inner.BindRoomToAI(ctx, roomID, aiID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
