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

package tracer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/tracer"
)

// if that test fails, it means the depth of the caller is different,
// therefore the caller information is not being retrieved correctly
func TestStartSpan(t *testing.T) {
	ctx := context.Background()

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	tracer.StarCustomTracer(tp.Tracer("test-tracer"))

	//nolint:errcheck
	defer tp.Shutdown(ctx)

	_, span := tracer.StartSpan(ctx, "test-span", oteltrace.SpanKindInternal)
	span.SetAttributes(
		attribute.String("key1", "value1"),
	)
	span.End()

	ended := sr.Ended()
	require.Len(t, ended, 1)
	spanRec := ended[0]

	got := attrsToMap(spanRec.Attributes())
	assert.Equal(t, "test-span", spanRec.Name())
	assert.Equal(t, oteltrace.SpanKindInternal, spanRec.SpanKind())
	assert.Equal(t, "value1", got["key1"])
}

func attrsToMap(attrs []attribute.KeyValue) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		m[string(kv.Key)] = fmt.Sprint(kv.Value.AsInterface())
	}
	return m
}
