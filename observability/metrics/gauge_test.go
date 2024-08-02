package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGauger_Gauge(t *testing.T) {
	client := NewFakeMetricAgent()
	gauger := NewGauger("test", client)

	assert.NoError(t, gauger.Gauge(10))
}
