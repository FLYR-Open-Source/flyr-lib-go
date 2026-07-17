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

package http // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/http"

import (
	"context"
	"net/http"
	"net/http/httptrace"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// transportConfig holds the configuration applied by Option functions.
type transportConfig struct {
	meterProvider metric.MeterProvider
	baseTransport http.RoundTripper
}

// Option configures the behavior of SetHttpTransport and NewHttpClient.
type Option func(*transportConfig)

// WithMeterProvider sets the metric.MeterProvider used both by otelhttp's own request
// metrics and by this package's connection-lifecycle metrics (when
// OTEL_ENABLE_HTTP_CLIENT_METRICS is enabled). If not provided, the globally registered
// OpenTelemetry MeterProvider (otel.GetMeterProvider()) is used for both.
func WithMeterProvider(provider metric.MeterProvider) Option {
	return func(c *transportConfig) {
		if provider != nil {
			c.meterProvider = provider
		}
	}
}

// WithBaseTransport sets the http.RoundTripper that otelhttp.NewTransport (and this
// package's connection-lifecycle instrumentation) wraps. If not provided, http.DefaultTransport
// is used. Use this to configure the underlying *http.Transport's connection pool (e.g.
// MaxConnsPerHost, MaxIdleConnsPerHost, IdleConnTimeout) instead of sharing the process-wide
// http.DefaultTransport.
func WithBaseTransport(base http.RoundTripper) Option {
	return func(c *transportConfig) {
		if base != nil {
			c.baseTransport = base
		}
	}
}

// getConnStateKey is the context key used to hand the per-request *getConnState from
// wrapRoundTripper's RoundTrip down to the otelhttp.WithClientTrace callback, so both can
// observe the same connection-acquisition timings for a request. This relies on
// context.Context values propagating through derivation (otelhttp derives its own ctx from
// the one passed to RoundTrip via context.WithValue-based calls, it never replaces it
// wholesale), which holds for otelhttp's current implementation.
type getConnStateKey struct{}

// SetHttpTransport configures the provided HTTP client to use OpenTelemetry's transport for tracing.
//
// This function takes an http.Client as an argument and sets its Transport to an
// OpenTelemetry-enabled transport created by otelhttp.NewTransport. This allows for
// tracing of outgoing HTTP requests made by the client, enabling better observability
// and monitoring of requests in a distributed system.
//
// OTEL_ENABLE_HTTP_CLIENT_TRACES, OTEL_ENABLE_HTTP_CLIENT_METRICS, and
// OTEL_ENABLE_HTTP_CLIENT_SPAN_ATTRIBUTES are independent: any combination can be enabled.
//   - OTEL_ENABLE_HTTP_CLIENT_TRACES adds DNS/connect/TLS/get-connection sub-spans.
//   - OTEL_ENABLE_HTTP_CLIENT_METRICS adds connection-lifecycle metrics (DNS, connect,
//     TLS, get-connection, TTFB, PutIdleConn errors, and a getconn.unacquired counter for
//     connections that were requested but never obtained) that are not otherwise emitted
//     by otelhttp.NewTransport (which only records http.client.request.duration and
//     http.client.request.body.size).
//   - OTEL_ENABLE_HTTP_CLIENT_SPAN_ATTRIBUTES adds the same connection-lifecycle data as
//     attributes on the request's own span (see otelhttp.NewTransport), rather than as
//     separate metric instruments — useful for seeing where one specific, individual
//     request spent its time, which an aggregated metric cannot show.
//
// Both otelhttp's own metrics and this package's connection-lifecycle metrics use the same
// metric.MeterProvider: by default the globally registered OpenTelemetry MeterProvider
// (otel.GetMeterProvider()), or the one passed via WithMeterProvider.
//
// By default the transport wraps http.DefaultTransport. Pass WithBaseTransport to use a
// differently configured *http.Transport instead (e.g. to set MaxConnsPerHost,
// MaxIdleConnsPerHost, or IdleConnTimeout) rather than sharing the process-wide default.
//
// Returns the configured http.Client with the OpenTelemetry transport set.
func SetHttpTransport(client http.Client, opts ...Option) http.Client {
	tc := &transportConfig{meterProvider: otel.GetMeterProvider(), baseTransport: http.DefaultTransport}
	for _, opt := range opts {
		opt(tc)
	}

	transportOpts := []otelhttp.Option{otelhttp.WithMeterProvider(tc.meterProvider)}

	cfg := config.NewMonitoringConfig()
	enableTraces := cfg.EnableHttpClientTraces()
	enableMetrics := cfg.EnableHttpClientMetrics()
	enableSpanAttributes := cfg.EnableHttpClientSpanAttributes()

	var metrics *clientMetrics
	if enableMetrics {
		if m, err := newClientMetrics(tc.meterProvider); err == nil {
			metrics = m
		} else {
			// TODO: log error if metrics creation fails? This is a non-fatal failure, but it may be worth logging to help diagnose why metrics aren't being emitted.
		}
	}

	if enableTraces || metrics != nil || enableSpanAttributes {
		transportOpts = append(transportOpts, otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
			var next *httptrace.ClientTrace
			if enableTraces {
				next = otelhttptrace.NewClientTrace(ctx)
			}
			if enableSpanAttributes {
				next = spanAttributeTrace(ctx, next)
			}

			state, _ := ctx.Value(getConnStateKey{}).(*getConnState)
			if metrics == nil || state == nil {
				// otelhttp passes whatever is returned here straight to
				// httptrace.WithClientTrace, which panics on nil.
				if next != nil {
					return next
				}
				return &httptrace.ClientTrace{}
			}

			return newClientMetricTrace(ctx, state, metrics, next)
		}))
	}

	base := otelhttp.NewTransport(tc.baseTransport, transportOpts...)

	if metrics != nil {
		client.Transport = wrapRoundTripper(base, metrics)
	} else {
		client.Transport = base
	}

	return client
}

// wrapRoundTripper wraps base so a *getConnState is created per request and made available
// to the otelhttp.WithClientTrace callback (via getConnStateKey), and so that once the
// request finishes, the unacquired-connection metric can be recorded — httptrace.ClientTrace
// has no "request done" hook to record it from directly. Span attributes (see spans.go) do
// not need this: they are set incrementally from within the httptrace hooks themselves,
// since otelhttp may end the request's span before RoundTrip returns (on error), and
// span.SetAttributes silently no-ops on an already-ended span.
func wrapRoundTripper(base http.RoundTripper, metrics *clientMetrics) http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		state := &getConnState{}
		ctx := context.WithValue(req.Context(), getConnStateKey{}, state)
		req = req.WithContext(ctx)

		resp, err := base.RoundTrip(req)

		if started, acquired := state.acquired(); started && !acquired {
			metrics.recordUnacquired(ctx)
		}

		return resp, err
	})
}

// roundTripperFunc adapts a function to the http.RoundTripper interface.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// NewHttpClient initializes a new HTTP client with OpenTelemetry tracing enabled.
//
// This function creates and returns an http.Client configured to use an OpenTelemetry
// transport by wrapping the default HTTP transport. This allows for tracing of all
// outgoing HTTP requests made by the client, providing enhanced observability for
// applications that rely on external HTTP communications.
//
// Returns a new http.Client with OpenTelemetry tracing configured.
func NewHttpClient(opts ...Option) http.Client {
	client := http.Client{}
	return SetHttpTransport(client, opts...)
}
