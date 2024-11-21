package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/FlyrInc/flyr-lib-go/config"
	"github.com/FlyrInc/flyr-lib-go/logger"
)

const (
	// You can pass the `OBSERVABILITY_SERVICE` environment variable to set the service name
	serviceName = "some-service"
	// You can pass the `OBSERVABILITY_ENV` environment variable to set the environment
	env = "dev"
	// You can pass the `OBSERVABILITY_FLYR_TENANT` environment variable to set the tenant
	flyrTenant = "fl"
	// You can pass the `OBSERVABILITY_VERSION` environment variable to set the version
	version = "v1.0.0"
	// You can pass the `LOG_LEVEL` environment variable to set the log level
	logLevel = "debug"
)

func getLoggingConfig() config.LoggerConfig {
	cfg := config.NewLoggerConfig()
	cfg.ServiceCfg = serviceName
	cfg.EnvCfg = env
	cfg.FlyrTenantCfg = flyrTenant
	cfg.VersionCfg = version
	cfg.LogLevelCfg = logLevel

	return cfg
}

func main() {
	ctx := context.Background()

	// Initialize the logger
	logger.InitLogger(getLoggingConfig())

	logger.Info(ctx, "This is an info message")
	// Output:
	// {
	// 	"time": "2024-11-07T17:53:20.108167Z",
	// 	"level": "INFO",
	// 	"message": "This is an info message",
	// 	"deployment.environment.name": "dev",
	// 	"service.version": "v1.0.0",
	// 	"service.name": "some-service",
	//  "env": "dev",
	// 	"flyr_tenant": "fl",
	// 	"code.filepath": ".../flyr-lib-go/examples/logging/main.go",
	// 	"code.lineno": 43,
	// 	"code.function": "main",
	// 	"code.namespace": "main"
	// }
	fmt.Println()

	logger.Debug(ctx, "This is a debug message")
	// Output:
	// {
	// 	"time": "2024-11-07T17:55:05.213138Z",
	// 	"level": "DEBUG",
	// 	"message": "This is a debug message",
	// 	"deployment.environment.name": "dev",
	// 	"service.version": "v1.0.0",
	// 	"service.name": "some-service",
	//  "env": "dev",
	// 	"flyr_tenant": "fl",
	// 	"code.filepath": ".../flyr-lib-go/examples/logging/main.go",
	// 	"code.lineno": 61,
	// 	"code.function": "main",
	// 	"code.namespace": "main"
	// }
	fmt.Println()

	logger.Warn(ctx, "This is a warning message")
	// Output:
	// {
	// 	"time": "2024-11-07T17:55:32.894155Z",
	// 	"level": "WARN",
	// 	"message": "This is a warning message",
	// 	"deployment.environment.name": "dev",
	// 	"service.version": "v1.0.0",
	// 	"service.name": "some-service",
	// 	"flyr_tenant": "fl",
	//  "env": "dev",
	// 	"code.filepath": ".../flyr-lib-go/examples/logging/main.go",
	// 	"code.lineno": 79,
	// 	"code.function": "main",
	// 	"code.namespace": "main"
	// }
	fmt.Println()

	logger.Error(ctx, "This is an error message", errors.New("this is an error"))
	// Output:
	// {
	// 	"time": "2024-11-07T17:55:48.871878Z",
	// 	"level": "ERROR",
	// 	"message": "This is an error message",
	// 	"deployment.environment.name": "dev",
	// 	"service.version": "v1.0.0",
	// 	"service.name": "some-service",
	// 	"flyr_tenant": "fl",
	//  "env": "dev",
	// 	"code.filepath": ".../flyr-lib-go/examples/logging/main.go",
	// 	"code.lineno": 97,
	// 	"code.function": "main",
	// 	"code.namespace": "main",
	// 	"error": "this is an error"
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
	// 	"deployment.environment.name": "dev",
	// 	"service.version": "v1.0.0",
	// 	"service.name": "some-service",
	//  "env": "dev",
	// 	"flyr_tenant": "fl",
	// 	"code.filepath": ".../flyr-lib-go/examples/logging/main.go",
	// 	"code.lineno": 117,
	// 	"code.function": "main",
	// 	"code.namespace": "main",
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
	// 	"deployment.environment.name": "dev",
	// 	"service.version": "v1.0.0",
	// 	"service.name": "some-service",
	//  "env": "dev",
	// 	"flyr_tenant": "fl",
	// 	"code.filepath": ".../flyr-lib-go/examples/logging/main.go",
	// 	"code.lineno": 147,
	// 	"code.function": "main",
	// 	"code.namespace": "main",
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
