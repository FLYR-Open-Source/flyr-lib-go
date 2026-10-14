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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/meter/units"
)

// useManualReaderMeter points the package-level defaultMeter at a meter backed by a
// manual reader, and restores the previous value when the test ends.
func useManualReaderMeter(t *testing.T) *sdkmetric.ManualReader {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	previous := defaultMeter
	defaultMeter = provider.Meter("instruments-test")

	t.Cleanup(func() {
		defaultMeter = previous
		require.NoError(t, provider.Shutdown(context.Background()))
	})

	return reader
}

// collectMetric collects from the reader and returns the metric with the given name.
func collectMetric(t *testing.T, reader sdkmetric.Reader, name string) metricdata.Metrics {
	t.Helper()

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))
	require.Len(t, rm.ScopeMetrics, 1)

	for _, m := range rm.ScopeMetrics[0].Metrics {
		if m.Name == name {
			return m
		}
	}

	require.Failf(t, "metric not found", "metric %q was not collected", name)
	return metricdata.Metrics{}
}

var testAttr = attribute.String("env", "test")

// instrumentCase describes one public metric constructor. create builds the instrument
// and returns a function that records a known measurement on it.
type instrumentCase struct {
	name        string
	defaultUnit units.Unit
	create      func(name string, input MetricInput) (record func(ctx context.Context), err error)
	// assertData asserts the data recorded by the record function.
	assertData func(t *testing.T, data metricdata.Aggregation)
}

func instrumentCases() []instrumentCase {
	return []instrumentCase{
		{
			name:        "FloatCounter",
			defaultUnit: units.Ratio,
			create: func(name string, input MetricInput) (func(context.Context), error) {
				c, err := FloatCounter(name, input)
				if err != nil {
					return nil, err
				}
				return func(ctx context.Context) {
					c.Add(ctx, 1.5, metric.WithAttributes(testAttr))
					c.Add(ctx, 2.25, metric.WithAttributes(testAttr))
				}, nil
			},
			assertData: func(t *testing.T, data metricdata.Aggregation) {
				sum, ok := data.(metricdata.Sum[float64])
				require.True(t, ok, "unexpected aggregation %T", data)
				assert.True(t, sum.IsMonotonic)
				require.Len(t, sum.DataPoints, 1)
				assert.InDelta(t, 3.75, sum.DataPoints[0].Value, 1e-9)
				assert.True(t, sum.DataPoints[0].Attributes.HasValue(testAttr.Key))
			},
		},
		{
			name:        "IntCounter",
			defaultUnit: units.Ratio,
			create: func(name string, input MetricInput) (func(context.Context), error) {
				c, err := IntCounter(name, input)
				if err != nil {
					return nil, err
				}
				return func(ctx context.Context) {
					c.Add(ctx, 2, metric.WithAttributes(testAttr))
					c.Add(ctx, 3, metric.WithAttributes(testAttr))
				}, nil
			},
			assertData: func(t *testing.T, data metricdata.Aggregation) {
				sum, ok := data.(metricdata.Sum[int64])
				require.True(t, ok, "unexpected aggregation %T", data)
				assert.True(t, sum.IsMonotonic)
				require.Len(t, sum.DataPoints, 1)
				assert.Equal(t, int64(5), sum.DataPoints[0].Value)
				assert.True(t, sum.DataPoints[0].Attributes.HasValue(testAttr.Key))
			},
		},
		{
			name:        "FloatGauge",
			defaultUnit: units.Bytes,
			create: func(name string, input MetricInput) (func(context.Context), error) {
				g, err := FloatGauge(name, input)
				if err != nil {
					return nil, err
				}
				return func(ctx context.Context) {
					g.Record(ctx, 10.5, metric.WithAttributes(testAttr))
					g.Record(ctx, 7.25, metric.WithAttributes(testAttr))
				}, nil
			},
			assertData: func(t *testing.T, data metricdata.Aggregation) {
				gauge, ok := data.(metricdata.Gauge[float64])
				require.True(t, ok, "unexpected aggregation %T", data)
				require.Len(t, gauge.DataPoints, 1)
				// a gauge keeps only the last recorded value
				assert.InDelta(t, 7.25, gauge.DataPoints[0].Value, 1e-9)
				assert.True(t, gauge.DataPoints[0].Attributes.HasValue(testAttr.Key))
			},
		},
		{
			name:        "IntGauge",
			defaultUnit: units.Bytes,
			create: func(name string, input MetricInput) (func(context.Context), error) {
				g, err := IntGauge(name, input)
				if err != nil {
					return nil, err
				}
				return func(ctx context.Context) {
					g.Record(ctx, 10, metric.WithAttributes(testAttr))
					g.Record(ctx, 7, metric.WithAttributes(testAttr))
				}, nil
			},
			assertData: func(t *testing.T, data metricdata.Aggregation) {
				gauge, ok := data.(metricdata.Gauge[int64])
				require.True(t, ok, "unexpected aggregation %T", data)
				require.Len(t, gauge.DataPoints, 1)
				assert.Equal(t, int64(7), gauge.DataPoints[0].Value)
				assert.True(t, gauge.DataPoints[0].Attributes.HasValue(testAttr.Key))
			},
		},
		{
			name:        "FloatHistogram",
			defaultUnit: units.Milliseconds,
			create: func(name string, input MetricInput) (func(context.Context), error) {
				h, err := FloatHistogram(name, HistogramMetricInput{MetricInput: input})
				if err != nil {
					return nil, err
				}
				return func(ctx context.Context) {
					h.Record(ctx, 1.5, metric.WithAttributes(testAttr))
					h.Record(ctx, 2.5, metric.WithAttributes(testAttr))
				}, nil
			},
			assertData: func(t *testing.T, data metricdata.Aggregation) {
				hist, ok := data.(metricdata.Histogram[float64])
				require.True(t, ok, "unexpected aggregation %T", data)
				require.Len(t, hist.DataPoints, 1)
				assert.Equal(t, uint64(2), hist.DataPoints[0].Count)
				assert.InDelta(t, 4.0, hist.DataPoints[0].Sum, 1e-9)
			},
		},
		{
			name:        "IntHistogram",
			defaultUnit: units.Milliseconds,
			create: func(name string, input MetricInput) (func(context.Context), error) {
				h, err := IntHistogram(name, HistogramMetricInput{MetricInput: input})
				if err != nil {
					return nil, err
				}
				return func(ctx context.Context) {
					h.Record(ctx, 3, metric.WithAttributes(testAttr))
					h.Record(ctx, 4, metric.WithAttributes(testAttr))
				}, nil
			},
			assertData: func(t *testing.T, data metricdata.Aggregation) {
				hist, ok := data.(metricdata.Histogram[int64])
				require.True(t, ok, "unexpected aggregation %T", data)
				require.Len(t, hist.DataPoints, 1)
				assert.Equal(t, uint64(2), hist.DataPoints[0].Count)
				assert.Equal(t, int64(7), hist.DataPoints[0].Sum)
			},
		},
		{
			name:        "FloatUpDownCounter",
			defaultUnit: units.Ratio,
			create: func(name string, input MetricInput) (func(context.Context), error) {
				c, err := FloatUpDownCounter(name, input)
				if err != nil {
					return nil, err
				}
				return func(ctx context.Context) {
					c.Add(ctx, 5.5, metric.WithAttributes(testAttr))
					c.Add(ctx, -2.0, metric.WithAttributes(testAttr))
				}, nil
			},
			assertData: func(t *testing.T, data metricdata.Aggregation) {
				sum, ok := data.(metricdata.Sum[float64])
				require.True(t, ok, "unexpected aggregation %T", data)
				assert.False(t, sum.IsMonotonic)
				require.Len(t, sum.DataPoints, 1)
				assert.InDelta(t, 3.5, sum.DataPoints[0].Value, 1e-9)
			},
		},
		{
			name:        "IntUpDownCounter",
			defaultUnit: units.Ratio,
			create: func(name string, input MetricInput) (func(context.Context), error) {
				c, err := IntUpDownCounter(name, input)
				if err != nil {
					return nil, err
				}
				return func(ctx context.Context) {
					c.Add(ctx, 5, metric.WithAttributes(testAttr))
					c.Add(ctx, -3, metric.WithAttributes(testAttr))
				}, nil
			},
			assertData: func(t *testing.T, data metricdata.Aggregation) {
				sum, ok := data.(metricdata.Sum[int64])
				require.True(t, ok, "unexpected aggregation %T", data)
				assert.False(t, sum.IsMonotonic)
				require.Len(t, sum.DataPoints, 1)
				assert.Equal(t, int64(2), sum.DataPoints[0].Value)
			},
		},
	}
}

