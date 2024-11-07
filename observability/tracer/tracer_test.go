package tracer

import (
	"context"
	"testing"

	"github.com/FlyrInc/flyr-lib-go/pkg/testhelpers"
	"github.com/stretchr/testify/assert"

	"github.com/FlyrInc/flyr-lib-go/config"
)

func TestIncludeCallerAttributes(t *testing.T) {

	_, span := testhelpers.GetFakeSpan(context.Background())
	includeCallerAttributes(&span)

	assert.Equal(t, config.CODE_PATH, string(span.FakeAttributes[0].Key))
	assert.Contains(t, span.FakeAttributes[0].Value.AsString(), "src/testing/testing.go")

	assert.Equal(t, config.CODE_LINE, string(span.FakeAttributes[1].Key))
	assert.Positive(t, span.FakeAttributes[1].Value.AsInt64())

	assert.Equal(t, config.CODE_FUNC, string(span.FakeAttributes[2].Key))
	assert.Contains(t, span.FakeAttributes[2].Value.AsString(), "tRunner")

	assert.Equal(t, config.CODE_NS, string(span.FakeAttributes[3].Key))
	assert.Contains(t, span.FakeAttributes[3].Value.AsString(), "testing")
}
