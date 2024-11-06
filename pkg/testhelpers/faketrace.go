package testhelpers

import (
	"context"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	sdktracetest "go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type traceContextKeyType int

const currentSpanKey traceContextKeyType = iota

func override(ctx context.Context, span oteltrace.Span) (context.Context, FakeSpan) {
	fakeSpan := FakeSpan{Span: span}
	ctx = context.WithValue(ctx, currentSpanKey, fakeSpan)
	return ctx, fakeSpan
}

func GetFakeSpan(ctx context.Context) (context.Context, FakeSpan) {
	tc := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(sdktracetest.NewInMemoryExporter()),
	)
	otel.SetTracerProvider(tc)
	defer func() {
		//nolint:errcheck
		tc.Shutdown(context.Background())
	}()

	spanCtx, span := tc.Tracer("test-tracer").Start(ctx, "test-span")

	return override(spanCtx, span)
}
