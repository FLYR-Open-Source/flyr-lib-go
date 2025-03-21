// MIT License
//
// Copyright (c) 2025 FLYR, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package meter // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/meter"

import (
	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/meter/units"
	"go.opentelemetry.io/otel/metric"
)

var (
	// LATENCY_EXPLICIT_BUCKET_BOUNDARIES_IN_MS is the default bucket boundaries for latency histograms (in milliseconds).
	// For measuring latency (e.g., HTTP request durations, database query times), you typically want finer
	// granularity at the lower end (where most values are expected) and coarser granularity at the higher
	// end (to capture outliers).
	LATENCY_EXPLICIT_BUCKET_BOUNDARIES_IN_MS = []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 10000}
	// LATENCY_EXPLICIT_BUCKET_BOUNDARIES_IN_SECONDS is the default bucket boundaries for latency histograms (in seconds).
	// For measuring latency (e.g., HTTP request durations, database query times), you typically want finer
	// granularity at the lower end (where most values are expected) and coarser granularity at the higher
	// end (to capture outliers).
	//
	// This bucket is heavily influenced by the default buckets of Prometheus clients.
	// e.g.
	//  - Java: https://github.com/prometheus/client_java/blob/6730f3e32199d6bf0e963b306ff69ef08ac5b178/simpleclient/src/main/java/io/prometheus/client/Histogram.java#L88
	//  - Go: https://github.com/prometheus/client_golang/blob/83d56b1144a0c2eb10d399e7abbae3333bebc463/prometheus/histogram.go#L68
	LATENCY_EXPLICIT_BUCKET_BOUNDARIES_IN_SECONDS = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
	// REQUEST_RESPONSE_EXPLICIT_BUCKET_BOUNDARIES is the default bucket boundaries for request/response
	// size histograms (in bytes).
	// For measuring the size of requests or responses (e.g., in bytes), you typically want exponential
	// or logarithmic bucket boundaries
	// to cover a wide range of sizes.
	REQUEST_RESPONSE_EXPLICIT_BUCKET_BOUNDARIES = []float64{0, 1024, 2048, 5120, 10240, 20480, 51200, 102400, 204800, 512000, 1048576}
	// ERROR_RATE_EXPLICIT_BUCKET_BOUNDARIES is the default bucket boundaries for error rate histograms(in %).
	// For measuring error rates, you most likely want linear bucket boundaries.
	ERROR_RATE_EXPLICIT_BUCKET_BOUNDARIES = []float64{0, 0.01, 0.1, 1, 5, 10, 20, 50, 100}
	// COUNT_METRIC_EXPLICIT_BUCKET_BOUNDARIES is the default bucket boundaries for count metrics.
	// For measuring counts (e.g., number of requests, number of items in a queue), you most likely want
	// linear bucket boundaries.
	COUNT_METRIC_EXPLICIT_BUCKET_BOUNDARIES = []float64{0, 1, 2, 5, 10, 25, 50, 100, 250, 500, 1000}
	// GENERIC_EXPLICIT_BUCKET_BOUNDARIES is the default bucket boundaries for generic histograms.
	// For generic histograms, you may want to use a mix of linear and exponential bucket boundaries.
	GENERIC_EXPLICIT_BUCKET_BOUNDARIES = []float64{0, 1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000}
)

// FloatHistogram returns a new go.opentelemetry.io/otel/metric.Float64Histogram instrument identified by
// name and configured with options. The instrument is used to
// synchronously record the distribution of float64 measurements during a
// computational operation.
//
// The name needs to conform to the OpenTelemetry instrument name syntax.
// See the Instrument Name section of the package documentation for more
// information (https://opentelemetry.io/docs/specs/semconv/general/metrics/).
//
// The description is optional, but recommended. It is used to describe
// the instrument in human-readable terms.
//
// The unit is optional, but recommended. It is used to describe the unit of
// the measurements. The default unit is "ms" (Milliseconds).
// However, for better documentation, the unit should be set to a more meaningful value.
//
// The explicitBucketBoundaries are the bucket boundaries for the histogram. Make sure
// to set the bucket boundaries according to the metric you are measuring.
// You can use the default bucket boundaries provided in this package or set your own.
func FloatHistogram(name string, input HistogramMetricInput) (metric.Float64Histogram, error) {
	if defaultMeter == nil {
		return nil, ErrMeterNotInitialized
	}

	if !nameRegex.MatchString(name) {
		return nil, ErrInvalidMetricName
	}

	unit := input.getUnit(units.Milliseconds)
	description := input.getDescription()

	return defaultMeter.Float64Histogram(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit.String()),
		metric.WithExplicitBucketBoundaries(input.ExplicitBucketBoundaries...),
	)
}

// IntHistogram returns a new go.opentelemetry.io/otel/metric.Int64Histogram instrument identified by
// name and configured with options. The instrument is used to
// synchronously record the distribution of int64 measurements during a
// computational operation.
//
// The name needs to conform to the OpenTelemetry instrument name syntax.
// See the Instrument Name section of the package documentation for more
// information (https://opentelemetry.io/docs/specs/semconv/general/metrics/).
//
// The description is optional, but recommended. It is used to describe
// the instrument in human-readable terms.
//
// The unit is optional, but recommended. It is used to describe the unit of
// the measurements. The default unit is "ms" (Milliseconds).
//
// The explicitBucketBoundaries are the bucket boundaries for the histogram. Make sure
// to set the bucket boundaries according to the metric you are measuring.
// You can use the default bucket boundaries provided in this package or set your own.
func IntHistogram(name string, input HistogramMetricInput) (metric.Int64Histogram, error) {
	if defaultMeter == nil {
		return nil, ErrMeterNotInitialized
	}

	if !nameRegex.MatchString(name) {
		return nil, ErrInvalidMetricName
	}

	unit := input.getUnit(units.Milliseconds)
	description := input.getDescription()

	return defaultMeter.Int64Histogram(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit.String()),
		metric.WithExplicitBucketBoundaries(input.ExplicitBucketBoundaries...),
	)
}
