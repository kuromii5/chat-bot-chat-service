package tracing

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type tagInner interface {
	UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error
	GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// TagService wraps the tag service and adds OTel spans around each method.
// It satisfies handlers/http/tag.Service via duck typing.
type TagService struct {
	inner tagInner
}

func NewTagService(inner tagInner) *TagService {
	return &TagService{inner: inner}
}

func (s *TagService) UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error {
	ctx, span := otel.Tracer("service/tag").Start(ctx, "tag.UpdateProfileTags")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.StringSlice("tags", tags),
	)

	err := s.inner.UpdateProfileTags(ctx, userID, tags)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (s *TagService) GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error) {
	ctx, span := otel.Tracer("service/tag").Start(ctx, "tag.GetProfileTags")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", userID.String()))

	result, err := s.inner.GetProfileTags(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}
