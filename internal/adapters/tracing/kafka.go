package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/kuromii5/chat-bot-chat-service/internal/adapters/kafka"
)

type kafkaProd interface {
	PublishNotification(ctx context.Context, event kafka.NotificationEvent) error
}

const kafkaTracer = "kafka/producer"

type Kafka struct {
	inner kafkaProd
}

func NewKafkaProd(inner kafkaProd) *Kafka {
	return &Kafka{inner: inner}
}

func (k *Kafka) PublishNotification(ctx context.Context, event kafka.NotificationEvent) (err error) {
	_, span := otel.Tracer(kafkaTracer).Start(ctx, "kafka.PublishNotification")
	defer span.End()
	span.SetAttributes(
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.operation", "publish"),
		attribute.String("messaging.destination", kafka.NotificationsTopic),
		attribute.String("event.id", event.ID.String()),
		attribute.String("event.type", string(event.Type)),
		attribute.String("recipient.id", event.RecipientID.String()),
	)

	if err = k.inner.PublishNotification(ctx, event); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
