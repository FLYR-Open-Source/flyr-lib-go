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

package logger

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// initLoggerCapturingStdout redirects os.Stdout to a pipe, runs InitLogger (the JSON handler binds
// os.Stdout at construction time) and returns a function that stops the capture and returns the
// JSON log lines written so far. slog.Default and os.Stdout are restored on cleanup.
func initLoggerCapturingStdout(t *testing.T) func() []map[string]interface{} {
	t.Helper()

	previousDefault := slog.Default()
	originalStdout := os.Stdout
	t.Cleanup(func() {
		os.Stdout = originalStdout
		slog.SetDefault(previousDefault)
	})

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	InitLogger()

	return func() []map[string]interface{} {
		t.Helper()

		os.Stdout = originalStdout
		require.NoError(t, w.Close())
		raw, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NoError(t, r.Close())

		var lines []map[string]interface{}
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			if line == "" {
				continue
			}
			var entry map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(line), &entry), "invalid JSON log line: %s", line)
			lines = append(lines, entry)
		}
		return lines
	}
}

// newRecordingSpan starts a span on an SDK tracer provider backed by a span recorder.
func newRecordingSpan(t *testing.T) (context.Context, sdktrace.ReadWriteSpan, *tracetest.SpanRecorder) {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	ctx, span := tp.Tracer("logger-test").Start(context.Background(), "test-span")
	rw, ok := span.(sdktrace.ReadWriteSpan)
	require.True(t, ok)

	return ctx, rw, recorder
}

func spanAttrs(span sdktrace.ReadOnlySpan) map[string]string {
	result := make(map[string]string)
	for _, kv := range span.Attributes() {
		result[string(kv.Key)] = kv.Value.String()
	}
	return result
}

func TestInitLogger(t *testing.T) {
	t.Run("Installs a JSON default logger", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("OTEL_SERVICE_NAME", "init-logger-test")
		stop := initLoggerCapturingStdout(t)

		slog.Info("direct slog call", "k", "v")

		lines := stop()
		require.Len(t, lines, 1)
		assert.Equal(t, "direct slog call", lines[0]["message"])
		assert.Equal(t, "INFO", lines[0]["level"])
		assert.Equal(t, "init-logger-test", lines[0]["service.name"])
		assert.Equal(t, "v", lines[0]["k"])
	})

	tests := []struct {
		name      string
		logLevel  string
		wantLevel []string
	}{
		{name: "debug level logs everything", logLevel: "debug", wantLevel: []string{"DEBUG", "INFO", "WARN", "ERROR"}},
		{name: "info level drops debug", logLevel: "info", wantLevel: []string{"INFO", "WARN", "ERROR"}},
		{name: "warn level drops debug and info", logLevel: "warn", wantLevel: []string{"WARN", "ERROR"}},
		{name: "error level only logs errors", logLevel: "error", wantLevel: []string{"ERROR"}},
		{name: "invalid level defaults to info", logLevel: "bogus", wantLevel: []string{"INFO", "WARN", "ERROR"}},
	}

	for _, tt := range tests {
		t.Run("Respects LOG_LEVEL: "+tt.name, func(t *testing.T) {
			t.Setenv("LOG_LEVEL", tt.logLevel)
			stop := initLoggerCapturingStdout(t)
			ctx := context.Background()

			Debug(ctx, "debug message")
			Info(ctx, "info message")
			Warn(ctx, "warn message")
			Error(ctx, "error message", errors.New("boom"))

			var got []string
			for _, line := range stop() {
				level, ok := line["level"].(string)
				require.True(t, ok)
				got = append(got, level)
			}
			assert.Equal(t, tt.wantLevel, got)
		})
	}
}

func TestLogFunctionsOutput(t *testing.T) {
	tests := []struct {
		name      string
		log       func(ctx context.Context)
		wantLevel string
		wantMsg   string
		wantError string
	}{
		{
			name:      "Debug",
			log:       func(ctx context.Context) { Debug(ctx, "debug msg", "id", 1) },
			wantLevel: "DEBUG",
			wantMsg:   "debug msg",
		},
		{
			name:      "Info",
			log:       func(ctx context.Context) { Info(ctx, "info msg", "id", 1) },
			wantLevel: "INFO",
			wantMsg:   "info msg",
		},
		{
			name:      "Warn",
			log:       func(ctx context.Context) { Warn(ctx, "warn msg", "id", 1) },
			wantLevel: "WARN",
			wantMsg:   "warn msg",
		},
		{
			name:      "Error",
			log:       func(ctx context.Context) { Error(ctx, "error msg", errors.New("kaboom"), "id", 1) },
			wantLevel: "ERROR",
			wantMsg:   "error msg",
			wantError: "kaboom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LOG_LEVEL", "debug")
			stop := initLoggerCapturingStdout(t)

			tt.log(context.Background())

			lines := stop()
			require.Len(t, lines, 1)
			line := lines[0]

			assert.Equal(t, tt.wantMsg, line["message"])
			assert.Equal(t, tt.wantLevel, line["level"])
			assert.JSONEq(t, `{"id":1}`, mustJSON(t, line["metadata"]))
			assert.Contains(t, line, "code.filepath")
			assert.Contains(t, line, "code.lineno")
			assert.Contains(t, line, "time")
			// no span in the context, so no correlation ids
			assert.NotContains(t, line, "trace_id")
			assert.NotContains(t, line, "span_id")

			if tt.wantError != "" {
				assert.Equal(t, tt.wantError, line["error"])
			} else {
				assert.NotContains(t, line, "error")
			}
		})
	}
}

