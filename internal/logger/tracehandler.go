package logger

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/FlyrInc/flyr-lib-go/internal/span"
)

const (
	// traceIDKey is the key used to store the trace id in the log record
	traceIDKey = "trace_id"
	// spanIDKey is the key used to store the span id in the log record
	spanIDKey = "span_id"
)

type TracingHandler struct {
	// next is the next handler in the chain
	next slog.Handler
	// level is the minimum level of log that will be handled
	level slog.Level
}

// Enabled returns true if the log level is greater than or equal to the handler's level
func (h *TracingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle adds the trace and span ids to the log record and passes it to the next handler
func (h *TracingHandler) Handle(ctx context.Context, record slog.Record) error {
	if h.next == nil {
		return errors.New("handler is missing")
	}

	traceId, spanId, found := span.ExtractTrace(ctx)

	if found {
		record.AddAttrs(slog.String(traceIDKey, traceId))
		record.AddAttrs(slog.String(spanIDKey, spanId))
	}

	return h.next.Handle(ctx, record)
}

// WithAttrs returns a new handler with the given attributes added to the log record
func (h *TracingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewTracingHandler(h.next.WithAttrs(attrs), strings.ToLower(h.level.String()))
}

// WithGroup returns a new handler with the given group name added to the log record
func (h *TracingHandler) WithGroup(name string) slog.Handler {
	return NewTracingHandler(h.next.WithGroup(name), strings.ToLower(h.level.String()))
}

// WithLevel returns a new handler with the given log level
func (h *TracingHandler) Handler() slog.Handler {
	return h.next
}

// NewTracingHandler creates a new TracingHandler with the specified log level.
//
// This function wraps a given slog.Handler with a TracingHandler, which
// can be used to add tracing information to log entries. If the provided
// handler is already a TracingHandler, it extracts the underlying handler
// to avoid redundant wrapping. The log level is parsed and set to control
// the tracing output level.
//
// Returns a pointer to the new TracingHandler.
func NewTracingHandler(next slog.Handler, level string) *TracingHandler {
	if th, ok := next.(*TracingHandler); ok {
		next = th.Handler()
	}

	return &TracingHandler{next: next, level: parseLogLevel(level)}
}
