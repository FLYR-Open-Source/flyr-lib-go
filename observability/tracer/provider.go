package tracer

import (
	"context"

	"github.com/FlyrInc/flyr-lib-go/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func getOtelCollectorExporter(ctx context.Context, cfg config.MonitoringConfig) (*otlptrace.Exporter, error) {
	return otlptracegrpc.New(ctx, otlptracegrpc.WithEndpointURL(cfg.ExporterEndpoint()))
}

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
		resource.WithOS(),
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

	defaultTracer = &Tracer{tracer: tc.Tracer(cfg.Service())}

	return tc, nil
}

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

	return tc, &Tracer{tracer: tc.Tracer(name)}, nil
}

func StopTracer(ctx context.Context, tc *sdktrace.TracerProvider) {
	tc.Shutdown(ctx)
}
