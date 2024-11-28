package logger // import "github.com/FlyrInc/flyr-lib-go/logger"

import (
	"context"

	"github.com/FlyrInc/flyr-lib-go/internal/config"
	internalLogger "github.com/FlyrInc/flyr-lib-go/internal/logger"
	slogmulti "github.com/samber/slog-multi"

	"log/slog"
)

// InitLogger initializes the logger with the given configuration.
//
// The logger is then selected as the default logger for the application.
func InitLogger() {
	cfg := config.NewLoggerConfig()
	loggerLevel := internalLogger.ParseLogLevel(cfg.LogLevel())

	jsonHanlder := internalLogger.NewJSONLogHandler(loggerLevel)
	tracingHanlder := internalLogger.NewTracingHandler(loggerLevel)
	sink := internalLogger.InjectRootAttrs(jsonHanlder, cfg)

	l := slog.New(
		slogmulti.
			Pipe(tracingHanlder).
			Handler(sink),
	)

	slog.SetDefault(l)
}

// Debug logs a message at the debug level.
//
// Any attributes passed as arguments are added to the log message in the group "metadata",
// and in the span that is retrieved from the given context.
func Debug(ctx context.Context, message string, args ...interface{}) {
	l := slog.Default()
	attrs := NewAttribute().
		WithOutInjectingAttrsToSpan().
		WithMetadata(args...).
		Get(ctx)
	l.LogAttrs(ctx, slog.LevelDebug, message, attrs...)
}

// Info logs a message at the info level.
//
// Any attributes passed as arguments are added to the log message in the group "metadata",
// and in the span that is retrieved from the given context.
func Info(ctx context.Context, message string, args ...interface{}) {
	l := slog.Default()
	attrs := NewAttribute().
		WithMetadata(args...).
		Get(ctx)
	l.LogAttrs(ctx, slog.LevelInfo, message, attrs...)
}

// Warn logs a message at the warn level.
//
// Any attributes passed as arguments are added to the log message in the group "metadata",
// and in the span that is retrieved from the given context.
func Warn(ctx context.Context, message string, args ...interface{}) {
	l := slog.Default()
	attrs := NewAttribute().
		WithMetadata(args...).
		Get(ctx)
	l.LogAttrs(ctx, slog.LevelWarn, message, attrs...)
}

// Error logs a message at the error level.
//
// Any attributes passed as arguments are added to the log message in the group "metadata",
// and in the span that is retrieved from the given context.
// Furthermore, if an error is passed as an argument, it is added to the log message in the attribute "error",
// and also sets the span as errored (if the a span cna be retrieved from the given context).
func Error(ctx context.Context, message string, err error, args ...interface{}) {
	l := slog.Default()
	attrs := NewAttribute().
		WithMetadata(args...).
		WithError(err).
		Get(ctx)
	l.LogAttrs(ctx, slog.LevelError, message, attrs...)
}
