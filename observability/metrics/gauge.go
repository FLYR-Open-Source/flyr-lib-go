package metrics

import (
	"fmt"

	"github.com/DataDog/datadog-go/v5/statsd"
)

type Gauger struct {
	name        string
	statsClient statsd.ClientInterface
}

func NewGauger(name string, statsClient statsd.ClientInterface) *Gauger {
	return &Gauger{
		name:        name,
		statsClient: statsClient,
	}
}

func (g *Gauger) Gauge(value float64) error {
	err := g.statsClient.Gauge(g.name, value, nil, 1)
	if err != nil {
		return fmt.Errorf("statsd client gauge: %v", err)
	}
	return nil
}
