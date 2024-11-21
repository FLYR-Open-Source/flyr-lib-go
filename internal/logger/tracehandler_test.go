package logger

import (
	"context"
	"log/slog"
	"testing"

	"github.com/FlyrInc/flyr-lib-go/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSpanExtractor simulates extracting trace and span IDs from the context
type MockSpanExtractor struct {
	mock.Mock
}

func (m *MockSpanExtractor) ExtractTrace(ctx context.Context) (traceID, spanID string, found bool) {
	args := m.Called(ctx)
	return args.String(0), args.String(1), args.Bool(2)
}

type MockHandler struct {
	r *slog.Record
}

func (h *MockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *MockHandler) Handle(ctx context.Context, record slog.Record) error {
	h.r = &record
	return nil
}

func (h *MockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *MockHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *MockHandler) Handler() slog.Handler {
	return h
}

func TestTracingHandler_Enabled(t *testing.T) {
	handler := NewTracingHandler(slog.LevelInfo)(&MockHandler{})
	assert.False(t, handler.Enabled(context.Background(), slog.LevelDebug))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelWarn))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelError))
}

func TestTracingHandler_Handle_AddsTraceIDs(t *testing.T) {
	ctx := context.Background()
	mockHanlder := &MockHandler{}
	record := slog.Record{
		Level: slog.LevelInfo,
	}

	spanCtx, span := testhelpers.GetFakeSpan(ctx)
	defer span.End()

	handler := NewTracingHandler(slog.LevelInfo)(mockHanlder)
	err := handler.Handle(spanCtx, record)
	require.NoError(t, err)

	// Check that the trace and span IDs were added to the log record
	assert.Equal(t, 2, mockHanlder.r.NumAttrs())

	// Check that the trace and span IDs match the span's trace and span IDs
	mockHanlder.r.Attrs(func(a slog.Attr) bool {
		if a.Key == traceIDKey {
			assert.Equal(t, span.SpanContext().TraceID().String(), a.Value.String())
		}

		if a.Key == spanIDKey {
			assert.Equal(t, span.SpanContext().SpanID().String(), a.Value.String())
		}

		return true
	})
}

func TestTracingHandler_Handle_NoTraceIDs(t *testing.T) {
	t.Run("No trace IDs when ctx does not contain the span details", func(t *testing.T) {
		ctx := context.Background()
		mockHanlder := &MockHandler{}
		record := slog.Record{
			Level: slog.LevelInfo,
		}

		handler := NewTracingHandler(slog.LevelInfo)(mockHanlder)
		err := handler.Handle(ctx, record)
		require.NoError(t, err)

		// Check that the trace and span IDs were added to the log record
		assert.Equal(t, 0, mockHanlder.r.NumAttrs())
	})

	t.Run("No trace IDs when log level is debug", func(t *testing.T) {
		ctx := context.Background()
		mockHanlder := &MockHandler{}
		record := slog.Record{
			Level: slog.LevelDebug,
		}

		spanCtx, span := testhelpers.GetFakeSpan(ctx)
		defer span.End()

		handler := NewTracingHandler(slog.LevelDebug)(mockHanlder)
		err := handler.Handle(spanCtx, record)
		require.NoError(t, err)

		// Check that the trace and span IDs were added to the log record
		assert.Equal(t, 0, mockHanlder.r.NumAttrs())
	})
}

func TestTracingHandler_WithAttrs(t *testing.T) {
	mockNext := &MockHandler{}
	tracingHandler := &TracingHandler{next: mockNext, level: slog.LevelDebug}

	newHandler := tracingHandler.WithAttrs([]slog.Attr{
		slog.String("additional", "info"),
	})

	assert.NotNil(t, newHandler)
}

func TestTracingHandler_WithGroup(t *testing.T) {
	mockNext := &MockHandler{}
	tracingHandler := &TracingHandler{next: mockNext, level: slog.LevelDebug}

	newHandler := tracingHandler.WithGroup("test-group")

	assert.NotNil(t, newHandler)
}

func TestNewTracingHandler(t *testing.T) {
	mockNext := &MockHandler{}
	handler := NewTracingHandler(slog.LevelInfo)(mockNext)

	assert.NotNil(t, handler)
	assert.Equal(t, slog.LevelInfo, handler.(*TracingHandler).level)
	assert.Equal(t, mockNext, handler.(*TracingHandler).next)
}
