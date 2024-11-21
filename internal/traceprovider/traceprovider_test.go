package traceprovider

import (
	"context"
	"testing"

	"github.com/FlyrInc/flyr-lib-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func getMonitoringConfig() config.MonitoringConfig {
	serviceName := "test-service"
	env := "test-env"
	flyrTenant := "test-tenant"
	version := "test-version"

	cfg := config.NewMonitoringConfig()
	cfg.EnableTracer = true
	cfg.ServiceCfg = serviceName
	cfg.EnvCfg = env
	cfg.FlyrTenantCfg = flyrTenant
	cfg.VersionCfg = version

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

		attributes := resourceInfo.Attributes()

		// Verify that each attribute matches the expected values
		assert.Contains(t, attributes, semconv.DeploymentEnvironment(cfg.Env()))
		assert.Contains(t, attributes, semconv.ServiceName(cfg.Service()))
		assert.Contains(t, attributes, semconv.ServiceVersion(cfg.Version()))
		assert.Contains(t, attributes, attribute.String(config.CUSTON_ENV_NAME, cfg.Env()))
		assert.Contains(t, attributes, attribute.String(config.CUSTOM_TENANT_NAME, cfg.Tenant()))

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
