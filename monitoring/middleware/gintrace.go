// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Based on https://github.com/DataDog/dd-trace-go/blob/8fb554ff7cf694267f9077ae35e27ce4689ed8b6/contrib/gin-gonic/gin/gintrace.go

package middleware // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/middleware"

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	internalConfig "github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"github.com/FLYR-Open-Source/flyr-lib-go/internal/utils"
	"github.com/FLYR-Open-Source/flyr-lib-go/internal/version"
)

// OtelGinMiddleware returns middleware that will trace incoming requests for the gin web framework.
// The service parameter should describe the name of the (virtual) server handling the request.
func OtelGinMiddleware() gin.HandlerFunc {
	cfg := config{}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	tracer := cfg.TracerProvider.Tracer(
		ScopeName,
		oteltrace.WithInstrumentationVersion(version.Version()),
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}
	if cfg.MonitoringConfig == nil {
		monitoringConfig := internalConfig.NewMonitoringConfig()
		cfg.MonitoringConfig = monitoringConfig
	}

	return func(c *gin.Context) {
		for _, f := range cfg.Filters {
			if !f(c.Request) {
				// Serve the request to the next middleware
				// if a filter rejects the request.
				c.Next()
				return
			}
		}
		c.Set(tracerKey, tracer)
		savedCtx := c.Request.Context()
		defer func() {
			c.Request = c.Request.WithContext(savedCtx)
		}()

		// Extract the context from incoming request headers
		ctx := cfg.Propagators.Extract(savedCtx, propagation.HeaderCarrier(c.Request.Header))
		opts := []oteltrace.SpanStartOption{
			oteltrace.WithAttributes(utils.ServerRequestMetrics(cfg.MonitoringConfig.Service(), c.Request)...),
			oteltrace.WithAttributes(semconv.HTTPRoute(c.FullPath())),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}

		spanName := c.FullPath()
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", c.Request.Method)
		}
		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// pass the span through the request context
		c.Request = c.Request.WithContext(ctx)

		// serve the request to the next middleware
		c.Next()

		status := c.Writer.Status()
		code, descr := utils.HTTPServerStatus(status)
		span.SetStatus(code, descr)
		if status > 0 {
			span.SetAttributes(semconv.HTTPStatusCode(status))
		}
		if status >= 500 && status < 600 {
			span.SetAttributes(attribute.String("Error", fmt.Sprintf("%d: %s", status, http.StatusText(status))))
		}
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("gin.errors", c.Errors.String()))
		}
	}
}