func TestInstruments_MeterNotInitialized(t *testing.T) {
	for _, tc := range instrumentCases() {
		t.Run(tc.name, func(t *testing.T) {
			previous := defaultMeter
			defaultMeter = nil
			t.Cleanup(func() { defaultMeter = previous })

			record, err := tc.create("valid.name", MetricInput{})
			require.ErrorIs(t, err, ErrMeterNotInitialized)
			assert.Nil(t, record)
		})
	}
}

func TestInstruments_InvalidName(t *testing.T) {
	invalidNames := []string{"", "Upper.Case", "with-dash", "with space", "bang!"}

	for _, tc := range instrumentCases() {
		for _, invalid := range invalidNames {
			t.Run(tc.name+"/"+invalid, func(t *testing.T) {
				useManualReaderMeter(t)

				record, err := tc.create(invalid, MetricInput{})
				require.ErrorIs(t, err, ErrInvalidMetricName)
				assert.Nil(t, record)
			})
		}
	}
}

func TestInstruments_MetadataAndValues(t *testing.T) {
	tests := []struct {
		name         string
		input        MetricInput
		expectedUnit func(defaultUnit units.Unit) string
		expectedDesc string
	}{
		{
			name:         "default unit",
			input:        MetricInput{Description: "Some measurement"},
			expectedUnit: func(defaultUnit units.Unit) string { return defaultUnit.String() },
			expectedDesc: `Some measurement (Measured in "").`,
		},
		{
			name:         "custom unit",
			input:        MetricInput{Description: "Some measurement", Unit: units.Seconds},
			expectedUnit: func(units.Unit) string { return "s" },
			expectedDesc: `Some measurement (Measured in "s").`,
		},
	}

	for _, tc := range instrumentCases() {
		for _, tt := range tests {
			t.Run(tc.name+"/"+tt.name, func(t *testing.T) {
				reader := useManualReaderMeter(t)

				record, err := tc.create("test.instrument_1", tt.input)
				require.NoError(t, err)
				require.NotNil(t, record)

				record(context.Background())

				m := collectMetric(t, reader, "test.instrument_1")
				assert.Equal(t, tt.expectedUnit(tc.defaultUnit), m.Unit)
				assert.Equal(t, tt.expectedDesc, m.Description)
				tc.assertData(t, m.Data)
			})
		}
	}
}

