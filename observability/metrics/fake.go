package metrics

import "github.com/DataDog/datadog-go/v5/statsd"

type FakeMetricAgent struct {
	*statsd.NoOpClient
}

func NewFakeMetricAgent() *FakeMetricAgent {
	return &FakeMetricAgent{NoOpClient: &statsd.NoOpClient{}}
}

type FakeCounter struct{}

func (f *FakeCounter) Count(_ int64) error { return nil }
func (f *FakeCounter) Incr() error         { return nil }
func (f *FakeCounter) Decr() error         { return nil }

type FakeGauger struct{}

func (f *FakeGauger) Gauge(_ float64) error { return nil }

type FakeHistogramer struct{}

func (f *FakeHistogramer) Histogram(_ float64) error { return nil }
