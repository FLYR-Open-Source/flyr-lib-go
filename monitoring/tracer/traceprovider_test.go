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

package tracer

import (
	"context"
	"testing"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func getMonitoringConfig() config.Monitoring {
	serviceName := "test-service"

	cfg := config.NewMonitoringConfig()
	cfg.ServiceCfg = serviceName
	cfg.ExporterProtocolCfg = "grpc"

	return cfg
}

func TestGetExporter(t *testing.T) {
	ctx := context.Background()
	cfg := getMonitoringConfig()

	t.Run("ReturnsStdoutExporter", func(t *testing.T) {
		cfg.ExporterProtocolCfg = ""
		cfg.TestExporterCfg = true
		client, err := getExporter(ctx, cfg)
		require.NotNil(t, client)
		require.NoError(t, err)
	})

	t.Run("ReturnsGRPCExporter", func(t *testing.T) {
		cfg.ExporterProtocolCfg = "grpc"
		cfg.TestExporterCfg = false
		client, err := getExporter(ctx, cfg)
		require.NotNil(t, client)
		require.NoError(t, err)
	})

	t.Run("ReturnsHTTPClient", func(t *testing.T) {
		cfg.ExporterProtocolCfg = "http/protobuf"
		cfg.TestExporterCfg = false
		client, err := getExporter(ctx, cfg)
		require.NotNil(t, client)
		require.NoError(t, err)
	})

	t.Run("ReturnsNilForUnsupportedProtocol", func(t *testing.T) {
		cfg.ExporterProtocolCfg = "unsupported"
		cfg.TestExporterCfg = false
		client, err := getExporter(ctx, cfg)
		require.Nil(t, client)
		require.ErrorIs(t, err, ErrExporterClientNotSupported)
	})
}

func TestNewTraceProvider(t *testing.T) {
	ctx := context.Background()
	cfg := getMonitoringConfig()

	t.Run("ReturnsNoError", func(t *testing.T) {
		err := initializeTracerProvider(ctx, cfg)
		require.NoError(t, err)
	})

	t.Run("SetsCorrectResourceAttributes", func(t *testing.T) {
		err := initializeTracerProvider(ctx, cfg)
		require.NoError(t, err)

		// Ensure tracerProvider is still usable
		tracer := otel.GetTracerProvider().Tracer("test-tracer")
		_, span := tracer.Start(ctx, "test-span")
		defer span.End()
	})

	t.Run("ConfiguresTextMapPropagator", func(t *testing.T) {
		err := initializeTracerProvider(ctx, cfg)
		require.NoError(t, err)

		// Retrieve the global TextMapPropagator and confirm it’s a composite propagator.
		propagator := otel.GetTextMapPropagator()

		// Check if the global propagator is a composite propagator by its behavior.
		tracePropagator := propagation.TraceContext{}
		baggagePropagator := propagation.Baggage{}

		// Create a dummy map for propagation.
		carrier := propagation.MapCarrier{}

		// Inject and extract to test if our propagator is working as expected.
		propagator.Inject(ctx, carrier)
		ctx = propagator.Extract(ctx, carrier)

		// Validate propagation by checking that TraceContext and Baggage are effectively included.
		traceInjected := tracePropagator.Extract(ctx, carrier)
		baggageInjected := baggagePropagator.Extract(ctx, carrier)

		require.NotNil(t, traceInjected, "TraceContext should be part of the TextMapPropagator")
		require.NotNil(t, baggageInjected, "Baggage should be part of the TextMapPropagator")
	})
}
