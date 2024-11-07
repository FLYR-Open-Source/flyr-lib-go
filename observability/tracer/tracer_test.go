package tracer

import (
	"context"
	"testing"

	"github.com/FlyrInc/flyr-lib-go/config"
	"github.com/FlyrInc/flyr-lib-go/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestStartSpan(t *testing.T) {
	pc, fakeTracer := testhelpers.GetFakeTracer()
	//nolint:errcheck
	defer pc.Shutdown(context.Background())

	tracer := Tracer{tracer: fakeTracer}

	_, span := tracer.StartSpan(context.Background(), "test-span", oteltrace.SpanKindInternal)
	defer span.End()

	testSpan := span.Span.(*testhelpers.FakeSpan)
	assert.Equal(t, config.CODE_PATH, string(testSpan.FakeAttributes[0].Key))
	assert.Contains(t, testSpan.FakeAttributes[0].Value.AsString(), "src/testing/testing.go")

	assert.Equal(t, config.CODE_LINE, string(testSpan.FakeAttributes[1].Key))
	assert.Positive(t, testSpan.FakeAttributes[1].Value.AsInt64())

	assert.Equal(t, config.CODE_FUNC, string(testSpan.FakeAttributes[2].Key))
	assert.Contains(t, testSpan.FakeAttributes[2].Value.AsString(), "tRunner")

	assert.Equal(t, config.CODE_NS, string(testSpan.FakeAttributes[3].Key))
	assert.Contains(t, testSpan.FakeAttributes[3].Value.AsString(), "testing")
}
