# flyr-lib-go

This library is an internal Go library for Flyr, providing essential utilities for logging and observability, leveraging `slog` for structured logging and OpenTelemetry (OTel) libraries for tracing. This library ensures consistent instrumentation across services by correlating logs and traces, making it easier to troubleshoot and understand application behavior in production.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Logging](#logging)
    - [Environment Variables](#environment-variables)
  - [Monitoring](#monitoring)
    - [Environment Variables](#environment-variables-1)
    - [Traces](#traces)
      - [Use the default Tracer](#use-the-default-tracer)
      - [Use your own Tracers](#use-your-own-tracers)
    - [Spans](#spans)
      - [Automatic Correlation](#automatic-correlation)
    - [Middleware](#middleware)
    - [Testing](#testing)
    - [Metrics](#metrics)
  - [Documentation](#documentation)
    - [Internal Documentation](#internal-documentation)
    - [Otel Documentation](#otel-documentation)

## Features

- **Structured Logging**: Provides a unified interface for logging, utilizing `slog` for JSON-based structured logs.
- **Tracing**: Supports distributed tracing through OpenTelemetry, enabling the creation and management of traces and spans.
- **Error Handling**: Flags spans as errored when an error is logged, providing clear visibility into failed operations.
- **Contextual Correlation**: Automatically correlates logs with spans by adding attributes from the log entries to the active span, allowing for streamlined debugging and monitoring.

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

The structure is by default in JSON format, and the attribute names follow the Otel naming convention for [Services](https://opentelemetry.io/docs/specs/semconv/resource/#service), [Deployments](https://opentelemetry.io/docs/specs/semconv/resource/deployment-environment/) and [Code](https://opentelemetry.io/docs/specs/semconv/attributes-registry/code/). That ensures correlation on naming between spans and logs.
The naming conversions can be found in [config/commontags.go](./config/commontags.go).

You can see examples of how to use the logger in [examples/logging/main.go](./examples/logging/main.go).

Furthermore, it is including in the logs the Trace and Span IDs from an active Span (See example in the [monitoring ](#monitoring) section).

#### Environment Variables

The logger accepts a config in which you can pass:

| Variable Name               | Description                                  | Default   |
|-----------------------------|----------------------------------------------|-----------|
| `OBSERVABILITY_SERVICE`     | The name of the service                      | empty     |
| `OBSERVABILITY_VERSION`     | The version of the service                   | empty     |
| `OBSERVABILITY_ENV`         | The environment name                         | empty     |
| `OBSERVABILITY_FLYR_TENANT` | The tenant name                              | empty     |
| `LOG_LEVEL`                 | The log level                                | `info`    |

### Monitoring

For monitoring, the application is using the [otel](https://pkg.go.dev/go.opentelemetry.io/otel) and [otel/sdk](https://pkg.go.dev/go.opentelemetry.io/otel/sdk) packages.

The library enchaches the Otel library by providing more features, but at the same time it exposes all the Otel functionalities for
more advanced uses.

You can see examples of how to use the monitoring package in [examples/monitoring](./examples/monitoring).

#### Environment Variables

The logger accepts a config in which you can pass:

| Variable Name                          | Description                                                   | Default   |
|----------------------------------------|---------------------------------------------------------------|-----------|
| `OBSERVABILITY_SERVICE`                | The name of the service                                       | empty     |
| `OBSERVABILITY_VERSION`                | The version of the service                                    | empty     |
| `OBSERVABILITY_ENV`                    | The environment name                                          | empty     |
| `OBSERVABILITY_FLYR_TENANT`            | The tenant name                                               | empty     |
| `OBSERVABILITY_TRACER_ENABLED`         | The log level                                                 | `false`   |
| `OBSERVABILITY_EXPORTER_OTLP_ENDPOINT` | A base endpoint URL for any signal type, with the port number | empty     |

#### Traces

The path of a request through your application.

**Traces** give us the big picture of what happens when a request is made to an application. Whether your application is a monolith with a single database or a sophisticated mesh of services, traces are essential to understanding the full “path” a request takes in your application.

The library exposes two ways of using Traces:
- Use the default Tracer that is initialised inside the library
- Use your own Tracers

##### Use the default Tracer

The default way to use the library, is by using the default Tracer. A Tracer has a name. In our case, the name is the service name (`OBSERVABILITY_SERVICE`). In Datadog the Tracer Name is the `operation_name`. Using the default Tracer, you can create Spans.

You can see examples of how to use the default Tracer in [examples/monitoring/default_tracer/main.go](./examples/monitoring/default_tracer/main.go).

##### Use your own Tracers

A more advance usage is to create your own Tracers. As mentioned above, the default Tracer's name is the service name (which is the `operation_name` in Datadog). You can group your operations by creating custom Tracers with custom names.

For example, you can have a Tracer per layer. That means a Tracer for your Service (Domain) layer, another one for your Repository layer and so on.

You can see examples of how to use custom Tracers in [examples/monitoring/custom_tracers/main.go](./examples/monitoring/custom_tracers/main.go).

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

Similarly to the [Logs](#logging), Spans follow the Otel naming convention for [Services](https://opentelemetry.io/docs/specs/semconv/resource/#service), [Deployments](https://opentelemetry.io/docs/specs/semconv/resource/deployment-environment/) and [Code](https://opentelemetry.io/docs/specs/semconv/attributes-registry/code/). That ensures correlation on naming between spans and logs.
The naming conversions can be found in [config/commontags.go](./config/commontags.go).

Creating a Span it requires a Tracer. The library provides a simple way to create Spans. Furthermore, exposing the Span interface you can easily add Attributes, Events, Links, Status and Record Errors in your Spans.

The examples of creating Spans, are the same examples as mentioned before for the Traces (since they are interchangeable with each other).

On top of the Span interface from Otel, the library also exposes two more methods on the Span, `EndWithError` and `EndSuccessfully`.

By default the Span interface from Otel exposes only the method `End` to end the Span.
The new method `EndWithError` is also ending the Span, but it will flag the Span as errored and will also include the error into the Span Events (as it must be based on Otel). On the other hand, `EndSuccessfully` ends the Span and updates the Status as `Ok`.

##### Automatic Correlation

In order for the library to give a better experience when working with Traces, Spans and Logs, it also provides some automations to make the usage easier.

When you create a new Span, it returns back a new `context.Context` value. This new context contains the Span information.

When logging inside a block of code that is wrapped in a Span, it is recommended to pass to the Logger the new context that was includes the Span. That won't only inluce the Trace and Span IDs into the log, but also will automatically include into the Spans the attributes that are being passed to the Logger.

That ensures useful information for debugging will be present at the same time in both the logs and the spans.

#### Middleware
TBD

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
