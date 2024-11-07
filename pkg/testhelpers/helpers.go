package testhelpers

import (
	"context"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	sdktracetest "go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func GetFakeTracer() (*sdktrace.TracerProvider, FakeTracer) {
	tc := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(sdktracetest.NewInMemoryExporter()),
	)
	otel.SetTracerProvider(tc)

	return tc, FakeTracer{Tracer: tc.Tracer("test-tracer")}
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

	tc, tracer := GetFakeTracer()
	newCtx, newSpan := tracer.Start(ctx, "test-span")
	return newCtx, FakeSpan{Span: newSpan}
}
