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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
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

func TestNewHttpClient(t *testing.T) {
	client := NewHttpClient()
	require.NotNil(t, client.Transport)
}

func TestSetHttpTransport_ClientTraces(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		wantConnTrace bool
	}{
		{name: "client traces enabled", envValue: "true", wantConnTrace: true},
		{name: "client traces disabled", envValue: "false", wantConnTrace: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OTEL_ENABLE_HTTP_CLIENT_TRACES", tt.envValue)
			config.ResetMonitoringConfig()
			t.Cleanup(config.ResetMonitoringConfig)

			previous := otel.GetTracerProvider()
			recorder := tracetest.NewSpanRecorder()
			tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
			otel.SetTracerProvider(tp)
			t.Cleanup(func() {
				otel.SetTracerProvider(previous)
				_ = tp.Shutdown(context.Background())
			})

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			t.Cleanup(srv.Close)

			client := SetHttpTransport(http.Client{})
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, http.NoBody)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)

			spans := recorder.Ended()
			var root sdktrace.ReadOnlySpan
			var connSpans int
			for _, s := range spans {
				switch {
				case s.SpanKind() == oteltrace.SpanKindClient && !s.Parent().IsValid():
					root = s
				case s.Name() == "http.getconn":
					connSpans++
				}
			}
			require.NotNil(t, root)
			assert.Equal(t, "HTTP GET", root.Name())

			if tt.wantConnTrace {
				assert.Positive(t, connSpans)
				assert.Greater(t, len(spans), 1)
			} else {
				assert.Zero(t, connSpans)
				assert.Len(t, spans, 1)
			}
		})
	}
}
