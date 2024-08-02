package logging

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func SetDefaultLoggerByLevel(level string) {
	l := slog.New(DatadogLogger{jsonLogHandler(level)})
	slog.SetDefault(l)
}

type DatadogLogger struct {
	slog.Handler
}

func (h DatadogLogger) Handle(ctx context.Context, r slog.Record) error {
	traceID, spanID, spanFound := extractTrace(ctx)

	if spanFound {
		r.Add("dd.trace_id", slog.StringValue(traceID))
		r.Add("dd.span_id", slog.StringValue(spanID))
	}

	return h.Handler.Handle(ctx, r)
}

func extractTrace(ctx context.Context) (traceID, spanID string, found bool) {
	span, foundSpan := tracer.SpanFromContext(ctx)
	if !foundSpan {
		return "", "", false
	}

	traceID = strconv.Itoa(int(span.Context().TraceID()))
	spanID = strconv.Itoa(int(span.Context().SpanID()))

	return traceID, spanID, true
}

func jsonLogHandler(level string) *slog.JSONHandler {
	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       parseLogLevel(level),
		ReplaceAttr: nil,
	})
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
