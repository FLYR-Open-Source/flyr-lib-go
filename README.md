# flyr-lib-go

This library is an internal Go library for Flyr, providing essential utilities for logging and observability, leveraging `slog` for structured logging and OpenTelemetry (OTel) libraries for tracing. This library ensures consistent instrumentation across services by correlating logs and traces, making it easier to troubleshoot and understand application behavior in production.

## Table of Contents

1. [Features](#features)
2. [Installation](#installation)
3. [Usage](#usage)
   - [Logging](#logging)
     - [Environment Variables](#environment-variables)
   - [Monitoring](#monitoring)
     - [Kubernetes Setup](#kubernetes-setup)
       - [Example in Namespace](#example-in-namespace)
       - [Example in a Pod](#example-in-a-pod)
     - [Traces](#traces)
       - [Use of the Tracer](#use-of-the-tracer)
     - [Spans](#spans)
       - [Automatic Correlation](#automatic-correlation)
     - [Distributed Tracing](#distributed-tracing)
       - [HTTP Tracing](#http-tracing)
       - [PubSub Tracing](#pubsub-tracing)
       - [RabbitMQ Tracing](#rabbitmq-tracing)
         - [RabbitMQ Provider](#rabbitmq-provider)
         - [RabbitMQ Consumer](#rabbitmq-consumer)
     - [Middleware](#middleware)
   - [Testing](#testing)
   - [Metrics](#metrics)
4. [Documentation](#documentation)
   - [Internal Documentation](#internal-documentation)
   - [Otel Documentation](#otel-documentation)

## Features

- **Structured Logging**: Provides a unified interface for logging, utilizing `slog` for JSON-based structured logs.
- **Tracing**: Supports distributed tracing through OpenTelemetry, enabling the creation and management of traces and spans.
- **Error Handling**: Flags spans as errored when an error is logged, providing clear visibility into failed operations.
- **Contextual Correlation**: Automatically correlates logs with spans by adding attributes from the log entries to the active span, allowing for streamlined debugging and monitoring.
- **Middlewares**: Provide Tracing middlewares for [gin](https://gin-gonic.com/) and [chi](https://go-chi.io/#/) frameworks.

## Installation

To install `flyr-lib-go`, add it to your module’s dependencies:

```bash
$ export GOPRIVATE=github.com/FlyrInc/flyr-lib-go/*
$ go get github.com/FlyrInc/flyr-lib-go
```

## Usage

### Logging

The logger is an implementation of [log/slog](https://pkg.go.dev/log/slog), a high-performance structured, and leveled logging that is build-in to Go standard library.

The logger is following the [FLYR Logging Standards](https://flyrlabs.atlassian.net/wiki/spaces/CE/pages/4323967442/Logging+Standards).

The structure is by default in JSON format, and the attribute names follow the Otel naming convention for [Services](https://opentelemetry.io/docs/specs/semconv/resource/#service) and [Code](https://opentelemetry.io/docs/specs/semconv/attributes-registry/code/). That ensures correlation on naming between spans and logs.
The [Deployments](https://opentelemetry.io/docs/specs/semconv/resource/deployment-environment/) naming convention as well as more attributes will be injected to the logs from the Otel Collector.
The naming conversions can be found in [internal/config/commontags.go](./internal/config/commontags.go).

You can see examples of how to use the logger in [examples/logging/main.go](./examples/logging/main.go).

Furthermore, it is including in the logs the Trace and Span IDs from an active Span (Learn more in the [Automatic Correlation](#automatic-correlation) section).

#### Environment Variables

The logger accepts a config that reads values from Environment Variables. The below table contains all the supported Environment Variables for the logger:

| Variable Name | Description                                                                         | Default   |
|---------------|-------------------------------------------------------------------------------------|-----------|
| `LOG_LEVEL`   | The log level. The accepted values can be one of (`debug`, `info`, `warn`, `error`) | `info`    |

### Monitoring

For monitoring, the application is using the [otel](https://pkg.go.dev/go.opentelemetry.io/otel) and [otel/sdk](https://pkg.go.dev/go.opentelemetry.io/otel/sdk) packages.

The library enchaches the Otel library by providing more features, but at the same time it exposes all the Otel functionalities for
more advanced uses.

You can see examples of how to use the monitoring package in [examples/monitoring](./examples/monitoring).

#### Kubernetes Setup

The library relies on automation that Otel supports. To make the experience smooth and without extra setup for the teams, there is no need for any environment variables.
However, there is one annotation you need to include.

How have two options, either pass the annotation on the namespace, that will inject the required environment variables to all the pods within the namespace,
or, you can add the annotation directly in the pods you want the required environment variables to be injected.

The annotation is: `instrumentation.opentelemetry.io/inject-sdk: otel-collector/default-instrumentation`.

> [!IMPORTANT]
> Regardless if you include the annotation to the Namespace or the Pod, you **must** make sure the Namespace contains the label `environment`.
> Also, to prevent both the Datadog Agent and the Otel Collector to collect logs, you **must** include the label `ad.datadoghq.com/logs_exclude: "true"` to all the pods.

##### Example in Namespace
```yaml
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    instrumentation.opentelemetry.io/inject-sdk: otel-collector/default-instrumentation
  labels:
    environment: {some-environment}
  name: {some-namespace}
```

> [!IMPORTANT]
> The precedence for the labels is [pod, namespace, node]. That means if in your pod you override the `environment` value (or even if you include an `env` label) the value from the Pod will be selected.

##### Example in a Pod

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {some-deployment-name}
  labels:
    app.kubernetes.io/version: {some-service-version}
    app.kubernetes.io/name: {some-service-name}
    app.kubernetes.io/component: {some-component}
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app.kubernetes.io/version: {some-service-version}
        app.kubernetes.io/name: {some-service-name}
        app.kubernetes.io/component: {some-component}
      annotations:
        instrumentation.opentelemetry.io/inject-sdk: otel-collector/default-instrumentation
        ad.datadoghq.com/logs_exlude: "true"
    spec:
      ...
```

#### Traces

The path of a request through your application.

**Traces** give us the big picture of what happens when a request is made to an application. Whether your application is a monolith with a single database or a sophisticated mesh of services, traces are essential to understanding the full “path” a request takes in your application.

##### Use of the Tracer

The library let's you initialise a Tracer. A Tracer has a name. In our case, the name is the service name (the APM name). In Datadog the Tracer Name is the `operation_name`. Using the default Tracer, you can create Spans.

You can see examples of how to use the default Tracer in [examples/monitoring/simple_example/main.go](./examples/monitoring/simple_example/main.go).

#### Spans

A **Span** represents a unit of work or operation. Spans are the building blocks of Trace. In OpenTelemetry, they include the following information:
- Name
- Parent span ID (empty for root spans)
- Start and End Timestamps
- [Span Context](https://opentelemetry.io/docs/concepts/signals/traces/#span-context)
- [Attributes](https://opentelemetry.io/docs/concepts/signals/traces/#attributes)
- [Span Events](https://opentelemetry.io/docs/concepts/signals/traces/#span-events)
- [Span Links](https://opentelemetry.io/docs/concepts/signals/traces/#span-links)
- [Span Status](https://opentelemetry.io/docs/concepts/signals/traces/#span-status)

Spans can be nested, as is implied by the presence of a parent span ID: child spans represent sub-operations. This allows spans to more accurately capture the work done in an application.

Similarly to the [Logs](#logging), Spans follow the Otel naming convention for [Services](https://opentelemetry.io/docs/specs/semconv/resource/#service), and [Code](https://opentelemetry.io/docs/specs/semconv/attributes-registry/code/). That ensures correlation on naming between spans and logs.
Similarly to the logs, the [Deployments](https://opentelemetry.io/docs/specs/semconv/resource/deployment-environment/) naming convention as well as more attributes will be injected to the spans from the Otel Collector.
The naming conversions can be found in [internal/config/commontags.go](./internal/config/commontags.go).

Creating a Span it requires a Tracer. The library provides a simple way to create Spans. Furthermore, exposing the Span interface you can easily add Attributes, Events, Links, Status and Record Errors in your Spans.

The examples of creating Spans, are the same examples as mentioned before for the Traces (since they are interchangeable with each other).

On top of the Span interface from Otel, the library also exposes two more methods on the Span, `EndWithError` and `EndSuccessfully`.

By default the Span interface from Otel exposes only the method `End` to end the Span.
The new method `EndWithError` is also ending the Span, but it will flag the Span as errored and will also include the error into the Span Events (as it must be based on Otel). On the other hand, `EndSuccessfully` ends the Span and updates the Status as `Ok`.

##### Automatic Correlation

In order for the library to give a better experience when working with Traces, Spans and Logs, it also provides some automations to make the usage easier.

When you create a new Span, it returns back a new `context.Context` value. This new context contains the Span information.

When logging inside a block of code that is wrapped in a Span, it is recommended to pass to the Logger the new context that includes the Span information. That won't only inluce the Trace and Span IDs into the log, but also will automatically include into the Spans the attributes that are being passed to the Logger.

That ensures useful information for debugging will be present at the same time in both the logs and the spans.

> [!IMPORTANT]
> Debug logs do not contain the Trace and Span ID and also do not attach any given metadata to the Span. That is useful because including the debug logs to the spans can flood them with unnecessary information, making them harder to interpret.

> [!WARNING]
> When you add an error log, the span will be flaged as errored and will also include the error into the Span Events (as it must be based on Otel).

#### Distributed Tracing

**Distributed Tracing** provides visibility into the flow of requests as they traverse through various services in a distributed system. This capability is essential for understanding the interactions between microservices, identifying performance bottlenecks, and diagnosing issues. Each trace captures the entire lifecycle of a request, recording its journey across different services and components. By analyzing traces, you can visualize how requests propagate through your system and identify areas for optimization. This library supports distributed tracing by leveraging OpenTelemetry, ensuring consistent trace propagation and context management across different services and frameworks.

The below sections provide examples to enable Distributed Tracing for various different communication protocols.

##### HTTP Tracing

**HTTP Tracing** captures and monitors HTTP requests and responses, allowing you to trace the path of an HTTP call through various services, measure latency, and identify potential issues in request handling.

The library provides a very simple way to enable Distributed Tracing for HTTP requests.
The package [monitoring/http](./monitoring/http/) exposes two different functions.

The `NewHttpClient()` can be used to create a new `http.Client` with Distributed Tracing enabled.
If you have a client already, you can use the function `SetHttpTransport(client http.Client)` that enables Distributed Tracing for the given client. The same client is returned back.

Create a new client using the library:
```go
package main

import "github.com/FlyrInc/flyr-lib-go/http"

func main() {
  client := http.NewHttpClient()

  resp, err := client.Get("https://example.com")
  if err != nil {
    ... // handle error
  }

  defer resp.Body.Close()
  ...
}
```

Enable distributed tracing to an existing `http.Client`:
```go
package main

import "github.com/FlyrInc/flyr-lib-go/http"

func main() {
  client := http.Client{}
  ... // add any setup to client

  client = SetHttpTransport(client)

  resp, err := client.Get("https://example.com")
  if err != nil {
    ...
  }

  defer resp.Body.Close()
  ...
}
```

##### PubSub Tracing

The **PubSub Tracing** enables tracking and monitoring of messages as they are published and consumed within a Pub/Sub system, providing visibility into the lifecycle and performance of message-driven workflows.

The library does not expose any functions that can be used, because the ["cloud.google.com/go/pubsub"]("cloud.google.com/go/pubsub") module already supports Open Telemetry!

Link: [https://cloud.google.com/pubsub/docs/open-telemetry-tracing](https://cloud.google.com/pubsub/docs/open-telemetry-tracing)

All you need to do is to enable it on the client - both for publishing and consuming messages.

```go
client, err := pubsub.NewClientWithConfig(ctx, projectID, &pubsub.ClientConfig{
        EnableOpenTelemetryTracing: true,
})
...
```

##### RabbitMQ Tracing

The **RabbitMQ Tracing** provides visibility into message flows by capturing spans for both message production and consumption, enabling detailed insights into the performance and behavior of RabbitMQ-based communication.

###### RabbitMQ Provider
TBD

###### RabbitMQ Consumer
TBD

#### Middleware

The library provides middleware for both the Gin and Chi frameworks in Go, responsible for creating the main span for incoming requests to endpoints, ensuring that each HTTP request is traced and correlated with the overall distributed trace.

```go
  package main

  import "github.com/FlyrInc/flyr-lib-go/middleware"

  func main() {
    ... setup chi or gin

    // for gin-gonic
    engine.Use(middleware.OtelGinMiddleware())
    // for chi
    engine.Use(middleware.OtelChiMiddleware())
  }
```

#### Testing

The library also exposes the [pkg/testhelpers/](./pkg/testhelpers/), which contains mock implementations for a Traces and a Span. You can use it to write tests in your codebase.

### Metrics
TBD

### Documentation

#### Internal Documentation

In order to access the internal documentation, you can run `make docs` and then visit `http://localhost:6060/pkg/github.com/FlyrInc/flyr-lib-go/`.

#### Otel Documentation

- [Otel Documentation](https://opentelemetry.io/docs/)
- [Traces](https://opentelemetry.io/docs/concepts/signals/traces/)
- [Observability primer](https://opentelemetry.io/docs/concepts/observability-primer/)
- [Go](https://opentelemetry.io/docs/languages/go/)
  - [Getting Started](https://opentelemetry.io/docs/languages/go/getting-started/)
