package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/FlyrInc/flyr-lib-go/config"
	"github.com/FlyrInc/flyr-lib-go/logger"
	"github.com/FlyrInc/flyr-lib-go/monitoring/tracer"
	"go.opentelemetry.io/otel/trace"
)

type MyStruct struct {
	Name string
	Age  int
}

const (
	serviceName = "some-service"
	env         = "dev"
	flyrTenant  = "fl"
	version     = "v1.0.0"
)

func getMonitoringConfig() config.MonitoringConfig {
	cfg := config.NewMonitoringConfig()
	cfg.EnableTracer = true
	cfg.ServiceCfg = serviceName
	cfg.EnvCfg = env
	cfg.FlyrTenantCfg = flyrTenant
	cfg.VersionCfg = version

	return cfg
}

func getLoggingConfig() config.LoggerConfig {
	cfg := config.NewLoggerConfig()
	cfg.ServiceCfg = serviceName
	cfg.EnvCfg = env
	cfg.FlyrTenantCfg = flyrTenant
	cfg.VersionCfg = version

	return cfg
}

func main() {
	ctx := context.Background()

	logger.InitLogger(getLoggingConfig())
	// start the default tracer
	tc, err := tracer.StartDefaultTracer(ctx, getMonitoringConfig())
	if err != nil {
		panic(err)
	}
	defer func() {
		if tc != nil {
			err := tracer.StopTracer(ctx, tc)
			if err != nil {
				logger.Error(ctx, "failed to stop tracer", err)
			}
		}
	}()

	parentCtx := context.Background()

	// myFunc1 with be wrapped with a span.
	// We pass the parentCtx to the function so that the span is created with the parent context.
	myFunc1(parentCtx)

	// myFunc1 with be wrapped with a span.
	// Since we are passing the parentCtx to the function, the span will be created with the parent context.
	// That means is a sibling of the span created in myFunc1.
	myFunc2(parentCtx)
}

func myFunc1(ctx context.Context) {
	fmt.Println()
	fmt.Println("------------------------ Start myFunc1 ------------------------")

	spanCtx, span := tracer.StartSpan(ctx, "myFunc1", trace.SpanKindInternal)
	defer span.End()

	logger.Info(spanCtx, "hello from myFunc1", slog.Any("some_attribute", MyStruct{Name: "go", Age: 15}), slog.String("hello", "is hola in Spanish"))

	fmt.Println()
	// logging an error using a context with a span, the span will be flagged as errored
	logger.Error(spanCtx, "error from myFunc1", errors.New("an error had occurred in myFunc1"))

	fmt.Println()
	// the span (in attributes) will contain the attributes we passed on the above logs
	// and the error (in events) since we logged an error
	fmt.Printf("%+v\n", span.Span)

	fmt.Println("------------------------ End myFunc1 ------------------------")
	fmt.Println()

	// childFunc will be wrapped with a span.
	// Since we are passing the spanCtx to the function, the span will be created with the span context.
	// That means is a child of the span created in myFunc1.
	childFunc(spanCtx)
}

func myFunc2(ctx context.Context) {
	fmt.Println()
	fmt.Println("------------------------ Start myFunc2 ------------------------")

	spanCtx, span := tracer.StartSpan(ctx, "myFunc1", trace.SpanKindInternal)
	defer span.End()

	logger.Info(spanCtx, "hello from myFunc2", slog.Any("response", MyStruct{Name: "java", Age: 28}), slog.String("hello", "is hallo in Dutch"))

	fmt.Println()
	// logging an error using a context with a span, the span will be flagged as errored
	logger.Error(spanCtx, "error from myFunc2", errors.New("an error had occurred in myFunc2"))

	fmt.Println()
	// the span (in attributes) will contain the attributes we passed on the above logs
	// and the error (in events) since we logged an error
	fmt.Printf("%+v\n", span.Span)

	fmt.Println("------------------------ End myFunc2 ------------------------")
	fmt.Println()
}

func childFunc(ctx context.Context) {
	fmt.Println()
	fmt.Println("------------------------ Start childFunc ------------------------")

	_, span := tracer.StartSpan(ctx, "childFunc", trace.SpanKindInternal)
	defer span.End()

	// the span must not contain any more attributes that the default ones (caller information)
	fmt.Printf("%+v\n", span.Span)

	fmt.Println("------------------------ Start childFunc ------------------------")
}
