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

package middleware // import "github.com/FlyrInc/flyr-lib-go/monitoring/middleware"

import (
	"fmt"
	"net/http"

	internalConfig "github.com/FlyrInc/flyr-lib-go/internal/config"
	"github.com/FlyrInc/flyr-lib-go/internal/utils"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// OtelChiMiddleware returns middleware that will trace incoming requests for the chi web framework.
// The service parameter should describe the name of the (virtual) server handling the request.
func OtelChiMiddleware() func(http.Handler) http.Handler {
	cfg := config{}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	tracer := cfg.TracerProvider.Tracer(
		ScopeName,
		oteltrace.WithInstrumentationVersion("v0.0.1"), // TODO: Update instrumentation version
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}
	if cfg.MonitoringConfig == nil {
		monitoringConfig := internalConfig.NewMonitoringConfig()
		cfg.MonitoringConfig = monitoringConfig
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, f := range cfg.Filters {
				if !f(r) {
					// Serve the request to the next middleware
					// if a filter rejects the request.
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract the context from incoming request headers
			ctx := cfg.Propagators.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			opts := []oteltrace.SpanStartOption{
				oteltrace.WithAttributes(utils.ServerRequestMetrics(cfg.MonitoringConfig.Service(), r)...),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			}

			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			ctx, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()

			// pass the span through the request context
			r = r.WithContext(ctx)

			// wrap the response writer to capture the status code
			ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				status := ww.Status()
				span.SetAttributes(semconv.HTTPStatusCode(status))

				if status >= 500 && status < 600 {
					span.SetAttributes(attribute.String("Error", fmt.Sprintf("%d: %s", status, http.StatusText(status))))
				}
			}()

			// serve the request to the next middleware
			next.ServeHTTP(ww, r)

			// set the route pattern in the span attributes
			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			span.SetAttributes(semconv.HTTPRoute(routePattern))
		})
	}
}
