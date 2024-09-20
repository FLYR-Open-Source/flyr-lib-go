package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Initialize OpenTelemetry Tracer
func InitTracer(serviceName string) trace.Tracer {
	tp := otel.Tracer(serviceName)
	return tp
}

// Start a new trace span
func StartSpan(ctx context.Context, tracer trace.Tracer, spanName string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, spanName)
	return ctx, span
}

// End the current span
func EndSpan(span trace.Span) {
	span.End()
}
