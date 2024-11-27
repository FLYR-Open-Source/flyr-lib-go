package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/FlyrInc/flyr-lib-go/logger"
	"github.com/FlyrInc/flyr-lib-go/monitoring/middleware"
	"github.com/FlyrInc/flyr-lib-go/monitoring/tracer"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	ctx := context.Background()

	logger.InitLogger()

	err := tracer.StartDefaultTracer(ctx)
	if err != nil {
		logger.Error(ctx, "failed to start the tracer", err)
		return
	}
	defer func() {
		tracer.ShutdownTracerProvider(ctx)
	}()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	// Add the Otel middleware to the Gin route
	r.Use(middleware.OtelGinMiddleware())

	r.GET("/ping", Ping)

	r.Run()
}

func Ping(c *gin.Context) {
	reqCtx := c.Request.Context()
	logger.Info(reqCtx, "start!")

	DoSomething(reqCtx)

	FetchFromDB(reqCtx)

	resp := gin.H{
		"message": "pong from Go!",
	}

	logger.Info(
		reqCtx,
		"end!",
		slog.Any("response_body", resp),
		slog.Int64("id", 10),
		slog.String("name", "test"),
		slog.Bool("is_active", true),
		slog.Duration("duration", 10*time.Second),
		slog.Float64("amount", 10.5),
	)

	c.JSON(http.StatusInternalServerError, resp)
}

func DoSomething(ctx context.Context) {
	spanCtx, span := tracer.StartSpan(ctx, "DoSomething", trace.SpanKindInternal)
	defer span.End()
	logger.Error(spanCtx, "some error", errors.New("oops"))
}

func FetchFromDB(ctx context.Context) {
	_, span := tracer.StartSpan(ctx, "FetchFromDB", trace.SpanKindInternal)
	span.End()
}
