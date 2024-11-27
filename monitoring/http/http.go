package http // import "github.com/FlyrInc/flyr-lib-go/http"

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// SetHttpTransport configures the provided HTTP client to use OpenTelemetry's transport for tracing.
//
// This function takes an http.Client as an argument and sets its Transport to an
// OpenTelemetry-enabled transport created by otelhttp.NewTransport. This allows for
// tracing of outgoing HTTP requests made by the client, enabling better observability
// and monitoring of requests in a distributed system.
//
// Returns the configured http.Client with the OpenTelemetry transport set.
func SetHttpTransport(client http.Client) http.Client {
	client.Transport = otelhttp.NewTransport(http.DefaultTransport)
	return client
}

// NewHttpClient initializes a new HTTP client with OpenTelemetry tracing enabled.
//
// This function creates and returns an http.Client configured to use an OpenTelemetry
// transport by wrapping the default HTTP transport. This allows for tracing of all
// outgoing HTTP requests made by the client, providing enhanced observability for
// applications that rely on external HTTP communications.
//
// Returns a new http.Client with OpenTelemetry tracing configured.
func NewHttpClient() http.Client {
	return http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}
