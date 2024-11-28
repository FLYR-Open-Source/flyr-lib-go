package config // import "github.com/FlyrInc/flyr-lib-go/internal/config"

import (
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// attribute names for the common tags that will be applied to every trace, span and log
const (
	SERVICE_NAME       = string(semconv.ServiceNameKey)
	SERVICE_VERSION    = string(semconv.ServiceVersionKey)
	SERVICE_INTANCE_ID = string(semconv.ServiceInstanceIDKey)
)

// attribute names for the caller tags that will be applied to every trace, span and log
const (
	FILE_PATH             = string(semconv.CodeFilepathKey)
	LINE_NUMBER           = string(semconv.CodeLineNumberKey)
	FUNCTION_NAME         = string(semconv.CodeFunctionKey)
	FUNCTION_PACKAGE_NAME = string(semconv.CodeNamespaceKey)
)

// attribute names for the logger
const (
	LOG_MESSAGE_KEY  = "message"
	LOG_ERROR_KEY    = "error"
	LOG_METADATA_KEY = "metadata"
)
