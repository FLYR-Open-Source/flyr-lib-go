package testhelpers

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"
)

type FakeTracer struct {
	oteltrace.Tracer
}

func (t FakeTracer) Start(ctx context.Context, name string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	spanCtx, span := t.Tracer.Start(ctx, name)

	newCtx, newSpan := overrideContextValue(spanCtx, span)
	return newCtx, &newSpan
}
