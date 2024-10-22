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

// Extract trace from HTTP headers
func StartSpanFromHTTPRequest(ctx context.Context, tracer trace.Tracer, spanName string, req *http.Request) (context.Context, trace.Span) {
	// Use the request headers as the carrier to extract trace information
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, propagation.HeaderCarrier(req.Header))

	// Start a new span (continuing the trace if exists)
	ctx, span := tracer.Start(ctx, spanName)

	return ctx, span
}

// End the current span
func EndSpan(span trace.Span) {
	span.End()
}
