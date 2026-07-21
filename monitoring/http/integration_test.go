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

// This file contains integration-style tests that drive a large volume of concurrent
// requests through SetHttpTransport with metrics, span attributes, and traces all enabled
// together, to catch concurrency bugs in wrapRoundTripper/getConnState (e.g. the mutex
// double-reset panic and lost-span-attribute bugs found and fixed during development) that
// small, sequential tests cannot reliably surface.

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/goleak"
)

// histogramCount sums the Count across all data points (one per distinct attribute set,
// e.g. reused=true/false) for the named Float64 histogram instrument. Returns 0 if the
// instrument was never recorded.
func histogramCount(t *testing.T, reader *sdkmetric.ManualReader, name string) uint64 {
	t.Helper()

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(t.Context(), &rm))

	var total uint64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			hist, ok := m.Data.(metricdata.Histogram[float64])
			if !ok {
				continue
			}
			for _, dp := range hist.DataPoints {
				total += dp.Count
			}
		}
	}
	return total
}

// TestIntegration_ConcurrentRequests drives a large number of concurrent requests through a
// client with metrics, span attributes, and traces all enabled simultaneously, against a
// local httptest.Server. It exists to catch concurrency bugs in wrapRoundTripper and
// getConnState (shared mutable per-request state, hook ordering, mutex handling) that
// wouldn't reliably show up with only a handful of sequential requests — this test is
// intended to always run with -race.
func TestIntegration_ConcurrentRequests(t *testing.T) {
	// Registered first so it runs LAST (t.Cleanup is LIFO), after every other t.Cleanup below
	// has torn down the server, tracer provider, and transport — otherwise it would flag
	// their own long-lived background goroutines (accept loop, batch span processor) as
	// leaks. What it actually guards against is a leak in wrapRoundTripper/getConnState
	// itself: neither allocates a goroutine, so none should remain once every request and
	// every piece of test infrastructure above it has shut down cleanly.
	t.Cleanup(func() { goleak.VerifyNone(t) })

	// OTEL_ENABLE_HTTP_CLIENT_TRACES is deliberately left disabled: it adds its own
	// otelhttptrace sub-spans per request (http.getconn/dns/connect/tls/send/receive),
	// which would make the "one span per request" assertion below about otelhttptrace's
	// behavior rather than about wrapRoundTripper/getConnState, the thing this test targets.
	t.Setenv("OTEL_ENABLE_HTTP_CLIENT_METRICS", "true")
	t.Setenv("OTEL_ENABLE_HTTP_CLIENT_SPAN_ATTRIBUTES", "true")
	config.ResetMonitoringConfig()
	t.Cleanup(config.ResetMonitoringConfig)

	previousTP := otel.GetTracerProvider()
	t.Cleanup(func() { otel.SetTracerProvider(previousTP) })
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	otel.SetTracerProvider(tp)

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(server.Close)

	// A modest per-host cap forces real contention/reuse across the concurrent goroutines
	// (some requests will genuinely wait for a connection to free up), exercising the same
	// GetConn/GotConn interleaving as production traffic, without artificially starving any
	// request into the unacquired path (that failure mode has its own dedicated test).
	base := &http.Transport{MaxConnsPerHost: 20, MaxIdleConnsPerHost: 20}
	client := SetHttpTransport(http.Client{Timeout: 10 * time.Second}, WithMeterProvider(mp), WithBaseTransport(base))
	t.Cleanup(base.CloseIdleConnections)

	const (
		totalRequests = 5000
		concurrency   = 100
	)

	var (
		wg       sync.WaitGroup
		sem      = make(chan struct{}, concurrency)
		errCount int64
		errMu    sync.Mutex
		okCount  int64
		okMu     sync.Mutex
	)

	for range totalRequests {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			resp, err := client.Get(server.URL)
			if err != nil {
				errMu.Lock()
				errCount++
				errMu.Unlock()
				return
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode == http.StatusOK {
				okMu.Lock()
				okCount++
				okMu.Unlock()
			}
		}()
	}
	wg.Wait()

	require.Equal(t, int64(0), errCount, "no requests should fail against a healthy local server")
	require.Equal(t, int64(totalRequests), okCount)

	// Every request goes through GetConn/GotConn (fresh or reused), so getconn.duration
	// must have exactly one observation per request — this is the sharpest possible check
	// that wrapRoundTripper/getConnState never double-counts, drops, or cross-contaminates
	// state between concurrent requests.
	require.EqualValues(t, totalRequests, histogramCount(t, reader, "http.client.getconn.duration"),
		"expected exactly one http.client.getconn.duration observation per request")
	require.EqualValues(t, totalRequests, histogramCount(t, reader, "http.client.ttfb.duration"),
		"expected exactly one http.client.ttfb.duration observation per request")
	// otelhttp's own request-level metric, sharing the same MeterProvider, should also match.
	require.EqualValues(t, totalRequests, histogramCount(t, reader, "http.client.request.duration"))

	require.NoError(t, tp.ForceFlush(context.Background()))
	spans := exporter.GetSpans()
	require.Len(t, spans, totalRequests, "expected exactly one span per request, got %d", len(spans))

	for _, span := range spans {
		attrs := map[string]bool{}
		for _, kv := range span.Attributes {
			attrs[string(kv.Key)] = true
		}
		require.True(t, attrs[attrGetConnAcquired], "span %s missing %s", span.Name, attrGetConnAcquired)
		require.True(t, attrs[attrGetConnDuration], "span %s missing %s", span.Name, attrGetConnDuration)
		require.True(t, attrs[attrTTFBDuration], "span %s missing %s", span.Name, attrTTFBDuration)
	}
}

// TestIntegration_NoGoroutineLeak isolates the goroutine-leak check from the load test above
// (which needs its own top-function ignore list for the transport's own persistent
// goroutines) by running a small, finite number of sequential requests and asserting no
// goroutines related to this package's wrapRoundTripper/getConnState are left running.
func TestIntegration_NoGoroutineLeak(t *testing.T) {
	t.Setenv("OTEL_ENABLE_HTTP_CLIENT_METRICS", "true")
	config.ResetMonitoringConfig()
	t.Cleanup(config.ResetMonitoringConfig)

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	base := &http.Transport{}
	client := SetHttpTransport(http.Client{}, WithMeterProvider(mp), WithBaseTransport(base))

	before := runtime.NumGoroutine()

	for i := 0; i < 200; i++ {
		resp, err := client.Get(server.URL)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
	}

	base.CloseIdleConnections()
	// Allow the transport's own read-loop goroutines time to exit after CloseIdleConnections.
	require.Eventually(t, func() bool {
		return runtime.NumGoroutine() <= before+2 // small slack for GC/runtime bookkeeping goroutines
	}, 2*time.Second, 10*time.Millisecond, "goroutine count did not return to baseline after requests completed")
}
