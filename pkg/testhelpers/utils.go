package testhelpers // import "github.com/FlyrInc/flyr-lib-go/pkg/testhelpers"

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"
)

type traceContextKeyType int

const currentSpanKey traceContextKeyType = iota

func overrideContextValue(ctx context.Context, span oteltrace.Span) (context.Context, FakeSpan) {
	fakeSpan := FakeSpan{Span: span}
	ctx = context.WithValue(ctx, currentSpanKey, fakeSpan)
	return ctx, fakeSpan
}
