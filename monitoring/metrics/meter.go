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

package metrics

import (
	"context"
	"time"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var (
	meter metric.Meter
)

// GetDefaultMeter initializes and starts the default OpenTelemetry Meter.
//
// This function checks if custom metrics are enabled in the provided configuration. If they are enabled,
// it creates a new Meter by using the default MeterProvider. It also validates that the
// meter name (the service name in that case) is set in the configuration. If the meter name is not provided,
// a noop Meter is initialised as a default.
//
// The function also sets the global default Meter to be used for custom metrics in the
// application. If custom metrics are not enabled, it returns a noop Meter.
//
// It returns the created Meter and an error if any occurred.
//
// For learning more about the Otel Metrics Data Model, please reference to https://opentelemetry.io/docs/specs/otel/metrics/data-model
func StartDefaultMeter(ctx context.Context) (metric.Meter, error) {
	if meter != nil {
		return meter, nil
	}

	cfg := config.NewMonitoringConfig()

	if cfg.Service() == "" {
		meter = noop.Meter{}
	}

	err := initializeMeterProvider(ctx, cfg, 60*time.Second) // send metrics every 60 seconds
	if err != nil {
		meter = noop.Meter{}
	} else {
		mt := otel.GetMeterProvider()
		meter = mt.Meter(cfg.Service())
	}

	return meter, nil
}
