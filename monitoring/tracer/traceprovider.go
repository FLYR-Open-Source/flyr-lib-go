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

package tracer // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/tracer"

import (
	"context"
	"errors"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// ErrTracerProviderNotInitialized is returned when the tracer provider is not initialized
var ErrTracerProviderNotInitialized = errors.New("tracer provider not initialized")

// ErrTracerProviderAlreadyInitialized is returned when the tracer provider is already initialized
var ErrTracerProviderAlreadyInitialized = errors.New("tracer provider already initialized")

// ErrExporterClientNotSupported is returned when the exporter client is not supported
var ErrExporterClientNotSupported = errors.New("exporter client not supported")

// getExporter returns an OTLP exporter based on the exporter protocol.
// If the exporter protocol is not supported, it returns nil.
//
// If the test flag is enabled, it returns a new stdout exporter.
//
// Valid values are:
//
// - grpc to use OTLP/gRPC
//
// - http/protobuf to use OTLP/HTTP + protobuf
func getExporter(ctx context.Context, cfg config.MonitoringConfig) (sdktrace.SpanExporter, error) {
	// If the the test flag is enabled, return a new stdout exporter
	if cfg.IsTestExporter() {
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	}

	switch cfg.ExporterTracesProtocol() {
	case "grpc":
		client := otlptracegrpc.NewClient()
		return otlptrace.New(ctx, client)
	case "http/protobuf":
		client := otlptracehttp.NewClient()
		return otlptrace.New(ctx, client)
	default:
		return nil, ErrExporterClientNotSupported
	}
}

// initializeTracerProvider creates and configures a new OpenTelemetry TracerProvider.
//
// This function initializes a TracerProvider with the specified configuration,
// including a resource that describes the service, version, environment, and tenant.
// This TracerProvider is also set as the global tracer provider for OpenTelemetry.
//
// Furthermore, it configures a composite text map propagator for trace context and baggage
// propagation, which is essential for distributed tracing.
//
// It returns an error if any occurred.
func initializeTracerProvider(ctx context.Context, cfg config.MonitoringConfig) error {
	if cfg.Service() == "" {
		otel.SetTracerProvider(noop.NewTracerProvider())
		return nil
	}

	exporter, err := getExporter(ctx, cfg)
	if err != nil {
		return err
	}

	resourceInfo, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			attribute.String(config.EXPORTER_PROTOCOL, cfg.ExporterTracesProtocol()),
		),
	)
	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resourceInfo),
	)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	return nil
}

// ShutdownTracerProvider gracefully shuts down the global TracerProvider.
func ShutdownTracerProvider(ctx context.Context) error {
	tp := otel.GetTracerProvider()

	if tp == nil {
		return ErrTracerProviderNotInitialized
	}

	tc, ok := tp.(*sdktrace.TracerProvider)
	if !ok {
		return nil
	}

	return tc.Shutdown(ctx)
}
