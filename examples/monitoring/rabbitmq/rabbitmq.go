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
	"log/slog"
	"os"

	"github.com/FLYR-Open-Source/flyr-lib-go/logger"
	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/rabbitmq"
	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/tracer"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	serviceName = "some-service"
)

// You don't need this part since it's automated in Kubernetes
func init() {
	os.Setenv("OTEL_SERVICE_NAME", serviceName)
	// this is a flag for exporting the traces in stdout
	os.Setenv("OTEL_EXPORTER_OTLP_TEST", "true")
	// set the log level to debug
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "k8s.container.name={some-container},k8s.deployment.name={some-deployment},k8s.deployment.uid={some-uid},k8s.namespace.name={some-namespace},k8s.node.name={some-node},k8s.pod.name={some-pod},k8s.pod.uid={some-uid},k8s.replicaset.name={some-replicaset},k8s.replicaset.uid={some-uid},service.instance.id={some-namespace}.{some-pod}.{some-container},service.version={some-version}")
}

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

	publishMessage(ctx)
	consumeMessage(ctx)
}

func publishMessage(ctx context.Context) {
	spanCtx, span := tracer.StartSpan(ctx, "publisher", oteltrace.SpanKindProducer)
	defer span.End()
	span.SetAttributes(attribute.String("queue", "some-queue"))
	span.SetAttributes(attribute.String("exchange", "some-exchange"))
	headers := rabbitmq.InjectAMQPHeaders(spanCtx)
	// pass the headers to the message and publish it
	logger.Debug(spanCtx, "message published", slog.Any("headers", headers))
}

func consumeMessage(ctx context.Context) {
	headers := map[string]interface{}{}
	ctxWithHeaders := rabbitmq.ExtractAMQPHeaders(ctx, headers)
	spanCtx, span := tracer.StartSpan(ctxWithHeaders, "consumer", oteltrace.SpanKindConsumer)
	defer span.End()

	logger.Debug(spanCtx, "message consumed", slog.Any("headers", headers))
}
