package testhelpers

import (
	"context"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	sdktracetest "go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func GetMockSpan(ctx context.Context) (context.Context, oteltrace.Span) {
	tc := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(sdktracetest.NewInMemoryExporter()),
	)
	otel.SetTracerProvider(tc)
	defer func() {
		//nolint:errcheck
		tc.Shutdown(context.Background())
	}()
	spanCtx, span := tc.Tracer("test-tracer").Start(ctx, "test-span")

	return spanCtx, span
}
