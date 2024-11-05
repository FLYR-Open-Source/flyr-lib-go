package tracer

import (
	"context"

	"github.com/FlyrInc/flyr-lib-go/internal/utils"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	// The depth of the caller in the stack trace
	callerDepth = 3
)

// Tracer is a wrapper around the OpenTelemetry Tracer
type Tracer struct {
	oteltrace.Tracer
}

// defaultTracer is the default tracer used by the application.
//
// The default tracer is initialized by the tracer.StartDefaultTracer(...) function.
var defaultTracer *Tracer

// StartSpan begins a new span for tracing with the specified name and kind.
//
// This method takes a context, a span name, and a span kind as arguments. It checks
// if the Tracer instance is not nil, then starts a new span using the Tracer's Start
// method. The caller's information is added to the span's attributes to provide
// context about where the span was created. The function returns the updated context
// and a Span object that wraps the created span.
//
// It returns the new context and the Span.
func (t *Tracer) StartSpan(ctx context.Context, name string, kind SpanKind) (context.Context, Span) {
	if t == nil {
		return ctx, Span{}
	}

	ctx, span := t.Start(ctx, name, oteltrace.WithSpanKind(kind))
	// Add the caller to the span attributes
	span.SetAttributes(attribute.String("caller", utils.GetCallerName(callerDepth).String()))

	return ctx, Span{span}
}
