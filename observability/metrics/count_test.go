package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounter_Count(t *testing.T) {
	client := NewFakeMetricAgent()
	counter := NewCounter("test_counter", client)

	assert.NoError(t, counter.Count(10))
}

func TestCounter_Incr(t *testing.T) {
	client := NewFakeMetricAgent()
	counter := NewCounter("test_counter", client)

	assert.NoError(t, counter.Incr())
}

func TestCounter_Decr(t *testing.T) {
	client := NewFakeMetricAgent()
	counter := NewCounter("test_counter", client)

	assert.NoError(t, counter.Decr())
}
