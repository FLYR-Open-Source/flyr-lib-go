package rabbitmq // import "github.com/FlyrInc/flyr-lib-go/monitoring/rabbitmq"

import (
	"context"

	"go.opentelemetry.io/otel"
)

// AmqpHeadersCarrier is a map of headers that implements the TextMapCarrier interface
type AmqpHeadersCarrier map[string]interface{}

// Get returns the value for the key
func (a AmqpHeadersCarrier) Get(key string) string {
	v, ok := a[key]
	if !ok {
		return ""
	}
	return v.(string)
}

// Set sets the value for the key
func (a AmqpHeadersCarrier) Set(key string, value string) {
	a[key] = value
}

// Keys returns the keys of the carrier
func (a AmqpHeadersCarrier) Keys() []string {
	i := 0
	r := make([]string, len(a))

	for k := range a {
		r[i] = k
		i++
	}

	return r
}

// InjectAMQPHeaders injects the tracing from the context into the header map
func InjectAMQPHeaders(ctx context.Context) map[string]interface{} {
	h := AmqpHeadersCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, h)
	return h
}

// ExtractAMQPHeaders extracts the tracing from the header and puts it into the context
//
// Returns a new context with the tracing information from the headers
func ExtractAMQPHeaders(ctx context.Context, headers map[string]interface{}) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, AmqpHeadersCarrier(headers))
}
