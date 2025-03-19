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

package meter

import (
	"testing"

	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/meter/units"
	"github.com/stretchr/testify/assert"
)

func TestMetricInput_getDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		unit        units.Unit
		expected    string
	}{
		{
			name:        "with unit",
			description: "CPU usage",
			unit:        units.Percent,
			expected:    "CPU usage (Measured in \"%\").",
		},
		{
			name:        "empty unit",
			description: "Memory usage",
			unit:        "",
			expected:    "Memory usage (Measured in \"\").",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricInput := &MetricInput{
				Description: tt.description,
				Unit:        tt.unit,
			}
			assert.Equal(t, tt.expected, metricInput.getDescription())
		})
	}
}

func TestMetricInput_getUnit(t *testing.T) {
	tests := []struct {
		name         string
		unit         units.Unit
		defaultUnit  units.Unit
		expectedUnit units.Unit
	}{
		{
			name:         "unit set",
			unit:         units.Seconds,
			defaultUnit:  units.Milliseconds,
			expectedUnit: units.Seconds,
		},
		{
			name:         "unit not set",
			unit:         "",
			defaultUnit:  units.Milliseconds,
			expectedUnit: units.Milliseconds,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricInput := &MetricInput{
				Unit: tt.unit,
			}
			assert.Equal(t, tt.expectedUnit, metricInput.getUnit(tt.defaultUnit))
		})
	}
}

func TestHistogramMetricInput(t *testing.T) {
	t.Run("histogram metric input with explicit bucket boundaries", func(t *testing.T) {
		histogramInput := &HistogramMetricInput{
			MetricInput: MetricInput{
				Description: "Request latency",
				Unit:        units.Milliseconds,
			},
			ExplicitBucketBoundaries: []float64{0.1, 0.5, 1.0},
		}

		assert.Equal(t, "Request latency (Measured in \"ms\").", histogramInput.getDescription())
		assert.Equal(t, units.Milliseconds, histogramInput.getUnit(units.Seconds))
		assert.Equal(t, []float64{0.1, 0.5, 1.0}, histogramInput.ExplicitBucketBoundaries)
	})
}

func TestNameRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid metric name",
			input:    "http.server.latency",
			expected: true,
		},
		{
			name:     "invalid metric name with uppercase",
			input:    "HTTP.server.latency",
			expected: false,
		},
		{
			name:     "invalid metric name with dashes",
			input:    "http.my-service.latency",
			expected: false,
		},
		{
			name:     "invalid metric name with special characters",
			input:    "http.server.latency!",
			expected: false,
		},
		{
			name:     "invalid metric name with spaces",
			input:    "http server latency",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, nameRegex.MatchString(tt.input))
		})
	}
}
