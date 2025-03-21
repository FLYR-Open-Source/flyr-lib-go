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

// FloatGauge returns a new go.opentelemetry.io/otel/metric.Float64Gauge instrument identified by name and
// configured with options. The instrument is used to synchronously record
// instantaneous float64 measurements during a computational operation.
//
// The name needs to conform to the OpenTelemetry instrument name syntax.
// See the Instrument Name section of the package documentation for more
// information (https://opentelemetry.io/docs/specs/semconv/general/metrics/).
//
// The description is optional, but recommended. It is used to describe
// the instrument in human-readable terms.
//
// The unit is optional, but recommended. It is used to describe the unit of
// the measurements. The default unit is "By" (Bytes).
// However, for better documentation, the unit should be set to a more meaningful value.
func FloatGauge(name string, input MetricInput) (metric.Float64Gauge, error) {
	if defaultMeter == nil {
		return nil, ErrMeterNotInitialized
	}

	if !nameRegex.MatchString(name) {
		return nil, ErrInvalidMetricName
	}

	unit := input.getUnit(units.Bytes)
	description := input.getDescription()

	return defaultMeter.Float64Gauge(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit.String()),
	)
}

// IntGauge returns a new go.opentelemetry.io/otel/metric.Int64Gauge instrument identified by name and
// configured with options. The instrument is used to synchronously record
// instantaneous int64 measurements during a computational operation.
//
// The name needs to conform to the OpenTelemetry instrument name syntax.
// See the Instrument Name section of the package documentation for more
// information (https://opentelemetry.io/docs/specs/semconv/general/metrics/).
//
// The description is optional, but recommended. It is used to describe
// the instrument in human-readable terms.
//
// The unit is optional, but recommended. It is used to describe the unit of
// the measurements. The default unit is "By" (Bytes).
// However, for better documentation, the unit should be set to a more meaningful value.
func IntGauge(name string, input MetricInput) (metric.Int64Gauge, error) {
	if defaultMeter == nil {
		return nil, ErrMeterNotInitialized
	}

	if !nameRegex.MatchString(name) {
		return nil, ErrInvalidMetricName
	}

	unit := input.getUnit(units.Bytes)
	description := input.getDescription()

	return defaultMeter.Int64Gauge(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit.String()),
	)
}
