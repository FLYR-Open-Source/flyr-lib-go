package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/FlyrInc/flyr-lib-go/config"
)

// NewJSONLogHandler creates a new JSON log handler with custom configurations.
//
// This function initializes a slog.Handler that outputs logs in JSON format
// to standard output. The handler options include adding the source of log
// entries, setting the log level according to the provided configuration,
// and replacing certain attributes using the replaceAttributes function for
// custom formatting.
func NewJSONLogHandler(cfg config.Logger) slog.Handler {
	return slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			AddSource:   true,
			Level:       parseLogLevel(cfg.LogLevel()),
			ReplaceAttr: replaceAttributes,
		})
}

// parseLogLevel converts a string to a slog.Level.
//
// If the string is not a valid log level, it defaults to info.
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
