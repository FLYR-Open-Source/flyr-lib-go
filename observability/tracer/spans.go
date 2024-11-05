package tracer

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FlyrInc/flyr-lib-go/internal/span"
)

// Span is a wrapper around oteltrace.Span
//
// It is used to provide a more convenient API with extra functionality
type Span struct {
	oteltrace.Span
}

type SpanKind = oteltrace.SpanKind
type Code = codes.Code
type KeyValue = attribute.KeyValue

// StartSpan starts a new span with the given name and kind
//
// It is using the default tracer. To start a new default tracer, first the function
// tracer.StartDefaultTracer(...) should be called.
//
// It returns the new context and the new span
func StartSpan(ctx context.Context, name string, kind SpanKind) (context.Context, Span) {
	if defaultTracer == nil {
		return ctx, Span{}
	}

	return defaultTracer.StartSpan(ctx, name, kind)
}

// EndWithError ends the span by updating the status to Error and recording the error
func (s Span) EndWithError(err error) {
	if err != nil {
		s.SetStatus(codes.Error, err.Error())
		s.RecordError(err)
	}
	s.End()
}

// EndSuccessfully ends the span by updating the status to Ok
func (s Span) EndSuccessfully() {
	s.SetStatus(codes.Ok, "")
	s.End()
}

// GetSpanFromContext retrieves the current Span from the context.
//
// This allows access to tracing information within that context.
// If no span is associated with the context, it returns a non-recording span,
// which is a placeholder that performs no operations. This is useful for
// tracing operations where span context is required.
//
// Returns the current Span.
func GetSpanFromContext(ctx context.Context) Span {
	return span.GetSpanFromContext(ctx).(Span)
}
