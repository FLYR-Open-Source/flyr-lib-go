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

package tracer // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/tracer"

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/span"
)

type SpanKind = oteltrace.SpanKind
type Code = codes.Code
type KeyValue = attribute.KeyValue

// StartSpan starts a new span with the given name and kind
//
// It is using the default tracer. To start a new default tracer, first the function
// tracer.StartDefaultTracer(...) should be called.
//
// It returns the new context and the new span
func StartSpan(ctx context.Context, name string, kind SpanKind) (context.Context, span.Span) {
	return defaultTracer.StartSpan(ctx, name, kind)
}

// GetSpanFromContext retrieves the current Span from the context.
//
// This allows access to tracing information within that context.
// If no span is associated with the context, it returns a non-recording span,
// which is a placeholder that performs no operations. This is useful for
// tracing operations where span context is required.
//
// Returns the current Span.
func GetSpanFromContext(ctx context.Context) span.Span {
	return span.GetSpanFromContext(ctx)
}
