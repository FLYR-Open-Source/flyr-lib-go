package main

import (
	"context"
	"fmt"

	"github.com/FlyrInc/flyr-lib-go/config"
	"github.com/FlyrInc/flyr-lib-go/monitoring/tracer"
	"go.opentelemetry.io/otel/trace"
)

type MyStruct struct {
	Name string
	Age  int
}

const (
	// You can pass the `OBSERVABILITY_SERVICE` environment variable to set the service name
	serviceName = "some-service"
	// You can pass the `OBSERVABILITY_ENV` environment variable to set the environment
	env = "dev"
	// You can pass the `OBSERVABILITY_FLYR_TENANT` environment variable to set the tenant
	flyrTenant = "fl"
	// You can pass the `OBSERVABILITY_VERSION` environment variable to set the version
	version = "v1.0.0"
	// You can pass the `OBSERVABILITY_TRACER_ENABLED` environment variable to enable the tracer
	enableTracer = true
)

func getMonitoringConfig() config.MonitoringConfig {
	cfg := config.NewMonitoringConfig()
	cfg.EnableTracer = enableTracer
	cfg.ServiceCfg = serviceName
	cfg.EnvCfg = env
	cfg.FlyrTenantCfg = flyrTenant
	cfg.VersionCfg = version

	return cfg
}

func main() {
	ctx := context.Background()

	// start the tracer for the service layer
	serviceLauyerTC, serviceLayerTracer, err := tracer.StartCustomTracer(ctx, getMonitoringConfig(), "service_layer")
	if err != nil {
		panic(err)
	}

	// start the tracer for the repository layer
	repositoryLauyerTC, repositoryLayerTracer, err := tracer.StartCustomTracer(ctx, getMonitoringConfig(), "repository_layer")
	if err != nil {
		panic(err)
	}

	defer func() {
		if serviceLauyerTC != nil {
			err := tracer.StopTracer(ctx, serviceLauyerTC)
			if err != nil {
				fmt.Println("failed to stop tracer", err)
			}
		}

		if repositoryLauyerTC != nil {
			err := tracer.StopTracer(ctx, repositoryLauyerTC)
			if err != nil {
				fmt.Println("failed to stop tracer", err)
			}
		}
	}()

	// someServiceLayer with be wrapped with a span.
	// We pass the serviceLayerTracer to the function so that the span is created with the service layer tracer.
	someServiceLayer(ctx, serviceLayerTracer)

	// someRepositoryLayer with be wrapped with a span.
	// We pass the repositoryLayerTracer to the function so that the span is created with the repository layer tracer.
	someRepositoryLayer(ctx, repositoryLayerTracer)
}

func someServiceLayer(ctx context.Context, tracer *tracer.Tracer) {
	_, span := tracer.StartSpan(ctx, "my_super_service_fn", trace.SpanKindInternal)
	defer span.End()

	// do some work
}

func someRepositoryLayer(ctx context.Context, tracer *tracer.Tracer) {
	_, span := tracer.StartSpan(ctx, "my_super_repository_fn", trace.SpanKindInternal)
	defer span.End()

	// do some work
}
