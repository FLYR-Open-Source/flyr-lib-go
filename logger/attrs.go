package logger

import (
	"context"
	"log/slog"

	"github.com/FlyrInc/flyr-lib-go/config"
	internalUtils "github.com/FlyrInc/flyr-lib-go/internal/utils"
)

const (
	// The depth of the caller in the stack trace
	callerDepth = 3
)

func getAttributes(ctx context.Context, err error, args ...interface{}) []slog.Attr {
	metadata := slog.Group(config.LOG_METADATA_KEY, args...)
	injectAttrsToSpan(ctx, metadata)

	caller := internalUtils.GetCallerName(callerDepth)
	callerAttrs := caller.LogAttributes()

	attrs := append(callerAttrs, metadata)

	var errorMessage slog.Attr
	if err != nil {
		errorMessage = slog.String(config.LOG_ERROR_KEY, err.Error())
		setErroredSpan(ctx, err) // Set spans as errored
		attrs = append(attrs, errorMessage)
	}

	return attrs
}
