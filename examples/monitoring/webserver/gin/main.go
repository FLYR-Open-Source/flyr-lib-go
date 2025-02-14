// MIT License
//
// Copyright (c) 2025 FLYR, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
