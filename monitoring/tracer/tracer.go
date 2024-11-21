package tracer // import "github.com/FlyrInc/flyr-lib-go/tracer"

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FlyrInc/flyr-lib-go/config"
	internalSpan "github.com/FlyrInc/flyr-lib-go/internal/span"
	internaltracer "github.com/FlyrInc/flyr-lib-go/internal/traceprovider"
	internalUtils "github.com/FlyrInc/flyr-lib-go/internal/utils"
)

var ErrTracerNameNotSet = errors.New("tracer name not set")

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

// StartDefaultTracer initializes and starts the default OpenTelemetry Tracer.
//
// This function checks if tracing is enabled in the provided configuration. If tracing
// is enabled, It creates a new Tracer by using the default Trace Provider. It also validates that the
// tracer name (the service name in that case) is set in the configuration. If the tracer name is not provided,
// it returns an error indicating that the service name must be set.
//
// The function also sets the global default tracer to be used for tracing in the
// application. If tracing is not enabled, it returns nil without starting a tracer.
//
// It returns an error if any occurred.
func StartDefaultTracer(ctx context.Context, cfg config.MonitoringConfig) error {
	if !cfg.TracerEnabled() {
		return nil
	}

	if cfg.Service() == "" {
		return ErrTracerNameNotSet
	}

	err := internaltracer.InitializeTracerProvider(ctx, cfg)
	if err != nil {
		return err
	}

	tc := otel.GetTracerProvider()
	tracer := &Tracer{
		tracer: tc.Tracer(
			cfg.Service(),
			oteltrace.WithInstrumentationVersion("v0.0.1"), // TODO: Update instrumentation version
		),
	}

	defaultTracer = tracer
	return nil
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
func (t *Tracer) StartSpan(parentCtx context.Context, name string, kind SpanKind) (context.Context, internalSpan.Span) {
	if t.tracer == nil {
		return parentCtx, internalSpan.Span{}
	}

	ctxWithSpan, span := t.tracer.Start(parentCtx, name, oteltrace.WithSpanKind(kind))

	// Add the caller to the span attributes
	caller := internalUtils.GetCallerName(callerDepth)
	attrs := caller.SpanAttributes()
	span.SetAttributes(attrs...)

	return ctxWithSpan, internalSpan.Span{Span: span}
}
