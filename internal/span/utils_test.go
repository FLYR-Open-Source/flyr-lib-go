package span

import (
	"context"
	"testing"

	"github.com/FlyrInc/flyr-lib-go/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestGetSpanFromContext(t *testing.T) {
	t.Run("Span exists in the context", func(t *testing.T) {
		ctx, _ := testhelpers.GetMockSpan(context.Background())
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
		ctx, span := testhelpers.GetMockSpan(context.Background())
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
		ctx, span := testhelpers.GetMockSpan(context.Background())
		span.End()
		traceID, spanID, found := ExtractTrace(ctx)

		assert.False(t, found)
		assert.Empty(t, traceID)
		assert.Empty(t, spanID)
	})
}
