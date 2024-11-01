package tracer

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"
)

func GetSpanFromContext(ctx context.Context) oteltrace.Span {
	return oteltrace.SpanFromContext(ctx)
}

func ExtractTrace(ctx context.Context) (traceID, spanID string, found bool) {
	span := GetSpanFromContext(ctx)
	if span == nil {
		return "", "", false
	}

	traceID = span.SpanContext().TraceID().String()
	spanID = span.SpanContext().SpanID().String()
	return traceID, spanID, true
}
