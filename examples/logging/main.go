package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/FlyrInc/flyr-lib-go/logger"
)

const (
	// You can pass the `OTEL_SERVICE_NAME` environment variable to set the service name
	serviceName = "some-service"
	// You can pass the `LOG_LEVEL` environment variable to set the log level
	logLevel = "debug"
)

func init() {
	os.Setenv("OTEL_SERVICE_NAME", serviceName)
	os.Setenv("LOG_LEVEL", logLevel)
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "k8s.container.name={some-container},k8s.deployment.name={some-deployment},k8s.deployment.uid={some-uid},k8s.namespace.name={some-namespace},k8s.node.name={some-node},k8s.pod.name={some-pod},k8s.pod.uid={some-uid},k8s.replicaset.name={some-replicaset},k8s.replicaset.uid={some-uid},service.instance.id={some-namespace}.{some-pod}.{some-container},service.version={some-version}")
}

func main() {
	ctx := context.Background()

	// Initialize the logger
	logger.InitLogger()

	logger.Info(ctx, "This is an info message")
	// Output:
	// {
	// 	"time": "2024-11-07T17:53:20.108167Z",
	//  "level": "INFO",
	//  "message": "This is an info message",
	//  "service.name": "some-service",
	//  "service.instance.id": "{some-namespace}.{some-pod}.{some-container}",
	//  "service.version": "{some-version}",
	//  "code.filepath": ".../flyr-lib-go/examples/logging/main.go",
	//  "code.lineno": ...,
	//  "code.function": "main",
	//  "code.namespace": "main"
	// }
	fmt.Println()

	logger.Debug(ctx, "This is a debug message")
	// Output:
	// {
	// 	"time": "2024-11-07T17:55:05.213138Z",
	// 	"level": "DEBUG",
	// 	"message": "This is a debug message",
	//  "service.name": "some-service",
	//  "service.instance.id": "{some-namespace}.{some-pod}.{some-container}",
	//  "service.version": "{some-version}",
	//  "code.filepath": ".../flyr-lib-go/examples/logging/main.go",
	//  "code.lineno": ...,
	//  "code.function": "main",
	//  "code.namespace": "main"
	// }
	fmt.Println()

	logger.Warn(ctx, "This is a warning message")
	// Output:
	// {
	// 	"time": "2024-11-07T17:55:32.894155Z",
	// 	"level": "WARN",
	// 	"message": "This is a warning message",
	//  "service.name": "some-service",
	//  "service.instance.id": "{some-namespace}.{some-pod}.{some-container}",
	//  "service.version": "{some-version}",
	//  "code.filepath": ".../flyr-lib-go/examples/logging/main.go",
	//  "code.lineno": ...,
	//  "code.function": "main",
	//  "code.namespace": "main"
	// }
	fmt.Println()

	logger.Error(ctx, "This is an error message", errors.New("this is an error"))
	// Output:
	// {
	// 	"time": "2024-11-07T17:55:48.871878Z",
	// 	"level": "ERROR",
	// 	"message": "This is an error message",
	//  "service.name": "some-service",
	//  "service.instance.id": "{some-namespace}.{some-pod}.{some-container}",
	//  "service.version": "{some-version}",
	//  "code.filepath": ".../flyr-lib-go/examples/logging/main.go",
	//  "code.lineno": ...,
	//  "code.function": "main",
	//  "code.namespace": "main"
	// }
	fmt.Println()

	// add metadata to the log
	logger.Info(
		ctx,
		"This is an info message with metadata",
		slog.String("someKey", "someValue"),
	)
	// Output:
	// {
	// 	"time": "2024-11-07T17:56:36.935963Z",
	// 	"level": "INFO",
	// 	"message": "This is an info message with metadata",
	//  "service.name": "some-service",
	//  "service.instance.id": "{some-namespace}.{some-pod}.{some-container}",
	//  "service.version": "{some-version}",
	//  "code.filepath": ".../flyr-lib-go/examples/logging/main.go",
	//  "code.lineno": ...,
	//  "code.function": "main",
	//  "code.namespace": "main"
	// 	"metadata": {
	// 		"someKey": "someValue"
	// 	}
	// }
	fmt.Println()

	type testStruct struct {
		SomeKey string
		SomeInt int
	}
	// add multiple metadata to the log
	logger.Info(
		ctx,
		"This is an info message with multiple metadata",
		slog.String("someKey", "someValue"),
		slog.Int("someInt", 1),
		slog.Any("someStruct", testStruct{SomeKey: "someKey", SomeInt: 100}),
	)
	// Output:
	// {
	// 	"time": "2024-11-07T17:59:25.607843Z",
	// 	"level": "INFO",
	// 	"message": "This is an info message with multiple metadata",
	//  "service.name": "some-service",
	//  "service.instance.id": "{some-namespace}.{some-pod}.{some-container}",
	//  "service.version": "{some-version}",
	//  "code.filepath": ".../flyr-lib-go/examples/logging/main.go",
	//  "code.lineno": ...,
	//  "code.function": "main",
	//  "code.namespace": "main"
	// 	"metadata": {
	// 		"someKey": "someValue",
	// 		"someInt": 1,
	// 		"someStruct": {
	// 			"SomeKey": "someKey",
	// 			"SomeInt": 100
	// 		}
	// 	}
	// }
	fmt.Println()
}
