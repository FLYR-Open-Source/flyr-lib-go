package metrics

import (
	"fmt"

	"github.com/DataDog/datadog-go/v5/statsd"
)

type Counter struct {
	name        string
	statsClient statsd.ClientInterface
}

func NewCounter(name string, statsClient statsd.ClientInterface) *Counter {
	return &Counter{
		name:        name,
		statsClient: statsClient,
	}
}

func (c *Counter) Count(value int64) error {
	err := c.statsClient.Count(c.name, value, nil, 1)
	if err != nil {
		return fmt.Errorf("statsd client count: %v", err)
	}
	return nil
}

func (c *Counter) Incr() error {
	err := c.statsClient.Incr(c.name, nil, 1)
	if err != nil {
		return fmt.Errorf("statsd client incr: %v", err)
	}
	return nil
}

func (c *Counter) Decr() error {
	err := c.statsClient.Decr(c.name, nil, 1)
	if err != nil {
		return fmt.Errorf("statsd client decr: %v", err)
	}
	return nil
}
