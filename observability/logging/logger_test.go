package logging

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func TestDatadogLogger_Handle_withoutTrace(t *testing.T) {
	l := slog.New(DatadogLogger{jsonLogHandler("info")})

	l.Info("test")
}

func TestDatadogLogger_Handle_withTrace(t *testing.T) {
	l := slog.New(DatadogLogger{jsonLogHandler("info")})

	ctx := context.Background()
	span := ddtracer.StartSpan("test")
	ctx = ddtracer.ContextWithSpan(ctx, span)

	l.InfoContext(ctx, "test")
}

func Test_extractTrace_traceNotFound(t *testing.T) {
	ctx := context.Background()

	traceID, spanID, found := extractTrace(ctx)

	assert.False(t, found)
	assert.Equal(t, "", traceID)
	assert.Equal(t, "", spanID)
}

func Test_extractTrace_traceFound(t *testing.T) {
	ctx := context.Background()
	span := ddtracer.StartSpan("test")
	ctx = ddtracer.ContextWithSpan(ctx, span)

	traceID, spanID, found := extractTrace(ctx)

	assert.True(t, found)
	assert.Equal(t, "0", traceID)
	assert.Equal(t, "0", spanID)
}

func Test_parseLogLevel(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected slog.Level
	}{
		"debug works":                   {"debug", slog.LevelDebug},
		"Debug works":                   {"Debug", slog.LevelDebug},
		"DEBUG works":                   {"DEBUG", slog.LevelDebug},
		"info works":                    {"info", slog.LevelInfo},
		"warn works":                    {"warn", slog.LevelWarn},
		"error works":                   {"error", slog.LevelError},
		"unknown defaults to info":      {"unknown", slog.LevelInfo},
		"empty string defaults to info": {"", slog.LevelInfo},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := parseLogLevel(test.input)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestSetDefaultLoggerByLevel(t *testing.T) {
	SetDefaultLoggerByLevel("info")
}
