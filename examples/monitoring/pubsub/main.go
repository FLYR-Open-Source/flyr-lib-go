package main

import (
	"context"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/FlyrInc/flyr-lib-go/logger"
	"github.com/FlyrInc/flyr-lib-go/monitoring/tracer"
	"google.golang.org/api/option"

	pubsubTrace "github.com/FlyrInc/flyr-lib-go/monitoring/pubsub"
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

	projectID := "some-gcp-project-id"
	config := &pubsub.ClientConfig{}                      // this also can be nil
	options := []option.ClientOption{ /* any options */ } // options is optional, therefore they can be omitted

	client, err := pubsubTrace.NewClient(ctx, projectID, config, options...)
	defer func() {
		err := client.Close()
		if err != nil {
			// do something with the error
		}
	}()
}
