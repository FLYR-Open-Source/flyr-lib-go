# Logger

The Logger package provides a custom JSON logger with built-in observability features. It extends [log/slog](https://pkg.go.dev/log/slog), a high-performance structured, and leveled logging that is build-in to Go standard library.

The Logger follows the [Open Telemetry Semantic Conventions](https://opentelemetry.io/docs/concepts/semantic-conventions/), to ensure compatibility with observability standards.

1. [SpanLogger](#spanlogger)
   - [Correlate the IDs with Spans](#correlate-the-ids-with-spans)
   - [Inject log attributes to Spans](#inject-log-attributes-to-spans)
2. [Environment Variables](#environment-variables)
3. [Examples](#examples)

## SpanLogger

You can use the logger idependently, however, combining with the [monitoring/](../monitoring/) package, the logger provides out-of-the-box the below extra features.

> To enable the below features, you must first follow the steps to setup observability: [monitoring/README.md](../monitoring/README.md).

### Correlate the IDs with Spans

When logging within a traced request, the logger automatically extracts and attaches trace and span IDs to each log entry. This ensures that logs can be correlated with their respective traces, making it easier to analyze distributed requests and debug issues across services. By linking logs and traces, you gain deeper visibility into request flows and system behavior.

### Inject log attributes to Spans

Developers often rely on logs as the primary source of truth when debugging. To enhance this, the logger automatically injects extra log attributes into the corresponding spans. This ensures that spans contain valuable contextual information, making it easier to analyze and debug issues by providing a more comprehensive view of the request flow.

## Environment Variables

The logger accepts a config that reads values from Environment Variables. The below table contains all the supported Environment Variables for the logger:

| Variable Name | Description                                                                         | Default   |
|---------------|-------------------------------------------------------------------------------------|-----------|
| `LOG_LEVEL`   | The log level. The accepted values can be one of (`debug`, `info`, `warn`, `error`) | `info`    |

## Examples

You can find examples for the logger:

- [Setup and use the logger](../examples/logging/)
- [Examples with Spans](../examples/monitoring/)
