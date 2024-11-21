package traceprovider

import (
	"context"
	"errors"

	"github.com/FlyrInc/flyr-lib-go/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// ErrServiceNameNotSet is returned when the service name is not set in the configuration
var ErrServiceNameNotSet = errors.New("service name not set")

// ErrTracerProviderNotInitialized is returned when the tracer provider is not initialized
var ErrTracerProviderNotInitialized = errors.New("tracer provider not initialized")

// ErrTracerProviderAlreadyInitialized is returned when the tracer provider is already initialized
var ErrTracerProviderAlreadyInitialized = errors.New("tracer provider already initialized")

var (
	tracerProvider *sdktrace.TracerProvider
	resourceInfo   *resource.Resource
)

// initializeTracerProvider creates and configures a new OpenTelemetry TracerProvider.
//
// This function initializes a TracerProvider with the specified configuration,
// including a resource that describes the service, version, environment, and tenant.
// It also sets up the exporter for exporting trace data. If the defaultProvider flag
// This TracerProvider is also set as the global tracer provider for OpenTelemetry.
//
// Furthermore, it configures a composite text map propagator for trace context and baggage
// propagation, which is essential for distributed tracing.
//
// It returns the initialized TracerProvider and an error if any occurred.
func InitializeTracerProvider(ctx context.Context, cfg config.MonitoringConfig) error {
	if tracerProvider != nil {
		return ErrTracerProviderAlreadyInitialized
	}

	if cfg.Service() == "" {
		return ErrServiceNameNotSet
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpointURL(cfg.ExporterEndpoint()))
	if err != nil {
		return err
	}

	resourceInfo, err = resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String(config.DEPLOYMENT_ENVIRONMENT, cfg.Env()),
			attribute.String(config.SERVICE_NAME, cfg.Service()),
			attribute.String(config.SERVICE_VERSION, cfg.Version()),

			attribute.String(config.CUSTON_ENV_NAME, cfg.Env()),
			attribute.String(config.CUSTOM_TENANT_NAME, cfg.Tenant()),
		),
		// resource.WithFromEnv(), // TODO: Uncomment this line when the Otel environment variables are set to the pods
		resource.WithTelemetrySDK(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return err
	}

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resourceInfo),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	return nil
}

// ShutdownTracerProvider gracefully shuts down the global TracerProvider.
func ShutdownTracerProvider(ctx context.Context) error {
	if tracerProvider == nil {
		return ErrTracerProviderNotInitialized
	}
	return tracerProvider.Shutdown(ctx)
}
