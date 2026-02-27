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
	"os"

	"github.com/FLYR-Open-Source/flyr-lib-go/logger"
	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/tracer"
	"go.opentelemetry.io/otel/trace"
)

type MyStruct struct {
	Name string
	Age  int
}

const (
	serviceName = "some-service"
)

// You don't need this part since it's automated in Kubernetes
func init() {
	_ = os.Setenv("OTEL_SERVICE_NAME", serviceName)
	// this is a flag for exporting the traces in stdout
	_ = os.Setenv("OTEL_EXPORTER_OTLP_TEST", "true")
	// set the log level to debug
	_ = os.Setenv("LOG_LEVEL", "debug")
	_ = os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "k8s.container.name={some-container},k8s.deployment.name={some-deployment},k8s.deployment.uid={some-uid},k8s.namespace.name={some-namespace},k8s.node.name={some-node},k8s.pod.name={some-pod},k8s.pod.uid={some-uid},k8s.replicaset.name={some-replicaset},k8s.replicaset.uid={some-uid},service.instance.id={some-namespace}.{some-pod}.{some-container},service.version={some-version}")
}

// run this file to see the output
func main() {
	ctx := context.Background()

	logger.InitLogger()

	// start the default tracer
	err := tracer.StartDefaultTracer(ctx)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = tracer.ShutdownTracerProvider(ctx)
		if err != nil {
			logger.Error(ctx, "failed to stop tracer", err)
		}
	}()

	// we create a parent span that will contain the rest as children
	parentCtx, span := tracer.StartSpan(ctx, "parent", trace.SpanKindInternal)
	defer span.End()

	// myFunc1 with be wrapped with a span.
	// We pass the parentCtx to the function so that the span is created with the parent context.
	myFunc1(parentCtx)

	// myFunc1 with be wrapped with a span.
	// Since we are passing the parentCtx to the function, the span will be created with the parent context.
	// That means is a sibling of the span created in myFunc1.
	myFunc2(parentCtx)
}

func myFunc1(ctx context.Context) {
	spanCtx, span := tracer.StartSpan(ctx, "myFunc1", trace.SpanKindInternal)
	defer span.End()

	logger.Info(spanCtx, "hello from myFunc1", slog.Any("some_attribute", MyStruct{Name: "go", Age: 15}), slog.String("hello", "is hola in Spanish"))

	// debug logs must not have correlation IDs
	logger.Debug(spanCtx, "debug must not have correlation IDs")

	// logging an error using a context with a span, the span will be flagged as errored
	logger.Error(spanCtx, "error from myFunc1", errors.New("an error had occurred in myFunc1"))

	// childFunc will be wrapped with a span.
	// Since we are passing the spanCtx to the function, the span will be created with the span context.
	// That means is a child of the span created in myFunc1.
	childFunc(spanCtx)
}

func myFunc2(ctx context.Context) {
	spanCtx, span := tracer.StartSpan(ctx, "myFunc1", trace.SpanKindInternal)
	defer span.End()

	logger.Info(spanCtx, "hello from myFunc2", slog.Any("response", MyStruct{Name: "java", Age: 28}), slog.String("hello", "is hallo in Dutch"))

	// logging an error using a context with a span, the span will be flagged as errored
	logger.Error(spanCtx, "error from myFunc2", errors.New("an error had occurred in myFunc2"))
}

func childFunc(ctx context.Context) {
	_, span := tracer.StartSpan(ctx, "childFunc", trace.SpanKindInternal)
	defer span.End()
}