func TestHistograms_ExplicitBucketBoundaries(t *testing.T) {
	customBounds := []float64{1, 5, 10}

	// Each constructor records the values 0.5 (0 for ints), 3, 7 and 50, which must
	// land in the buckets (-inf,1], (1,5], (5,10] and (10,+inf).
	tests := []struct {
		name string
		// record creates the histogram, records the values and returns the collected data point's
		// bucket boundaries and counts.
		record func(t *testing.T, input HistogramMetricInput) ([]float64, []uint64)
	}{
		{
			name: "FloatHistogram",
			record: func(t *testing.T, input HistogramMetricInput) ([]float64, []uint64) {
				reader := useManualReaderMeter(t)
				h, err := FloatHistogram("test.float_histogram", input)
				require.NoError(t, err)
				for _, v := range []float64{0.5, 3, 7, 50} {
					h.Record(context.Background(), v)
				}
				hist, ok := collectMetric(t, reader, "test.float_histogram").Data.(metricdata.Histogram[float64])
				require.True(t, ok)
				require.Len(t, hist.DataPoints, 1)
				return hist.DataPoints[0].Bounds, hist.DataPoints[0].BucketCounts
			},
		},
		{
			name: "IntHistogram",
			record: func(t *testing.T, input HistogramMetricInput) ([]float64, []uint64) {
				reader := useManualReaderMeter(t)
				h, err := IntHistogram("test.int_histogram", input)
				require.NoError(t, err)
				for _, v := range []int64{0, 3, 7, 50} {
					h.Record(context.Background(), v)
				}
				hist, ok := collectMetric(t, reader, "test.int_histogram").Data.(metricdata.Histogram[int64])
				require.True(t, ok)
				require.Len(t, hist.DataPoints, 1)
				return hist.DataPoints[0].Bounds, hist.DataPoints[0].BucketCounts
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/custom boundaries", func(t *testing.T) {
			bounds, counts := tt.record(t, HistogramMetricInput{ExplicitBucketBoundaries: customBounds})

			assert.Equal(t, customBounds, bounds)
			assert.Equal(t, []uint64{1, 1, 1, 1}, counts)
		})

		t.Run(tt.name+"/sdk default boundaries", func(t *testing.T) {
			bounds, _ := tt.record(t, HistogramMetricInput{})

			assert.NotEmpty(t, bounds)
			assert.NotEqual(t, customBounds, bounds)
		})
	}
}

func TestHistogramBucketBoundaryPresets(t *testing.T) {
	tests := []struct {
		name       string
		boundaries []float64
	}{
		{"LATENCY_EXPLICIT_BUCKET_BOUNDARIES_IN_MS", LATENCY_EXPLICIT_BUCKET_BOUNDARIES_IN_MS},
		{"LATENCY_EXPLICIT_BUCKET_BOUNDARIES_IN_SECONDS", LATENCY_EXPLICIT_BUCKET_BOUNDARIES_IN_SECONDS},
		{"REQUEST_RESPONSE_EXPLICIT_BUCKET_BOUNDARIES", REQUEST_RESPONSE_EXPLICIT_BUCKET_BOUNDARIES},
		{"ERROR_RATE_EXPLICIT_BUCKET_BOUNDARIES", ERROR_RATE_EXPLICIT_BUCKET_BOUNDARIES},
		{"COUNT_METRIC_EXPLICIT_BUCKET_BOUNDARIES", COUNT_METRIC_EXPLICIT_BUCKET_BOUNDARIES},
		{"GENERIC_EXPLICIT_BUCKET_BOUNDARIES", GENERIC_EXPLICIT_BUCKET_BOUNDARIES},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := useManualReaderMeter(t)

			h, err := FloatHistogram("test.preset", HistogramMetricInput{ExplicitBucketBoundaries: tt.boundaries})
			require.NoError(t, err)
			h.Record(context.Background(), 1)

			hist, ok := collectMetric(t, reader, "test.preset").Data.(metricdata.Histogram[float64])
			require.True(t, ok)
			require.Len(t, hist.DataPoints, 1)
			assert.Equal(t, tt.boundaries, hist.DataPoints[0].Bounds)
			assert.Len(t, hist.DataPoints[0].BucketCounts, len(tt.boundaries)+1)
		})
	}
}
