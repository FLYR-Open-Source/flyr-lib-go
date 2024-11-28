package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/FlyrInc/flyr-lib-go/logger"
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
)

// You don't need this part since it's automated in Kubernetes
func init() {
	os.Setenv("OTEL_SERVICE_NAME", serviceName)
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "k8s.container.name={some-container},k8s.deployment.name={some-deployment},k8s.deployment.uid={some-uid},k8s.namespace.name={some-namespace},k8s.node.name={some-node},k8s.pod.name={some-pod},k8s.pod.uid={some-uid},k8s.replicaset.name={some-replicaset},k8s.replicaset.uid={some-uid},service.instance.id={some-namespace}.{some-pod}.{some-container},service.version={some-version}")
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

	fmt.Println()
	fmt.Println("------------------------ Parent Span ------------------------")
	fmt.Printf("%+v\n", span.Span)
	fmt.Println("------------------------ End Parent Span ------------------------")
	fmt.Println()

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
	// debug logs must not have correlation IDs
	logger.Debug(spanCtx, "debug must not have correlation IDs")

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
