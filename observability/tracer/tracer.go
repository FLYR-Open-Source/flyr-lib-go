package tracer

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FlyrInc/flyr-lib-go/config"
	internalSpan "github.com/FlyrInc/flyr-lib-go/internal/span"
	internalUtils "github.com/FlyrInc/flyr-lib-go/internal/utils"
)

const (
	// The depth of the caller in the stack trace
	callerDepth = 3
)

// Tracer is a wrapper around the OpenTelemetry Tracer
type Tracer struct {
	tracer oteltrace.Tracer
}

// defaultTracer is the default tracer used by the application.
//
// The default tracer is initialized by the tracer.StartDefaultTracer(...) function.
var defaultTracer *Tracer

func includeCallerAttributes(span oteltrace.Span) {
	caller := internalUtils.GetCallerName(callerDepth)
	codeFilePath := attribute.String(config.CODE_PATH, caller.FilePath)
	codeLineNumber := attribute.Int(config.CODE_LINE, caller.LineNumber)
	codeFunctionName := attribute.String(config.CODE_FUNC, caller.FunctionName)
	codeNamespace := attribute.String(config.CODE_NS, caller.Namespace)
	span.SetAttributes(codeFilePath, codeLineNumber, codeFunctionName, codeNamespace)
}

// StartSpan begins a new span for tracing with the specified name and kind.
//
// This method takes a context, a span name, and a span kind as arguments. It checks
// if the Tracer instance is not nil, then starts a new span using the Tracer's Start
// method. The caller's information is added to the span's attributes to provide
// context about where the span was created. The function returns the updated context
// and a Span object that wraps the created span.
//
// It returns the new context and the Span.
func (t *Tracer) StartSpan(ctx context.Context, name string, kind SpanKind) (context.Context, internalSpan.Span) {
	if t.tracer == nil {
		return ctx, internalSpan.Span{}
	}

	ctx, span := t.tracer.Start(ctx, name, oteltrace.WithSpanKind(kind))
	// Add the caller to the span attributes
	includeCallerAttributes(span)

	return ctx, internalSpan.Span{Span: span}
}
