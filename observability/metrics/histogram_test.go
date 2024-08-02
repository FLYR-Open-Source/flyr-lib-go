package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistogramer_Histogram(t *testing.T) {
	client := NewFakeMetricAgent()
	histogramer := NewHistogramer("test", client)

	assert.NoError(t, histogramer.Histogram(10))
}
