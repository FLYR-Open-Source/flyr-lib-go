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

package middleware // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/middleware"

import (
	"context"
	"testing"

	internalConfig "github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const (
	testService = "test-service"
	// A valid W3C traceparent header and the IDs it carries.
	testTraceparent = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	testTraceID     = "4bf92f3577b34da6a3ce929d0e0e4736"
	testParentSpan  = "00f067aa0ba902b7"
)

// setupTracing installs an in-memory tracer provider and a W3C trace-context
// propagator as the OTel globals, and a known service name in the monitoring
// config. Everything is restored when the test finishes. It must be called
// before the middleware constructors, which capture the globals.
func setupTracing(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()

	t.Setenv("OTEL_SERVICE_NAME", testService)
	internalConfig.ResetMonitoringConfig()

	previousTP := otel.GetTracerProvider()
	previousProp := otel.GetTextMapPropagator()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	t.Cleanup(func() {
		otel.SetTracerProvider(previousTP)
		otel.SetTextMapPropagator(previousProp)
		_ = tp.Shutdown(context.Background())
		internalConfig.ResetMonitoringConfig()
	})

	return recorder
}

// attrMap indexes span attributes by key.
func attrMap(attrs []attribute.KeyValue) map[attribute.Key]attribute.Value {
	m := make(map[attribute.Key]attribute.Value, len(attrs))
	for _, kv := range attrs {
		m[kv.Key] = kv.Value
	}
	return m
}

// onlySpan asserts exactly one span was recorded and returns it.
func onlySpan(t *testing.T, recorder *tracetest.SpanRecorder) sdktrace.ReadOnlySpan {
	t.Helper()

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	return spans[0]
}
