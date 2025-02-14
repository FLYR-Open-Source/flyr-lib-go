// MIT License
//
// Copyright (c) 2025 FLYR, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package http // import "github.com/FlyrInc/flyr-lib-go/monitoring/http"

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
