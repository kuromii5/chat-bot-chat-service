package tracing

import (
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type OTelHook struct{}

func (h *OTelHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *OTelHook) Fire(entry *logrus.Entry) error {
	ctx := entry.Context
	if ctx == nil {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		entry.Data["trace_id"] = span.SpanContext().TraceID().String()
	}
	return nil
}
