package tracer // import "github.com/FlyrInc/flyr-lib-go/tracer"

import (
	"context"
	"errors"

	"github.com/FlyrInc/flyr-lib-go/internal/config"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// ErrTracerProviderNotInitialized is returned when the tracer provider is not initialized
var ErrTracerProviderNotInitialized = errors.New("tracer provider not initialized")

// ErrTracerProviderAlreadyInitialized is returned when the tracer provider is already initialized
var ErrTracerProviderAlreadyInitialized = errors.New("tracer provider already initialized")

var (
	tracerProvider *sdktrace.TracerProvider
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
	if cfg.Service() == "" {
		otel.SetTracerProvider(noop.NewTracerProvider())
		return nil
	}

	exporter, err := otlptrace.New(ctx, otlptracehttp.NewClient())

	if err != nil {
		return err
	}

	resourceInfo, err := resource.New(
		ctx,
		resource.WithFromEnv(),
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
