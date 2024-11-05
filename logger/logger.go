package logger

import (
	"context"

	"github.com/FlyrInc/flyr-lib-go/config"
	internalLogger "github.com/FlyrInc/flyr-lib-go/internal/logger"

	"log/slog"
)

// InitLogger initializes the logger with the given configuration.
//
// The logger is then selected as the default logger for the application.
func InitLogger(cfg config.LoggerConfig) {
	jsonHanlder := internalLogger.NewJSONLogHandler(cfg)
	tracingHanlder := internalLogger.NewTracingHandler(jsonHanlder, cfg.LogLevel())

	l := slog.New(internalLogger.InjectRootAttrs(tracingHanlder, cfg))
	slog.SetDefault(l)
}

// Debug logs a message at the debug level.
//
// Any attributes passed as arguments are added to the log message in the group "metadata",
// and in the span that is retrieved from the given context.
func Debug(ctx context.Context, message string, args ...interface{}) {
	metadata := slog.Group(internalLogger.METADATA_KEY, args...)
	slog.DebugContext(ctx, message, metadata)
}

// Info logs a message at the info level.
//
// Any attributes passed as arguments are added to the log message in the group "metadata",
// and in the span that is retrieved from the given context.
func Info(ctx context.Context, message string, args ...interface{}) {
	metadata := slog.Group(internalLogger.METADATA_KEY, args...)

	injectAttrsToSpan(ctx, metadata)
	slog.InfoContext(ctx, message, metadata)
}

// Warn logs a message at the warn level.
//
// Any attributes passed as arguments are added to the log message in the group "metadata",
// and in the span that is retrieved from the given context.
func Warn(ctx context.Context, message string, args ...interface{}) {
	metadata := slog.Group(internalLogger.METADATA_KEY, args...)
	slog.WarnContext(ctx, message, metadata)
}

// Error logs a message at the error level.
//
// Any attributes passed as arguments are added to the log message in the group "metadata",
// and in the span that is retrieved from the given context.
// Furthermore, if an error is passed as an argument, it is added to the log message in the attribute "error",
// and also sets the span as errored (if the a span cna be retrieved from the given context).
func Error(ctx context.Context, message string, err error, args ...interface{}) {
	var errorMessage slog.Attr
	if err != nil {
		errorMessage = slog.String(internalLogger.ERROR_KEY, err.Error())
		setErroredSpan(ctx, err) // Set spans as errored
	}

	metadata := slog.Group(internalLogger.METADATA_KEY, args...)
	slog.ErrorContext(ctx, message, errorMessage, metadata)
}
