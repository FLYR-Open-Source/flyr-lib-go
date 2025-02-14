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

package logger // import "github.com/FlyrInc/flyr-lib-go/internal/logger"

import (
	"context"
	"errors"
	"log/slog"

	"github.com/FlyrInc/flyr-lib-go/internal/span"
	slogmulti "github.com/samber/slog-multi"
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

	// Do not add trace and span ids to debug logs
	if record.Level <= slog.LevelDebug {
		return h.next.Handle(ctx, record)
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
	return &TracingHandler{
		next:  h.next.WithAttrs(attrs),
		level: h.level,
	}
}

// WithGroup returns a new handler with the given group name added to the log record
func (h *TracingHandler) WithGroup(name string) slog.Handler {
	return &TracingHandler{
		next:  h.next.WithGroup(name),
		level: h.level,
	}
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
// Returns an slogmulti.Middleware
func NewTracingHandler(level slog.Level) slogmulti.Middleware {
	return func(next slog.Handler) slog.Handler {
		return &TracingHandler{
			next:  next,
			level: level,
		}
	}
}
