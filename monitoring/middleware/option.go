// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Based on https://github.com/DataDog/dd-trace-go/blob/8fb554ff7cf694267f9077ae35e27ce4689ed8b6/contrib/gin-gonic/gin/option.go

package middleware // import "github.com/FlyrInc/flyr-lib-go/monitoring/middleware"

import (
	"net/http"

	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"

	internalConfig "github.com/FlyrInc/flyr-lib-go/internal/config"
)

const (
	tracerKey = "otel-http-tracer"
	// ScopeName is the instrumentation scope name.
	ScopeName = "request"
)

type config struct {
	TracerProvider   oteltrace.TracerProvider
	Propagators      propagation.TextMapPropagator
	MonitoringConfig internalConfig.MonitoringConfig
	Filters          []filter
}

// Filter is a predicate used to determine whether a given http.request should
// be traced. A Filter must return true if the request should be traced.
type filter func(*http.Request) bool
