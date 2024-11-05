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

func main() {
	ctx := context.Background()

	monitoringCfg := config.NewMonitoringConfig()
	monitoringCfg.EnableTracer = true
	monitoringCfg.ServiceCfg = "some-service"
	monitoringCfg.EnvCfg = "dev"
	monitoringCfg.FlyrTenantCfg = "fl"
	monitoringCfg.VersionCfg = "v1.0.0"

	loggerCfg := config.NewLoggerConfig()
	loggerCfg.ServiceCfg = "some-service"
	loggerCfg.EnvCfg = "dev"
	loggerCfg.FlyrTenantCfg = "fl"
	loggerCfg.VersionCfg = "v1.0.0"

	logger.InitLogger(loggerCfg)
	tc, err := tracer.StartDefaultTracer(ctx, monitoringCfg)
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

	logger.Info(spanCtx, "hello", slog.Any("response", Response{Message: "hello", Status: 200}), slog.String("hola", "geia"))
	logger.Error(spanCtx, "world", errors.New("an error had occurred"))

	fmt.Printf("%+v\n", span.Span)

	span.EndSuccessfully()
}
