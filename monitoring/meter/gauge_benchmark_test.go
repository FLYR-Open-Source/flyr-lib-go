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
	"fmt"
	"math/rand/v2"
	"os"
	"testing"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/meter/units"
	fakemonitoring "github.com/FLYR-Open-Source/flyr-lib-go/pkg/testhelpers/monitoring"
)

func BenchmarkFloatGauge(b *testing.B) {
	config.ResetMonitoringConfig()
	_ = os.Setenv("OTEL_SERVICE_NAME", "test-service")
	_ = os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")

	mockLogServer := &fakemonitoring.MockOtelMetricsServer{}
	_, options, err := fakemonitoring.NewOtelMetricsGrpcServer(mockLogServer)
	if err != nil {
		b.Fatalf("failed to create mock logging service server: %v", err)
		return
	}

	otlpEndpoint := fmt.Sprintf("http://%v", options[0])

	_ = os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", otlpEndpoint)

	b.ResetTimer()
	ctx := b.Context()
	_, err = StartDefaultMeter(ctx)
	if err != nil {
		b.Fatal(err)
	}

	gauge, err := FloatGauge("test_float_gauge", MetricInput{
		Description: "test",
		Unit:        units.Milliseconds,
	})
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		n := float64(rand.IntN(999) + 1)
		gauge.Record(ctx, n)
	}
}

func BenchmarkIntGauge(b *testing.B) {
	config.ResetMonitoringConfig()
	_ = os.Setenv("OTEL_SERVICE_NAME", "test-service")
	_ = os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")

	mockLogServer := &fakemonitoring.MockOtelMetricsServer{}
	_, options, err := fakemonitoring.NewOtelMetricsGrpcServer(mockLogServer)
	if err != nil {
		b.Fatalf("failed to create mock logging service server: %v", err)
		return
	}

	otlpEndpoint := fmt.Sprintf("http://%v", options[0])

	_ = os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", otlpEndpoint)

	b.ResetTimer()
	ctx := b.Context()
	_, err = StartDefaultMeter(ctx)
	if err != nil {
		b.Fatal(err)
	}

	gauge, err := IntGauge("test_int_gauge", MetricInput{
		Description: "test",
		Unit:        units.Milliseconds,
	})
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		n := float64(rand.IntN(999) + 1)
		gauge.Record(ctx, int64(n))
	}
}
