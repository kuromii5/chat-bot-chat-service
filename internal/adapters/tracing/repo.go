package tracing

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

// postgresRepo is the union of all repo interfaces defined in the service layer.
// Repo satisfies them all via duck typing — no service package imports needed.
type postgresRepo interface {
	SaveWithOutbox(ctx context.Context, msg *domain.Message, eventType domain.EventType, humanID uuid.UUID) (*domain.Message, error)
	GetLastMessages(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.Message, error)
	UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error
	GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error)
	CreateRoom(ctx context.Context, humanID uuid.UUID) (*domain.Room, error)
	GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error)
	ClaimRoom(ctx context.Context, roomID uuid.UUID, aiID uuid.UUID) error
	CloseRoom(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error
	FetchPending(ctx context.Context, limit int) ([]*domain.OutboxEvent, error)
	MarkPublished(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}

const dbTracer = "postgres"

// Repo wraps any postgresRepo and adds an OTel span to every DB call.
// The postgres adapter stays OTel-free.
type Repo struct {
	inner postgresRepo
}

func NewRepo(inner postgresRepo) *Repo {
	return &Repo{inner: inner}
}

func (r *Repo) SaveWithOutbox(ctx context.Context, msg *domain.Message, eventType domain.EventType, humanID uuid.UUID) (*domain.Message, error) {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.SaveWithOutbox")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "core.messages, core.outbox_events"),
		attribute.String("user.id", msg.SenderID.String()),
		attribute.String("event.type", string(eventType)),
	)

	result, err := r.inner.SaveWithOutbox(ctx, msg, eventType, humanID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (r *Repo) GetLastMessages(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.Message, error) {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.GetLastMessages")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "core.messages"),
		attribute.String("room.id", roomID.String()),
		attribute.Int("limit", limit),
	)

	result, err := r.inner.GetLastMessages(ctx, roomID, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (r *Repo) UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.UpdateProfileTags")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.table", "core.profile_tags, core.outbox_events"),
		attribute.String("user.id", userID.String()),
	)

	err := r.inner.UpdateProfileTags(ctx, userID, tags)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (r *Repo) GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error) {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.GetProfileTags")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "core.profile_tags"),
		attribute.String("user.id", userID.String()),
	)

	result, err := r.inner.GetProfileTags(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (r *Repo) CreateRoom(ctx context.Context, humanID uuid.UUID) (*domain.Room, error) {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.CreateRoom")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "core.rooms"),
		attribute.String("human.id", humanID.String()),
	)

	result, err := r.inner.CreateRoom(ctx, humanID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (r *Repo) GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error) {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.GetRoom")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "core.rooms"),
		attribute.String("room.id", roomID.String()),
	)

	result, err := r.inner.GetRoom(ctx, roomID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (r *Repo) ClaimRoom(ctx context.Context, roomID uuid.UUID, aiID uuid.UUID) error {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.ClaimRoom")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.table", "core.rooms"),
		attribute.String("room.id", roomID.String()),
		attribute.String("ai.id", aiID.String()),
	)

	err := r.inner.ClaimRoom(ctx, roomID, aiID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (r *Repo) CloseRoom(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.CloseRoom")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.table", "core.rooms"),
		attribute.String("room.id", roomID.String()),
		attribute.String("user.id", userID.String()),
	)

	err := r.inner.CloseRoom(ctx, roomID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (r *Repo) FetchPending(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.FetchPending")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "SELECT FOR UPDATE"),
		attribute.String("db.table", "core.outbox_events"),
		attribute.Int("limit", limit),
	)

	result, err := r.inner.FetchPending(ctx, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (r *Repo) MarkPublished(ctx context.Context, id uuid.UUID) error {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.MarkPublished")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.table", "core.outbox_events"),
		attribute.String("event.id", id.String()),
	)

	err := r.inner.MarkPublished(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (r *Repo) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.MarkFailed")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.table", "core.outbox_events"),
		attribute.String("event.id", id.String()),
	)

	err := r.inner.MarkFailed(ctx, id, errMsg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
