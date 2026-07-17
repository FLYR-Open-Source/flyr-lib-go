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

package main

import (
	"context"
	"net/http"
	"os"

	internalConfig "github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"github.com/FLYR-Open-Source/flyr-lib-go/logger"
	httpTrace "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/http"
	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/tracer"
	"go.opentelemetry.io/otel"
)

const (
	serviceName = "some-service"
)

func setupEnv() {
	_ = os.Setenv("OTEL_SERVICE_NAME", serviceName)
	// this is a flag for exporting the traces in stdout
	_ = os.Setenv("OTEL_EXPORTER_OTLP_TEST", "true")
	_ = os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "k8s.container.name={some-container},k8s.deployment.name={some-deployment},k8s.deployment.uid={some-uid},k8s.namespace.name={some-namespace},k8s.node.name={some-node},k8s.pod.name={some-pod},k8s.pod.uid={some-uid},k8s.replicaset.name={some-replicaset},k8s.replicaset.uid={some-uid},service.instance.id={some-namespace}.{some-pod}.{some-container},service.version={some-version}")
}

// You don't need this part since it's automated in Kubernetes
func init() {
	setupEnv()
}

func main() {
	ctx := context.Background()

	logger.InitLogger()

	// start the default tracer
	err := tracer.StartDefaultTracer(ctx)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = tracer.ShutdownTracerProvider(ctx)
		if err != nil {
			logger.Error(ctx, "failed to stop tracer", err)
		}
	}()

	withNewClient()

	withExistingClient()

	WithHttpClientTracing()

	WithHttpClientMetrics()

	WithHttpClientSpanAttributes()
}

func withNewClient() {
	ctx := context.Background()
	url := "https://flyr.com/"
	client := httpTrace.NewHttpClient() // start new HTTP client with tracing
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	res, err := client.Do(req)
	if err != nil {
		logger.Error(ctx, "failed to make request", err)
	}

	defer res.Body.Close()
}

func withExistingClient() {
	ctx := context.Background()
	url := "https://flyr.com/"

	client := http.Client{}
	// add any setup to client

	client = httpTrace.SetHttpTransport(client)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := client.Do(req)
	if err != nil {
		logger.Error(ctx, "failed to make request", err)
	}

	defer res.Body.Close()
}

func WithHttpClientTracing() {
	internalConfig.ResetMonitoringConfig()
	setupEnv()
	_ = os.Setenv("OTEL_ENABLE_HTTP_CLIENT_TRACES", "true")

	ctx := context.Background()
	url := "https://flyr.com/"
	client := httpTrace.NewHttpClient() // start new HTTP client with tracing
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	res, err := client.Do(req)
	if err != nil {
		logger.Error(ctx, "failed to make request", err)
	}

	defer res.Body.Close()
}

// WithHttpClientMetrics enables the connection-lifecycle metrics (DNS, connect, TLS,
// get-connection, TTFB, PutIdleConn errors) that otelhttp.NewTransport does not provide
// out of the box. This is independent of OTEL_ENABLE_HTTP_CLIENT_TRACES: either, both, or
// neither can be enabled.
func WithHttpClientMetrics() {
	internalConfig.ResetMonitoringConfig()
	setupEnv()
	_ = os.Setenv("OTEL_ENABLE_HTTP_CLIENT_METRICS", "true")

	ctx := context.Background()
	url := "https://flyr.com/"

	// WithMeterProvider is optional: if omitted, both otelhttp's own request metrics and
	// the connection-lifecycle metrics use the globally registered OpenTelemetry
	// MeterProvider (otel.GetMeterProvider()).
	client := httpTrace.NewHttpClient(httpTrace.WithMeterProvider(otel.GetMeterProvider()))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	res, err := client.Do(req)
	if err != nil {
		logger.Error(ctx, "failed to make request", err)
	}

	defer res.Body.Close()
}

// WithHttpClientSpanAttributes enables the same connection-lifecycle data as
// WithHttpClientMetrics, but as attributes on the request's own span (e.g.
// http.client.getconn.acquired=false surfaces a connection that was requested but never
// obtained) instead of separate metric instruments. This is independent of
// OTEL_ENABLE_HTTP_CLIENT_TRACES and OTEL_ENABLE_HTTP_CLIENT_METRICS: any combination can be
// enabled.
func WithHttpClientSpanAttributes() {
	internalConfig.ResetMonitoringConfig()
	setupEnv()
	_ = os.Setenv("OTEL_ENABLE_HTTP_CLIENT_SPAN_ATTRIBUTES", "true")

	ctx := context.Background()
	url := "https://flyr.com/"
	client := httpTrace.NewHttpClient()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	res, err := client.Do(req)
	if err != nil {
		logger.Error(ctx, "failed to make request", err)
	}

	defer res.Body.Close()
}
