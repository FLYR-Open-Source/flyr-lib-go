# Monitoring

The monitoring package is built on [OpenTelemetry](https://opentelemetry.io/) and provides implementations for creating custom spans, sending custom metrics, and propagating Trace IDs across HTTP requests, GCP Pub/Sub clients, and RabbitMQ clients.

1. [Traces](#traces)
    - [Use of the Tracer](#use-of-the-tracer)
    - [Distributed Tracing](#distributed-tracing)
2. [Spans](#spans)
    - [Automatic Correlation](#automatic-correlation)
3. [Trace Propagation](#trace-proagation)
    - [HTTP Tracing](#http-tracing)
    - [PubSub Tracing](#pubsub-tracing)
    - [RabbitMQ Tracing](#rabbitmq-tracing)
    - [Middleware](#middleware)
4. [Metrics](#metrics)
    - [Use of the Metrics](#use-of-the-metrics)
5. [Kubernetes Setup](#kubernetes-setup)
6. [Examples](#examples)
7. [Otel Documentation](#otel-documentation)

## Traces

The path of a request through your application.

**Traces** give us the big picture of what happens when a request is made to an application. Whether your application is a monolith with a single database or a sophisticated mesh of services, traces are essential to understanding the full “path” a request takes in your application.

### Distributed Tracing

**Distributed Tracing** provides visibility into the flow of requests as they traverse through various services in a distributed system. This capability is essential for understanding the interactions between microservices, identifying performance bottlenecks, and diagnosing issues. Each trace captures the entire lifecycle of a request, recording its journey across different services and components. By analyzing traces, you can visualize how requests propagate through your system and identify areas for optimization. This library supports distributed tracing by leveraging OpenTelemetry, ensuring consistent trace propagation and context management across different services and frameworks.

### Use of the Tracer

The library let's you initialise a Tracer. A Tracer has a name. In our case, the name is the service name (the APM name). Using the default Tracer, you can create Spans.

Before you initialise the Tracer, first see the [Kubernetes Setup](#kubernetes-setup).
Also, you can find examples of the Traces in the [examples](#examples).

## Spans

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

Spans follow the [Open Telemetry Semantic Conventions](https://opentelemetry.io/docs/concepts/semantic-conventions/). That ensures correlation on naming between spans and [logs](../logger/).

The naming conversions can be found in [internal/config/commontags.go](../internal/config/commontags.go).

Creating a Span requires a Tracer. The library provides a simple way to create Spans. Furthermore, the exposed Span interface allows you to easily add a Status, Attributes, Events, Links to a Span and Record Errors in a Span.

On top of the Span interface from Otel, the library also exposes two more methods on the Span, `EndWithError` and `EndSuccessfully`.

By default the Span interface from Otel exposes only the method `End` to end the Span.
The new method `EndWithError` is also ending the Span, but it will flag the Span as errored and will also include the error into the Span Events (as it must be based on Otel). On the other hand, `EndSuccessfully` ends the Span and updates the Status as `Ok`.

> [!IMPORTANT]
> You must **always** close the Spans, otherwise you might experience OOM-kills in your services. One way to always ensure the spans are closing, you can use the [go-spancheck](https://github.com/jjti/go-spancheck) rule in the [golangci-lint](https://github.com/golangci/golangci-lint).

### Automatic Correlation

In order for the library to give a better experience when working with Traces, Spans and Logs, it also provides some automations to make the usage easier.

When you create a new Span, it returns back a new `context.Context` value. This new context contains the Span information.

When logging inside a block of code that is wrapped in a Span, it is recommended to pass to the Logger the new context that includes the Span information. That won't only inluce the Trace and Span IDs into the log, but also will automatically include into the Spans the attributes that are being passed to the Logger.

That ensures useful information for debugging will be present in both the logs and the spans.

> [!IMPORTANT]
> Debug logs do not contain the Trace and Span ID and also do not attach any given metadata to the Span. That is useful because including the debug logs to the spans can flood them with unnecessary information, making them harder to interpret.

> [!WARNING]
> When you add an error log, the span will be flaged as errored and will also include the error into the Span Events (as it must be based on Otel).

## Trace Propagation

For tracing to be useful, it is essential that a trace is propagated between the different components and services of a system. The following section describe how to propagate a trace using various communication protocols.

### HTTP Tracing

**HTTP Tracing** captures and monitors HTTP requests and responses, allowing you to trace the path of an HTTP call through various services, measure latency, and identify potential issues in request handling.

The library provides a very simple way to enable Distributed Tracing for HTTP requests.
The package [monitoring/http](./http/) exposes two different functions.

The `NewHttpClient()` can be used to create a new `http.Client` with Distributed Tracing enabled.
If you have a client already, you can use the function `SetHttpTransport(client http.Client)` that enables Distributed Tracing for the given client. The same client is returned back.

Also, you can find examples: [examples](#examples).

### PubSub Tracing

The **PubSub Tracing** enables tracking and monitoring of messages as they are published and consumed within a Pub/Sub system, providing visibility into the lifecycle and performance of message-driven workflows.

The library does not expose any functions that can be used, because the ["cloud.google.com/go/pubsub"]("cloud.google.com/go/pubsub") module already supports Open Telemetry!

Link: [https://cloud.google.com/pubsub/docs/open-telemetry-tracing](https://cloud.google.com/pubsub/docs/open-telemetry-tracing)

All you need to do is to enable it on the client - both for publishing and consuming messages.

Also, you can find examples: [examples](#examples).

### RabbitMQ Tracing

The **RabbitMQ Tracing** provides visibility into message flows by capturing spans for both message production and consumption, enabling detailed insights into the performance and behavior of RabbitMQ-based communication.

Also, you can find examples: [examples](#examples).

### Middleware

The library provides middleware for both the Gin and Chi frameworks in Go, responsible for creating the main span for incoming requests to endpoints, ensuring that each HTTP request is traced and correlated with the overall distributed trace.

Also, you can find examples: [examples](#examples).

## Metrics

A measurement captured at runtime.

A **Metric** is a measurement of a service captured at runtime. The moment of capturing a measurements is known as a metric event, which consists not only of the measurement itself, but also the time at which it was captured and associated metadata.

Application and request metrics are important indicators of availability and performance. Custom metrics can provide insights into how availability indicators impact user experience or the business. Collected data can be used to alert of an outage or trigger scheduling decisions to scale up a deployment automatically upon high demand.

Before you initialise the Meter, first see the [Kubernetes Setup](#kubernetes-setup).
Also, you can find examples of the Metrics in the [examples](#examples).

### Use of the Metrics

The library let's you initialise a Meter. This initialised Meter will be used internally by the library to help you send custom metrics. This library provides wrappers for all the available Metric Types. This is done because we want to "force" the developers to always include description and Unit on the custom metrics, so they are self explanatory.

## Environment Variables

The monitoring package accepts a config that reads values from Environment Variables. The below table contains all the supported Environment Variables:

| Variable Name                        | Description                                                                                                            |
|--------------------------------------|------------------------------------------------------------------------------------------------------------------------|
| `OTEL_SERVICE_NAME`                  | The name of the service. This value normally is the same as the value in the Kubernetes label `app.kubernetes.io/name` |
| `OTEL_EXPORTER_OTLP_PROTOCOL`        | Specifies the OTLP transport protocol to be used for all telemetry data                                                |
| `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL` | Specifies the OTLP transport protocol to be used for trace data.                                                       |
| `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL`| Specifies the OTLP transport protocol to be used for metric data.                                                      |
| `OTEL_EXPORTER_OTLP_TEST`            | Specifies whether the OTLP exporter should be used in test mode. Usefull for debugging traces and metrics. Setting this value to true, will send the traces and metrics in the stdout.|

## Kubernetes Setup

Before using the Tracer and Meter, you need to ensure you can the correct setup in your Kubernetes deployment.
We have provided a sample you can see here: [examples/monitoring/k8s](../examples/monitoring/k8s/k8s.yaml).

The `Instrumentation` resource, configures "automatic" instrumentation. Howvever, here we need it to define the exporter URL, any environment variables we need to attach to our pods etc.
Finally, to use the `Instrumentation` you'd need to include the annotation:

```yaml
instrumentation.opentelemetry.io/inject-sdk: "{instrumentation}"
```

The `{instrumentation}` pattern is `{namespace}/{instrumentation-name}`. If the `Instrumentation` is on the same name with the Pods, then the pattern is `{instrumentation-name}`.

## Examples

You can find examples for the monitoring package:

- [Traces & Spans](../examples/monitoring/spans/)
- Distributed Tracing
  - [HTTP Tracing](../examples/monitoring/http/)
  - [PubSub Tracing](../examples/monitoring/pubsub/)
  - [RabbitMQ Tracing](../examples/monitoring/rabbitmq/)
  - [Middleware](../examples/monitoring/webserver/)
- Metrics
  - [Custom Metrics](../examples/monitoring/metrics/)

## Otel Documentation

- [Otel Documentation](https://opentelemetry.io/docs/)
- [Traces](https://opentelemetry.io/docs/concepts/signals/traces/)
- [Observability primer](https://opentelemetry.io/docs/concepts/observability-primer/)
- [Go](https://opentelemetry.io/docs/languages/go/)
  - [Getting Started](https://opentelemetry.io/docs/languages/go/getting-started/)
- [RabbitMQ]
  - [Otel Example](https://devandchill.com/posts/2021/12/go-step-by-step-guide-for-implementing-tracing-on-a-microservices-architecture-2/2/)
- [GCP PubSub]
  - [Otel Example](https://cloud.google.com/pubsub/docs/open-telemetry-tracing)
