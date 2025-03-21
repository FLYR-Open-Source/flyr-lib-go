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
	"errors"
	"fmt"
	"regexp"

	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/meter/units"
)

// MetricInput is the input metadata for creating a metric.
type MetricInput struct {
	// Description is the human-readable description of the metric.
	Description string
	// Unit is the unit of the metric.
	Unit units.Unit
}

// getDescription returns the description of the metric, including the unit in parentheses.
func (m *MetricInput) getDescription() string {
	return fmt.Sprintf("%s (Measured in %q).", m.Description, m.Unit)
}

// getUnit returns the unit of the metric. If the unit is not set, it returns the default unit.
func (m *MetricInput) getUnit(defaultUnit units.Unit) units.Unit {
	if m.Unit == "" {
		return defaultUnit
	}
	return m.Unit
}

// MetricInput is the input metadata for creating a histogram metric.
type HistogramMetricInput struct {
	MetricInput

	// ExplicitBucketBoundaries are the bucket boundaries for the histogram.
	ExplicitBucketBoundaries []float64
}

// The regular expression to validate the metric name.
// A name must be a sequence of one or more components separated by dots (.).
//
// Example: "http.server.latency"
var nameRegex = regexp.MustCompile("^[a-z0-9_.]+$")

// Errors
var (
	// ErrMeterNotInitialized is returned when a metric is created before the meter is initialized.
	ErrMeterNotInitialized = errors.New("meter not initialized")
	// ErrInvalidMetricName is returned when the metric name does not conform to the OpenTelemetry instrument name syntax.
	ErrInvalidMetricName = errors.New("invalid metric name. Must match the regex: " + nameRegex.String())
)
