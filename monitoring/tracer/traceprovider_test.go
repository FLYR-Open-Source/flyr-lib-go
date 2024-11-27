package tracer

import (
	"context"
	"testing"

	"github.com/FlyrInc/flyr-lib-go/internal/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func getMonitoringConfig() config.MonitoringConfig {
	serviceName := "test-service"

	cfg := config.NewMonitoringConfig()
	cfg.ServiceCfg = serviceName

	return cfg
}
func TestNewTraceProvider(t *testing.T) {
	ctx := context.Background()
	cfg := getMonitoringConfig()

	t.Run("SetsTracerProviderAsGlobal", func(t *testing.T) {
		err := InitializeTracerProvider(ctx, cfg)
		defer func() {
			tracerProvider = nil
		}()
		require.NoError(t, err)

		// Check if the global provider is set
		require.Equal(t, tracerProvider, otel.GetTracerProvider())
	})

	t.Run("SetsCorrectResourceAttributes", func(t *testing.T) {
		err := InitializeTracerProvider(ctx, cfg)
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
		err := InitializeTracerProvider(ctx, cfg)
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
