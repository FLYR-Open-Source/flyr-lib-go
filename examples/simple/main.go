package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/FlyrInc/flyr-lib-go/config"
	"github.com/FlyrInc/flyr-lib-go/logger"
	"github.com/FlyrInc/flyr-lib-go/observability/tracer"
	"go.opentelemetry.io/otel/trace"
)

type Response struct {
	Message string
	Status  int
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
	tc, err := tracer.StartDefaultTracer(ctx, getMonitoringConfig())
	if err != nil {
		panic(err)
	}
	defer func() {
		if tc != nil {
			tracer.StopTracer(ctx, tc)
		}
	}()

	parentCtx := context.Background()
	spanCtx, span := tracer.StartSpan(parentCtx, "DoSomething", trace.SpanKindInternal)

	fmt.Println("------------------------ logger info ------------------------")
	logger.Info(spanCtx, "hello", slog.Any("response", Response{Message: "hello", Status: 200}), slog.String("hola", "geia"))
	fmt.Println("------------------------ logger info ------------------------")

	fmt.Println()
	fmt.Println("------------------------ logger error ------------------------")
	logger.Error(spanCtx, "world", errors.New("an error had occurred"))
	fmt.Println("------------------------ logger error ------------------------")

	fmt.Println()
	fmt.Println("------------------------ first span ------------------------")
	fmt.Printf("%+v\n", span.Span)
	fmt.Println("------------------------ first span ------------------------")

	dosomething(spanCtx)

	span.EndSuccessfully()
}

func dosomething(ctx context.Context) {
	spanCtx, span := tracer.StartSpan(ctx, "Do Something Else", trace.SpanKindInternal)

	fmt.Println()
	fmt.Println("------------------------ dosomething func ------------------------")

	fmt.Println("------------------------ logger info ------------------------")
	logger.Info(spanCtx, "hello from the other func", slog.Any("response", Response{Message: "hello", Status: 404}))
	fmt.Println("------------------------ logger info ------------------------")

	fmt.Println()
	fmt.Println("------------------------ nested span ------------------------")
	fmt.Printf("%+v\n", span.Span)
	fmt.Println("------------------------ nested span ------------------------")

	fmt.Println("------------------------ dosomething func ------------------------")

	span.End()
}
