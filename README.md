# Go Observability Library

This library provides common utilities for logging, tracing, and metrics using OpenTelemetry. It allows services to easily integrate observability features such as distributed tracing, structured logging, and metrics collection. The library includes three main components:

- **logs**: A utility for structured logging with trace context.
- **metrics**: A utility for metrics collection and exporting.
- **trace**: A utility for setting up and using distributed tracing.

## Installation

To use this library, you need to install the OpenTelemetry Go SDK and related dependencies. You can install these with Go modules:

```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
go get go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc
```

Make sure your `go.mod` file includes the following:

```bash
go get FlyrInc/flyr-lib-go/common-library/logs
go get FlyrInc/flyr-lib-go/common-library/metrics
go get FlyrInc/flyr-lib-go/common-library/trace
```

## Usage

### 1. Logging with Trace Context (`logs`)

This module provides structured logging with trace information such as `trace_id` and `span_id`. These trace details will be automatically included in your log messages if there’s an active span.

#### Example:

```go
package main

import (
    "FlyrInc/flyr-lib-go/common-library/logs"
)

func main() {
    logger := logs.InitLogger("my-service")
    
    // Log an informational message
    logs.Log(logger, "INFO", "User login succeeded", map[string]interface{}{
        "user_id": "123",
    })

    // Log an error message
    logs.Log(logger, "ERROR", "Failed to connect to database", map[string]interface{}{
        "error_code": "DB_CONN_ERR",
    })
}
```

In this example, the `Log()` function logs messages at different levels (INFO, ERROR) with additional context data like `user_id` or `error_code`. If there is an active span, the trace ID and span ID will be included in the log output.

### 2. Metrics Collection (`metrics`)

This module provides a way to initialize and record metrics. The metrics can be exported to an OpenTelemetry collector.

#### Example:

```go
package main

import (
    "FlyrInc/flyr-lib-go/common-library/metrics"
)

func main() {
    // Initialize metrics exporter (send metrics to OpenTelemetry collector)
    meter := metrics.InitMetrics("my-service", "http://localhost:4317")

    // Record a custom metric (e.g., request count)
    metrics.RecordMetric(meter, "http_requests_total", 1.0)
}
```

In this example, we initialize the metrics system with `InitMetrics()`. You can specify an OTLP endpoint (e.g., `http://localhost:4317`) to send metrics to an OpenTelemetry collector.

### 3. Distributed Tracing (`trace`)

This module provides a way to set up distributed tracing using OpenTelemetry. Spans are created and closed to trace execution across services.

#### Example:

```go
package main

import (
    "context"
    "FlyrInc/flyr-lib-go/common-library/trace"
)

func main() {
    tracer := trace.InitTracer("my-service", "http://localhost:4317")

    ctx, span := trace.StartSpan(context.Background(), tracer, "database-query")
    defer trace.EndSpan(span)

    // Simulate a database query operation
    println("Executing database query...")
}
```

In this example, a trace is initiated for a specific service. The `StartSpan()` function starts a new span and tracks the execution of a code block.

## Full Example: Using Logging, Tracing, and Metrics Together

Here’s a full example of how you can use logging, tracing, and metrics in a Go service:

```go
package main

import (
    "context"
    "FlyrInc/flyr-lib-go/common-library/logs"
    "FlyrInc/flyr-lib-go/common-library/metrics"
    "FlyrInc/flyr-lib-go/common-library/trace"
)

func main() {
    // Initialize logging, tracing, and metrics
    logger := logs.InitLogger("my-service")
    tracer := trace.InitTracer("my-service", "http://localhost:4317")
    meter := metrics.InitMetrics("my-service", "http://localhost:4317")

    // Example of logging, tracing, and recording metrics in a single flow
    ctx, span := trace.StartSpan(context.Background(), tracer, "http-request")
    defer trace.EndSpan(span)

    // Log a message with trace context
    logs.Log(logger, "INFO", "Received user request", map[string]interface{}{
        "user_id": "123",
    })

    // Simulate processing the request
    println("Processing request...")

    // Record a metric for request processing
    metrics.RecordMetric(meter, "http_requests_total", 1.0)

    // Log completion of request
    logs.Log(logger, "INFO", "User request processed successfully", map[string]interface{}{
        "user_id": "123",
    })
}
```

### 4. Configuration

For production, it’s important to set up an OpenTelemetry Collector and configure the exporter endpoints appropriately. You can set the `endpoint` parameter to the URL of your OpenTelemetry collector. For local development, you can leave the `endpoint` parameter as `localhost:4317` if you are running the collector locally.

### Requirements

- Go 1.15+
- OpenTelemetry Go libraries

### Key Features

- **Structured Logging**: Log messages are structured and enriched with trace context.
- **Distributed Tracing**: Easily track requests across services with distributed tracing.
- **Metrics Collection**: Collect and export custom metrics for monitoring service performance.



