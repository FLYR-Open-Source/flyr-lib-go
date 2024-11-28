// Package logger provides a simple logger interface for logging messages.
// It also provides an out of the box support with the Otel span.
// The logger is passing metadata to the span, so the logs can be easily correlated with the traces.
// Furthermore, the logger is recording the span as errored for error logs with a given error.
package logger
