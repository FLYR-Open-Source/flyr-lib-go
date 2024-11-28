package main

import (
	"context"
	"os"

	"github.com/FlyrInc/flyr-lib-go/logger"
	"github.com/FlyrInc/flyr-lib-go/monitoring/rabbitmq"
	"github.com/FlyrInc/flyr-lib-go/monitoring/tracer"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	// You can pass the `OBSERVABILITY_SERVICE` environment variable to set the service name
	serviceName = "some-service"
)

// You don't need this part since it's automated in Kubernetes
func init() {
	os.Setenv("OTEL_SERVICE_NAME", serviceName)
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
	/* header */ _ = rabbitmq.InjectAMQPHeaders(spanCtx)
	// pass the headers to the message and publish it
	logger.Debug(spanCtx, "message published")
}

func consumeMessage(ctx context.Context) {
	headers := map[string]interface{}{}
	ctxWithHeaders := rabbitmq.ExtractAMQPHeaders(ctx, headers)
	spanCtx, _ := tracer.StartSpan(ctxWithHeaders, "consumer", oteltrace.SpanKindConsumer)

	logger.Debug(spanCtx, "message consumed")
}
