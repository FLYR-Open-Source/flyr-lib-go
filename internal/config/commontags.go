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
// SOFTWARE.

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

// exporter attributes
const (
	EXPORTER_PROTOCOL = "otel.exporter.protocol"
)
