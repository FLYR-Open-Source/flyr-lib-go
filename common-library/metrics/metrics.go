package metrics

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

// InitMetrics initializes a Meter with the provided service name
func InitMetrics(serviceName string) metric.Meter {
	// Get the global meter provider
	meterProvider := otel.GetMeterProvider()

	// If no meter provider is configured, use a no-op provider
	if meterProvider == nil  {
		meterProvider = noop.NewMeterProvider() // No-op meter for testing or when no metrics provider is set
	}

	return meterProvider.Meter(serviceName)
}

// RecordMetric records a custom metric with the given name and value
func RecordMetric(meter metric.Meter, name string, value float64) {
	// Add logic to filter out high-cardinality metrics
	if isHighCardinalityMetric(name) {
		log.Printf("Skipping high-cardinality metric: %s", name)
		return
	}

	// Create and record a float64 counter directly through the meter
	counter, err := meter.Float64Counter(name)
	if err != nil {
		log.Printf("Error creating counter: %v", err)
		return
	}

	counter.Add(context.Background(), value)
}

// isHighCardinalityMetric checks if the metric is of high cardinality
func isHighCardinalityMetric(name string) bool {
	// Add your logic to determine high-cardinality metrics here
	return false
}
