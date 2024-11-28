package rabbitmq

import (
	"context"
	"testing"

	"github.com/FlyrInc/flyr-lib-go/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestAmqpHeadersCarrier(t *testing.T) {
	// Initialize AmqpHeadersCarrier
	headers := AmqpHeadersCarrier{
		"traceID": "some-trace-id",
		"spanID":  "some-span-id",
	}

	// Test Get method
	t.Run("Get", func(t *testing.T) {
		assert.Equal(t, "some-trace-id", headers.Get("traceID"))
		assert.Equal(t, "", headers.Get("nonexistent"))
	})

	// Test Set method
	t.Run("Set", func(t *testing.T) {
		headers.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-3fd3e2d5e8d13b58-01")
		assert.Equal(t, "00-4bf92f3577b34da6a3ce929d0e0e4736-3fd3e2d5e8d13b58-01", headers.Get("traceparent"))
	})

	// Test Keys method
	t.Run("Keys", func(t *testing.T) {
		keys := headers.Keys()
		assert.Contains(t, keys, "traceID")
		assert.Contains(t, keys, "spanID")
		assert.Contains(t, keys, "traceparent")
	})
}

func TestInjectAMQPHeaders(t *testing.T) {
	// Setup a mock context with a span
	ctx, span := testhelpers.GetFakeSpan(context.Background())
	defer span.End()

	// Inject headers
	headers := InjectAMQPHeaders(ctx)

	// Test that the headers contain the expected tracing data
	assert.NotEmpty(t, headers)
	assert.Contains(t, headers, "traceparent")
}

func TestExtractAMQPHeaders(t *testing.T) {
	// Setup a mock context
	ctx := context.Background()

	// Example headers that would be passed to ExtractAMQPHeaders
	headers := map[string]interface{}{
		"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-3fd3e2d5e8d13b58-01",
	}

	// Extract headers into context
	extractedCtx := ExtractAMQPHeaders(ctx, headers)

	// Test that the context is now carrying the correct tracing info
	_, span := testhelpers.GetFakeSpan(extractedCtx)
	assert.NotNil(t, span)
	span.End()
}
