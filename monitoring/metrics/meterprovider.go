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

package metrics

import (
	"context"
	"errors"
	"time"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// ErrMetricsProviderNotInitialized is returned when the metrics provider is not initialized
var ErrMetricsProviderNotInitialized = errors.New("metrics provider not initialized")

// ErrExporterProtocolNotSupported is returned when the exporter protocol is not supported
var ErrExporterProtocolNotSupported = errors.New("exporter protocol not supported")

var (
	meterProvider *sdkmetric.MeterProvider
)

// getExporter returns an OTLP exporter based on the exporter protocol.
// If the exporter protocol is not supported, it returns nil.
//
// Valid values are:
//
// - grpc to use OTLP/gRPC
//
// - http/protobuf to use OTLP/HTTP + protobuf
func getExporter(ctx context.Context, cfg config.MonitoringConfig) (sdkmetric.Exporter, error) {
	// If the the test flag is enabled, return a new stdout exporter
	if cfg.IsTestExporter() {
		return stdoutmetric.New()
	}

	switch cfg.ExporterTracesProtocol() {
	case "grpc":
		return otlpmetricgrpc.New(ctx)
	case "http/protobuf":
		return otlpmetrichttp.New(ctx)
	default:
		return nil, ErrExporterProtocolNotSupported
	}
}

// initializeMetricsProvider creates and configures a new OpenTelemetry MeterProvider.
//
// This function initializes a MeterProvider with the specified configuration,
// including a resource that describes the service, version, environment, and tenant.
// This MeterProvider is also set as the global meter provider for OpenTelemetry.
//
// It returns an error if any occurred.
func initializeMeterProvider(ctx context.Context, cfg config.MonitoringConfig, interval time.Duration) error {
	if cfg.Service() == "" {
		otel.SetMeterProvider(noop.NewMeterProvider())
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
		// TODO: verify if we need it
		resource.WithAttributes(
			attribute.String(config.EXPORTER_PROTOCOL, cfg.ExporterMetricsProtocol()),
		),
	)
	if err != nil {
		return err
	}

	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(interval))),
		sdkmetric.WithResource(resourceInfo),
	)

	otel.SetMeterProvider(meterProvider)

	return nil
}

// ShutdownMeterProvider gracefully shuts down the global MeterProvider.
func ShutdownMeterProvider(ctx context.Context) error {
	if meterProvider == nil {
		return ErrMetricsProviderNotInitialized
	}
	return meterProvider.Shutdown(ctx)
}
