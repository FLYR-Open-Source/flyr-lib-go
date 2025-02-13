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

package testhelpers // import "github.com/FlyrInc/flyr-lib-go/pkg/testhelpers"

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	sdktracetest "go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func GetFakeTracer() (*sdktrace.TracerProvider, FakeTracer) {
	tc := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(sdktracetest.NewInMemoryExporter()),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

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
