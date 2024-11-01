package logger

import (
	"context"
	"os"
	"time"

	"github.com/FlyrInc/flyr-lib-go/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type Logger struct {
	zerolog.Logger
}

// WithContext returns a copy of ctx with the logger attached.
func (l Logger) WithContext(ctx context.Context) context.Context {
	return l.Logger.WithContext(ctx)
}

// New creates a logger with passed options and default settings to allow for contextual logging
func New(cfg config.Logger) Logger {
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack // Error Logging with Stacktrace
	level := zerolog.InfoLevel                           // default to info
	if lvl, err := zerolog.ParseLevel(cfg.LogLevel()); err == nil && lvl != zerolog.NoLevel {
		level = lvl
	}
	zerolog.SetGlobalLevel(level)
	logger := zerolog.New(os.Stdout).
		Level(level).
		With().   // Add contextual fields to the global logger
		Caller(). // Add file and line number to log
		Timestamp().
		Str(config.ENV_NAME, cfg.Env()).
		Str(config.VERSION_NAME, cfg.Version()).
		Str(config.SERVICE_NAME, cfg.Service()).
		Str(config.TENANT_NAME, cfg.Tenant()).
		Logger().
		Hook(TracingHook{}) // hook trace and span IDs to the logs

	zerolog.DefaultContextLogger = &logger
	return Logger{logger}
}

// Ctx returns the logger associated with given ctx if available
// If ctx does not have logger, return zerolog.DefaultContextLogger
// Otherwise falls back to zerolog.disabledLogger
func Ctx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
