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

package span // import "github.com/FLYR-Open-Source/flyr-lib-go/internal/span"

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"
)

// GetSpanFromContext retrieves the current OpenTelemetry span from the context.
//
// This allows access to tracing information within that context.
// If no span is associated with the context, it returns a non-recording span,
// which is a placeholder that performs no operations. This is useful for
// tracing operations where span context is required.
//
// Returns the custom Span type.
func GetSpanFromContext(ctx context.Context) Span {
	return Span{oteltrace.SpanFromContext(ctx)}
}

// ExtractTrace extracts the trace and span IDs from the current span in the context.
//
// This function retrieves the trace ID and span ID from the active OpenTelemetry
// span associated with the provided context. If a valid trace or span ID is not
// present, or if the span is not currently recording, it returns false, indicating
// that tracing information was not found. When found, it returns the trace ID,
// span ID, and a boolean indicating success.
//
// Returns:
//   - traceID: A string representation of the trace ID.
//   - spanID: A string representation of the span ID.
//   - found: A boolean indicating whether trace and span IDs were successfully extracted.
func ExtractTrace(ctx context.Context) (traceID, spanID string, found bool) {
	span := GetSpanFromContext(ctx)

	if !span.SpanContext().HasTraceID() || !span.SpanContext().HasSpanID() || !span.IsRecording() {
		return "", "", false
	}

	traceID = span.SpanContext().TraceID().String()
	spanID = span.SpanContext().SpanID().String()
	return traceID, spanID, true
}
