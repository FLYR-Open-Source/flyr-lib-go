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

func getAttributes(ctx context.Context, err error, args ...interface{}) []interface{} {
	metadata := slog.Group(config.LOG_METADATA_KEY, args...)
	injectAttrsToSpan(ctx, metadata)

	caller := internalUtils.GetCallerName(callerDepth)
	codeFilePath := slog.String(config.CODE_PATH, caller.FilePath)
	codeLineNumber := slog.Int(config.CODE_LINE, caller.LineNumber)
	codeFunctionName := slog.String(config.CODE_FUNC, caller.FunctionName)
	codeNamespace := slog.String(config.CODE_NS, caller.Namespace)

	attrs := []interface{}{codeFilePath, codeLineNumber, codeFunctionName, codeNamespace, metadata}

	var errorMessage slog.Attr
	if err != nil {
		errorMessage = slog.String(config.LOG_ERROR_KEY, err.Error())
		setErroredSpan(ctx, err) // Set spans as errored
		attrs = append(attrs, errorMessage)
	}

	return attrs
}
