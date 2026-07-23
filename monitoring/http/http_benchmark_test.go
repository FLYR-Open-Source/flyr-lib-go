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

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// benchmarkServer returns a local httptest.Server so benchmark results reflect this
// package's own overhead, not network latency.
func benchmarkServer(b *testing.B) *httptest.Server {
	b.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	b.Cleanup(server.Close)
	return server
}

// runClientBenchmark issues repeated GETs against server through client, reusing the same
// (warmed-up) connection every iteration, so the measured cost is steady-state per-request
// overhead rather than connection setup.
func runClientBenchmark(b *testing.B, client http.Client, server *httptest.Server) {
	b.Helper()

	// Warm up: establish and cache the pooled connection before the timed loop starts, so
	// GetConn/DNS/connect/TLS one-time setup cost isn't attributed to per-request overhead.
	warmup, err := client.Get(server.URL)
	if err != nil {
		b.Fatal(err)
	}
	_ = warmup.Body.Close()

	b.ReportAllocs()
	for b.Loop() {
		resp, err := client.Get(server.URL)
		if err != nil {
			b.Fatal(err)
		}
		_ = resp.Body.Close()
	}
}

// BenchmarkClient_Bare measures a plain http.Client with no instrumentation at all — the
// baseline every other benchmark in this file is compared against.
func BenchmarkClient_Bare(b *testing.B) {
	server := benchmarkServer(b)
	client := http.Client{}
	runClientBenchmark(b, client, server)
}

// BenchmarkClient_OtelhttpOnly measures otelhttp.NewTransport with no flyr-lib-go
// connection-lifecycle instrumentation enabled — isolates otelhttp's own baseline overhead
// (its http.client.request.duration/body.size metrics and span) from what this package adds
// on top in the benchmarks below.
func BenchmarkClient_OtelhttpOnly(b *testing.B) {
	config.ResetMonitoringConfig()
	b.Cleanup(config.ResetMonitoringConfig)

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewManualReader()))
	b.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	server := benchmarkServer(b)
	client := SetHttpTransport(http.Client{}, WithMeterProvider(mp))
	runClientBenchmark(b, client, server)
}

// BenchmarkClient_MetricsEnabled measures the added overhead of
// OTEL_ENABLE_HTTP_CLIENT_METRICS (the httptrace hooks in metrics.go, plus the
// wrapRoundTripper layer) on top of otelhttp's own instrumentation.
func BenchmarkClient_MetricsEnabled(b *testing.B) {
	b.Setenv("OTEL_ENABLE_HTTP_CLIENT_METRICS", "true")
	config.ResetMonitoringConfig()
	b.Cleanup(config.ResetMonitoringConfig)

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewManualReader()))
	b.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	server := benchmarkServer(b)
	client := SetHttpTransport(http.Client{}, WithMeterProvider(mp))
	runClientBenchmark(b, client, server)
}

// BenchmarkClient_SpanAttributesEnabled measures the added overhead of
// OTEL_ENABLE_HTTP_CLIENT_SPAN_ATTRIBUTES (the httptrace hooks in spans.go) on top of
// otelhttp's own instrumentation.
func BenchmarkClient_SpanAttributesEnabled(b *testing.B) {
	b.Setenv("OTEL_ENABLE_HTTP_CLIENT_SPAN_ATTRIBUTES", "true")
	config.ResetMonitoringConfig()
	b.Cleanup(config.ResetMonitoringConfig)

	previousTP := otel.GetTracerProvider()
	b.Cleanup(func() { otel.SetTracerProvider(previousTP) })
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(tracetest.NewInMemoryExporter()))
	b.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	otel.SetTracerProvider(tp)

	server := benchmarkServer(b)
	client := SetHttpTransport(http.Client{})
	runClientBenchmark(b, client, server)
}

// BenchmarkClient_AllEnabled measures the combined overhead of traces, metrics, and span
// attributes all enabled together — the worst case for how much this package's
// instrumentation can add to a single request.
func BenchmarkClient_AllEnabled(b *testing.B) {
	b.Setenv("OTEL_ENABLE_HTTP_CLIENT_TRACES", "true")
	b.Setenv("OTEL_ENABLE_HTTP_CLIENT_METRICS", "true")
	b.Setenv("OTEL_ENABLE_HTTP_CLIENT_SPAN_ATTRIBUTES", "true")
	config.ResetMonitoringConfig()
	b.Cleanup(config.ResetMonitoringConfig)

	previousTP := otel.GetTracerProvider()
	b.Cleanup(func() { otel.SetTracerProvider(previousTP) })
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(tracetest.NewInMemoryExporter()))
	b.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	otel.SetTracerProvider(tp)

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewManualReader()))
	b.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	server := benchmarkServer(b)
	client := SetHttpTransport(http.Client{}, WithMeterProvider(mp))
	runClientBenchmark(b, client, server)
}
