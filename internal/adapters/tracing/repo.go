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
	Save(ctx context.Context, msg *domain.Message) (*domain.Message, error)
	GetLastMessages(ctx context.Context, roomID string, limit int) ([]*domain.Message, error)
	UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) (oldTags []string, err error)
	GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error)
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

func (r *Repo) Save(ctx context.Context, msg *domain.Message) (*domain.Message, error) {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.Save")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "core.messages"),
		attribute.String("user.id", msg.SenderID.String()),
	)

	result, err := r.inner.Save(ctx, msg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (r *Repo) GetLastMessages(ctx context.Context, roomID string, limit int) ([]*domain.Message, error) {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.GetLastMessages")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "core.messages"),
		attribute.String("room.id", roomID),
		attribute.Int("limit", limit),
	)

	result, err := r.inner.GetLastMessages(ctx, roomID, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (r *Repo) UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) ([]string, error) {
	ctx, span := otel.Tracer(dbTracer).Start(ctx, "postgres.UpdateProfileTags")
	defer span.End()
	span.SetAttributes(
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.table", "core.profile_tags"),
		attribute.String("user.id", userID.String()),
	)

	result, err := r.inner.UpdateProfileTags(ctx, userID, tags)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
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
