// MIT License
//
// Copyright (c) 2025 FLYR, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package tracer // import "github.com/FlyrInc/flyr-lib-go/monitoring/tracer"

import (
	"context"

	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/FlyrInc/flyr-lib-go/internal/config"
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

// StartDefaultTracer initializes and starts the default OpenTelemetry Tracer.
//
// This function checks if tracing is enabled in the provided configuration. If tracing
// is enabled, It creates a new Tracer by using the default TraceProvider. It also validates that the
// tracer name (the service name in that case) is set in the configuration. If the tracer name is not provided,
// a noop Tracer is initialised as a default.
//
// The function also sets the global default tracer to be used for tracing in the
// application. If tracing is not enabled, it returns nil without starting a tracer.
//
// It returns an error if any occurred.
func StartDefaultTracer(ctx context.Context) error {
	cfg := config.NewMonitoringConfig()

	err := initializeTracerProvider(ctx, cfg)
	if err != nil {
		return err
	}

	tc := otel.GetTracerProvider()

	var tracer Tracer
	if cfg.Service() == "" {
		tracer.tracer = noop.Tracer{}
	} else {
		tracer.tracer = tc.Tracer(
			cfg.Service(),
			oteltrace.WithInstrumentationVersion("v0.0.1"), // TODO: Update instrumentation version
		)
	}

	defaultTracer = &tracer
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
