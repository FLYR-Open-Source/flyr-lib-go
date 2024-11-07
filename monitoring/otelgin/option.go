// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Based on https://github.com/DataDog/dd-trace-go/blob/8fb554ff7cf694267f9077ae35e27ce4689ed8b6/contrib/gin-gonic/gin/option.go

package otelgin // import "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type config struct {
	TracerProvider    oteltrace.TracerProvider
	Propagators       propagation.TextMapPropagator
	Filters           []filter
	GinFilters        []ginFilter
	SpanNameFormatter spanNameFormatter
}

// Filter is a predicate used to determine whether a given http.request should
// be traced. A Filter must return true if the request should be traced.
type filter func(*http.Request) bool

// Adding new Filter parameter (*gin.Context)
// gin.Context has FullPath() method, which returns a matched route full path.
type ginFilter func(*gin.Context) bool

// SpanNameFormatter is used to set span name by http.request.
type spanNameFormatter func(r *http.Request) string
