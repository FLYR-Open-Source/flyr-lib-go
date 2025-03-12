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

package span

import (
	"context"
	"testing"

	testhelpers "github.com/FLYR-Open-Source/flyr-lib-go/pkg/testhelpers/monitoring"
	"github.com/stretchr/testify/assert"
)

func TestGetSpanFromContext(t *testing.T) {
	t.Run("Span exists in the context", func(t *testing.T) {
		ctx, _ := testhelpers.GetFakeSpan(context.Background())
		span := GetSpanFromContext(ctx)
		defer span.End()
		assert.NotNil(t, span)
		assert.True(t, span.IsRecording())
		assert.True(t, span.SpanContext().HasSpanID())
		assert.True(t, span.SpanContext().HasTraceID())
	})

	t.Run("Span does not exists in the context", func(t *testing.T) {
		ctx := context.Background()
		span := GetSpanFromContext(ctx)
		assert.False(t, span.IsRecording())
		assert.False(t, span.SpanContext().HasSpanID())
		assert.False(t, span.SpanContext().HasTraceID())
	})
}

func TestExtractTrace(t *testing.T) {
	t.Run("Span exists in the context with valid trace and span IDs", func(t *testing.T) {
		ctx, span := testhelpers.GetFakeSpan(context.Background())
		traceID, spanID, found := ExtractTrace(ctx)
		defer span.End()

		assert.True(t, found)
		assert.NotEmpty(t, traceID)
		assert.NotEmpty(t, spanID)
	})

	t.Run("Span does not exists in the context", func(t *testing.T) {
		ctx := context.Background()
		traceID, spanID, found := ExtractTrace(ctx)
		assert.False(t, found)
		assert.Empty(t, traceID)
		assert.Empty(t, spanID)
	})

	t.Run("Span exists but it is not recording", func(t *testing.T) {
		ctx, span := testhelpers.GetFakeSpan(context.Background())
		span.End()
		traceID, spanID, found := ExtractTrace(ctx)

		assert.False(t, found)
		assert.Empty(t, traceID)
		assert.Empty(t, spanID)
	})
}
