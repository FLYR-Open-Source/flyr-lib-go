package metrics

import (
	"fmt"

	"github.com/DataDog/datadog-go/v5/statsd"
)

type Histogramer struct {
	name         string
	statsdClient statsd.ClientInterface
}

func NewHistogramer(name string, statsdClient statsd.ClientInterface) *Histogramer {
	return &Histogramer{name: name, statsdClient: statsdClient}
}

func (h *Histogramer) Histogram(value float64) error {
	err := h.statsdClient.Histogram(h.name, value, nil, 1)
	if err != nil {
		return fmt.Errorf("statsd client histogram: %v", err)
	}
	return nil
}
