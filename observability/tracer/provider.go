package tracer

import (
	"context"
	"errors"

	"github.com/FlyrInc/flyr-lib-go/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// ErrServiceNameNotSet is returned when the service name is not set in the configuration
var ErrServiceNameNotSet = errors.New("service name not set")

// getOtelCollectorExporter returns an OTLP exporter that exports to the OpenTelemetry Collector
func getOtelCollectorExporter(ctx context.Context, cfg config.MonitoringConfig) (*otlptrace.Exporter, error) {
	return otlptracegrpc.New(ctx, otlptracegrpc.WithEndpointURL(cfg.ExporterEndpoint()))
}

// newTraceProvider creates and configures a new OpenTelemetry TracerProvider.
//
// This function initializes a TracerProvider with the specified configuration,
// including a resource that describes the service, version, environment, and tenant.
// It also sets up the exporter for exporting trace data. If the defaultProvider flag
// is true, it sets this TracerProvider as the global tracer provider for OpenTelemetry.
//
// It also configures a composite text map propagator for trace context and baggage
// propagation, which is essential for distributed tracing.
//
// It returns the initialized TracerProvider and an error if any occurred.
func newTraceProvider(ctx context.Context, cfg config.MonitoringConfig, exporter *otlptrace.Exporter, defaultProvider bool) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.Service()),
			attribute.String(config.VERSION_NAME, cfg.Version()),
			attribute.String(config.ENV_NAME, cfg.Env()),
			attribute.String(config.TENANT_NAME, cfg.Tenant()),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, err
	}

	tc := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	if defaultProvider {
		otel.SetTracerProvider(tc)
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tc, nil
}

// StartDefaultTracer initializes and starts the default OpenTelemetry TracerProvider.
//
// This function checks if tracing is enabled in the provided configuration. If tracing
// is enabled, it retrieves the OpenTelemetry collector exporter and creates a new
// TracerProvider using the newTraceProvider function. It also validates that the
// service name is set in the configuration. If the service name is not provided,
// it returns an error indicating that the service name must be set.
//
// The function also sets the global default tracer to be used for tracing in the
// application. If tracing is not enabled, it returns nil without starting a tracer.
//
// It returns the initialized TracerProvider and an error if any occurred.
func StartDefaultTracer(ctx context.Context, cfg config.MonitoringConfig) (*sdktrace.TracerProvider, error) {
	if !cfg.TracerEnabled() {
		return nil, nil
	}

	exporter, err := getOtelCollectorExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}

	tc, err := newTraceProvider(ctx, cfg, exporter, true)
	if err != nil {
		return nil, err
	}

	if cfg.Service() == "" {
		return nil, ErrServiceNameNotSet
	}

	defaultTracer = &Tracer{
		tracer: tc.Tracer(
			cfg.Service(),
			oteltrace.WithInstrumentationVersion("1"),
		),
	}

	return tc, nil
}

// StartCustomTracer initializes and starts a custom OpenTelemetry TracerProvider with a specified name.
//
// This function checks if tracing is enabled in the provided configuration. If tracing
// is enabled, it retrieves the OpenTelemetry collector exporter and creates a new
// TracerProvider using the newTraceProvider function. The function does not set the
// default provider, allowing for multiple tracers to coexist with different configurations.
//
// The custom tracer is returned alongside the TracerProvider, enabling applications to
// create spans using the specified tracer name. If tracing is not enabled, it returns
// nil values without starting a tracer.
func StartCustomTracer(ctx context.Context, cfg config.MonitoringConfig, name string) (*sdktrace.TracerProvider, *Tracer, error) {
	if !cfg.TracerEnabled() {
		return nil, nil, nil
	}

	exporter, err := getOtelCollectorExporter(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	tc, err := newTraceProvider(ctx, cfg, exporter, false)
	if err != nil {
		return nil, nil, err
	}

	t := &Tracer{
		tracer: tc.Tracer(
			name,
			oteltrace.WithInstrumentationVersion("1"),
		),
	}

	return tc, t, nil
}

// StopTracer gracefully shuts down the provided OpenTelemetry TracerProvider.
func StopTracer(ctx context.Context, tc *sdktrace.TracerProvider) error {
	return tc.Shutdown(ctx)
}
