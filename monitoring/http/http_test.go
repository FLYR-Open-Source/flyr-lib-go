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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestSetHttpTransport(t *testing.T) {
	client := http.Client{}
	client = SetHttpTransport(client)

	require.NotNil(t, client.Transport)
}

func TestSetHttpTransport_WithHttpClientTracesEnabled(t *testing.T) {
	t.Setenv("OTEL_ENABLE_HTTP_CLIENT_TRACES", "true")
	config.ResetMonitoringConfig()
	t.Cleanup(config.ResetMonitoringConfig)

	client := http.Client{}
	client = SetHttpTransport(client)

	require.NotNil(t, client.Transport)
}

func TestSetHttpTransport_WithHttpClientMetricsEnabled_NoOptionGiven(t *testing.T) {
	t.Setenv("OTEL_ENABLE_HTTP_CLIENT_METRICS", "true")
	config.ResetMonitoringConfig()
	t.Cleanup(config.ResetMonitoringConfig)

	// No MeterProvider is registered globally or passed via WithMeterProvider, so
	// instrument creation may fail against the no-op default. SetHttpTransport must still
	// return a usable client either way.
	client := http.Client{}
	client = SetHttpTransport(client)

	require.NotNil(t, client.Transport)
}

// instrumentNames collects all metric instrument names exported to reader's scope.
func instrumentNames(t *testing.T, reader *sdkmetric.ManualReader) map[string]bool {
	t.Helper()

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(t.Context(), &rm))

	names := map[string]bool{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			names[m.Name] = true
		}
	}
	return names
}

func TestSetHttpTransport_WithHttpClientMetricsEnabled(t *testing.T) {
	t.Setenv("OTEL_ENABLE_HTTP_CLIENT_METRICS", "true")
	config.ResetMonitoringConfig()
	t.Cleanup(config.ResetMonitoringConfig)

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(server.Close)

	client := SetHttpTransport(http.Client{}, WithMeterProvider(mp))
	require.NotNil(t, client.Transport)

	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, "ok", string(body))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// A second request against the same client should reuse the pooled connection,
	// exercising the GotConn(Reused=true) path instead of the DNS/connect/TLS phases.
	resp2, err := client.Get(server.URL)
	require.NoError(t, err)
	require.NoError(t, resp2.Body.Close())
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	names := instrumentNames(t, reader)
	require.True(t, names["http.client.getconn.duration"], "expected http.client.getconn.duration to be recorded, got: %v", names)
	require.True(t, names["http.client.ttfb.duration"], "expected http.client.ttfb.duration to be recorded, got: %v", names)
	// otelhttp's own request-level metric should also be present, sharing the same
	// MeterProvider passed via WithMeterProvider.
	require.True(t, names["http.client.request.duration"], "expected otelhttp's http.client.request.duration to share the same MeterProvider, got: %v", names)
}

func TestSetHttpTransport_WithBothTracesAndMetricsEnabled(t *testing.T) {
	t.Setenv("OTEL_ENABLE_HTTP_CLIENT_TRACES", "true")
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

	client := SetHttpTransport(http.Client{}, WithMeterProvider(mp))
	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNewHttpClient(t *testing.T) {
	client := NewHttpClient()
	require.NotNil(t, client.Transport)
}

func TestNewHttpClient_WithMeterProvider(t *testing.T) {
	t.Setenv("OTEL_ENABLE_HTTP_CLIENT_METRICS", "true")
	config.ResetMonitoringConfig()
	t.Cleanup(config.ResetMonitoringConfig)

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	client := NewHttpClient(WithMeterProvider(mp))
	require.NotNil(t, client.Transport)
}

func TestSetHttpTransport_WithSpanAttributesEnabled(t *testing.T) {
	t.Setenv("OTEL_ENABLE_HTTP_CLIENT_SPAN_ATTRIBUTES", "true")
	config.ResetMonitoringConfig()
	t.Cleanup(config.ResetMonitoringConfig)

	previousTP := otel.GetTracerProvider()
	t.Cleanup(func() { otel.SetTracerProvider(previousTP) })

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	otel.SetTracerProvider(tp)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := SetHttpTransport(http.Client{})
	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusOK, resp.StatusCode)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := map[string]bool{}
	for _, kv := range spans[0].Attributes {
		attrs[string(kv.Key)] = true
	}
	require.True(t, attrs[attrGetConnAcquired], "expected %s attribute on the request span, got: %v", attrGetConnAcquired, attrs)
	require.True(t, attrs[attrGetConnDuration], "expected %s attribute on the request span, got: %v", attrGetConnDuration, attrs)
}

func TestSetHttpTransport_UnacquiredConnection(t *testing.T) {
	t.Setenv("OTEL_ENABLE_HTTP_CLIENT_METRICS", "true")
	t.Setenv("OTEL_ENABLE_HTTP_CLIENT_SPAN_ATTRIBUTES", "true")
	config.ResetMonitoringConfig()
	t.Cleanup(config.ResetMonitoringConfig)

	previousTP := otel.GetTracerProvider()
	t.Cleanup(func() { otel.SetTracerProvider(previousTP) })
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	otel.SetTracerProvider(tp)

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	block := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(func() {
		close(block)
		server.Close()
	})

	// A transport limited to a single connection per host: the first (long-running)
	// request holds the only connection, so a second request blocks at GetConn until its
	// context expires, without GotConn ever firing. This is the exact "stuck waiting for
	// connection" failure mode neither httptrace nor otelhttp can otherwise signal.
	base := &http.Transport{MaxConnsPerHost: 1}
	client := SetHttpTransport(http.Client{}, WithMeterProvider(mp), WithBaseTransport(base))

	firstStarted := make(chan struct{})
	go func() {
		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		close(firstStarted)
		resp, err := client.Do(req) //nolint:bodyclose // best-effort background request, cleaned up by t.Cleanup closing block
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
	<-firstStarted
	time.Sleep(50 * time.Millisecond) // give the first request time to occupy the only connection

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	_, err := client.Do(req)
	require.Error(t, err, "expected the second request to fail waiting for a connection")

	names := instrumentNames(t, reader)
	require.True(t, names["http.client.getconn.unacquired"], "expected http.client.getconn.unacquired to be recorded, got: %v", names)

	var sawUnacquired bool
	for _, span := range exporter.GetSpans() {
		for _, kv := range span.Attributes {
			if string(kv.Key) == attrGetConnAcquired && !kv.Value.AsBool() {
				sawUnacquired = true
			}
		}
	}
	require.True(t, sawUnacquired, "expected a span with %s=false", attrGetConnAcquired)
}
