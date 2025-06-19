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

package middleware // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/middleware"

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func WithExcludingEndpoints(filters []string) configOption {
	return optionFunc(func(cfg config) {

		for _, f := range filters {
			cfg.filters = append(cfg.filters, func(r *http.Request) bool {
				return r.URL.Path != f
			})
		}
	})
}

func OtelTraceMiddleware(opts ...configOption) func(http.ResponseWriter, *http.Request, func()) {
	cfg := newConfig()

	// make sure to apply the options to the config
	for _, o := range opts {
		o.apply(cfg)
	}

	return func(w http.ResponseWriter, r *http.Request, next func()) {
		ew := extendResponseWriter(w)

		for _, f := range cfg.filters {
			if !f(r) {
				// Serve the request to the next middleware if a filter rejects the request.
				next()
				return
			}
		}

		// Save the original context to restore it after the request is handled
		savedCtx := r.Context()
		defer func() {
			r = r.WithContext(savedCtx)
		}()

		// Extract the context from incoming request headers
		ctx := otel.GetTextMapPropagator().Extract(savedCtx, propagation.HeaderCarrier(r.Header))

		opts := []oteltrace.SpanStartOption{
			oteltrace.WithAttributes(utils.ServerRequestMetrics(cfg.monitoringCfg.Service(), r)...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}

		spanName := fmt.Sprintf("%s %s", strings.ToUpper(r.Method), r.URL.Path)
		ctx, span := cfg.tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// pass the span through the request context
		r = r.WithContext(ctx)

		// Capture the status code from the response writer after the request is handled
		defer func() {
			ew.Done()

			// do nothing if the span is not recording
			if !span.IsRecording() {
				return
			}

			span.SetAttributes(semconv.HTTPRoute(r.Pattern))

			status := ew.StatusCode
			code, descr := utils.HTTPServerStatus(status)
			span.SetStatus(code, descr)
			if status > 0 {
				span.SetAttributes(semconv.HTTPStatusCode(status))
			}
			if status >= 500 && status < 600 {
				span.SetAttributes(attribute.String("Error", fmt.Sprintf("%d: %s", status, http.StatusText(status))))
			}
		}()

		// serve the request to the next middleware
		next()
	}
}