func mustJSON(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

func TestLogFunctionsWithRecordingSpan(t *testing.T) {
	nested := map[string]interface{}{"a": "b", "deep": map[string]interface{}{"c": 3}}

	t.Run("Info and Warn add trace ids and inject flattened attributes into the span", func(t *testing.T) {
		tests := []struct {
			name string
			log  func(ctx context.Context)
		}{
			{name: "Info", log: func(ctx context.Context) { Info(ctx, "msg", "id", 7, "nested", nested) }},
			{name: "Warn", log: func(ctx context.Context) { Warn(ctx, "msg", "id", 7, "nested", nested) }},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Setenv("LOG_LEVEL", "debug")
				stop := initLoggerCapturingStdout(t)
				ctx, span, _ := newRecordingSpan(t)

				tt.log(ctx)

				lines := stop()
				require.Len(t, lines, 1)
				assert.Equal(t, span.SpanContext().TraceID().String(), lines[0]["trace_id"])
				assert.Equal(t, span.SpanContext().SpanID().String(), lines[0]["span_id"])

				attrs := spanAttrs(span)
				assert.Equal(t, "7", attrs["metadata.id"])
				assert.Equal(t, "b", attrs["metadata.nested.a"])
				assert.Equal(t, "3", attrs["metadata.nested.deep.c"])
				// status must be untouched
				assert.Equal(t, codes.Unset, span.Status().Code)
			})
		}
	})

	t.Run("Debug neither adds trace ids nor injects attributes", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "debug")
		stop := initLoggerCapturingStdout(t)
		ctx, span, _ := newRecordingSpan(t)

		Debug(ctx, "debug msg", "id", 7)

		lines := stop()
		require.Len(t, lines, 1)
		assert.Equal(t, "DEBUG", lines[0]["level"])
		assert.NotContains(t, lines[0], "trace_id")
		assert.NotContains(t, lines[0], "span_id")
		assert.JSONEq(t, `{"id":7}`, mustJSON(t, lines[0]["metadata"]))
		assert.Empty(t, spanAttrs(span))
	})

	t.Run("Error marks the span errored and records the error event", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "info")
		stop := initLoggerCapturingStdout(t)
		ctx, span, _ := newRecordingSpan(t)

		Error(ctx, "failed", errors.New("db down"), "id", 7)

		lines := stop()
		require.Len(t, lines, 1)
		assert.Equal(t, "db down", lines[0]["error"])
		assert.Equal(t, span.SpanContext().TraceID().String(), lines[0]["trace_id"])
		assert.Equal(t, span.SpanContext().SpanID().String(), lines[0]["span_id"])

		assert.Equal(t, codes.Error, span.Status().Code)
		assert.Equal(t, "db down", span.Status().Description)
		assert.Equal(t, "7", spanAttrs(span)["metadata.id"])

		require.Len(t, span.Events(), 1)
		event := span.Events()[0]
		assert.Equal(t, "exception", event.Name)
		assert.Contains(t, event.Attributes, attribute.String("exception.message", "db down"))
	})

	t.Run("Error with a nil error does not mark the span errored", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "info")
		stop := initLoggerCapturingStdout(t)
		ctx, span, _ := newRecordingSpan(t)

		Error(ctx, "no error value", nil)

		lines := stop()
		require.Len(t, lines, 1)
		assert.NotContains(t, lines[0], "error")
		assert.Equal(t, codes.Unset, span.Status().Code)
		assert.Empty(t, span.Events())
	})
}

func TestAttributeWithOutInjectingAttrsToSpan(t *testing.T) {
	ctx, span, _ := newRecordingSpan(t)

	attrs := NewAttribute().
		WithOutInjectingAttrsToSpan().
		WithMetadata("id", 7).
		Get(ctx)

	assert.NotEmpty(t, attrs)
	assert.Empty(t, spanAttrs(span))

	// the default builder does inject
	NewAttribute().WithMetadata("id", 7).Get(ctx)
	assert.Equal(t, "7", spanAttrs(span)["metadata.id"])
}
