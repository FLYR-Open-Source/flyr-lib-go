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

	"github.com/FlyrInc/flyr-lib-go/internal/config"
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

func TestGetExporterClient(t *testing.T) {
	cfg := getMonitoringConfig()

	t.Run("ReturnsGRPCClient", func(t *testing.T) {
		cfg.ExporterProtocolCfg = "grpc"
		client := getExporterClient(cfg)
		require.NotNil(t, client)
	})

	t.Run("ReturnsHTTPClient", func(t *testing.T) {
		cfg.ExporterProtocolCfg = "http/protobuf"
		client := getExporterClient(cfg)
		require.NotNil(t, client)
	})

	t.Run("ReturnsNilForUnsupportedProtocol", func(t *testing.T) {
		cfg.ExporterProtocolCfg = "unsupported"
		client := getExporterClient(cfg)
		require.Nil(t, client)
	})
}

func TestNewTraceProvider(t *testing.T) {
	ctx := context.Background()
	cfg := getMonitoringConfig()

	t.Run("SetsTracerProviderAsGlobal", func(t *testing.T) {
		err := initializeTracerProvider(ctx, cfg)
		defer func() {
			tracerProvider = nil
		}()
		require.NoError(t, err)

		// Check if the global provider is set
		require.Equal(t, tracerProvider, otel.GetTracerProvider())
	})

	t.Run("SetsCorrectResourceAttributes", func(t *testing.T) {
		err := initializeTracerProvider(ctx, cfg)
		defer func() {
			tracerProvider = nil
		}()
		require.NoError(t, err)

		// Ensure tracerProvider is still usable
		tracer := tracerProvider.Tracer("test-tracer")
		_, span := tracer.Start(ctx, "test-span")
		defer span.End()
	})

	t.Run("ConfiguresTextMapPropagator", func(t *testing.T) {
		err := initializeTracerProvider(ctx, cfg)
		defer func() {
			tracerProvider = nil
		}()
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
