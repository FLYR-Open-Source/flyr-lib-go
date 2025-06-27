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

package config // import "github.com/FLYR-Open-Source/flyr-lib-go/internal/config"

import (
	"sync"
	"time"
)

const (
	// defaultMetricsInterval is the default interval at which metrics are exported.
	defaultMetricsInterval = float64(60.0)
)

var (
	monitoringConfigInstance Monitoring
	monitoringConfigOnce     sync.Once
)

// ResetMonitoringConfig resets the singleton state.
// This function is primarily intended for testing purposes.
func ResetMonitoringConfig() {
	monitoringConfigOnce = sync.Once{}
	monitoringConfigInstance = Monitoring{}
}

type MonitoringConfig interface {
	// Generic configuration
	Service() string
	IsTestExporter() bool
	// Traces configuration
	ExporterTracesProtocol() string
	// Metrics configuration
	ExporterMetricsProtocol() string
	MetricsInterval() time.Duration
}

type Monitoring struct {
	// Generic configuration
	ServiceCfg      string `env:"OTEL_SERVICE_NAME"`
	TestExporterCfg bool   `env:"OTEL_EXPORTER_OTLP_TEST"` // Specifies whether the OTLP exporter should be used in test mode.
	// Exporter configuration
	ExporterProtocolCfg string `env:"OTEL_EXPORTER_OTLP_PROTOCOL"` // Specifies the OTLP transport protocol to be used for all telemetry data.
	// Traces configuration
	ExporterTraceProtocolCfg string `env:"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"` // Specifies the OTLP transport protocol to be used for trace data.
	// Metrics configuration
	ExporterMetricsProtocolCfg string  `env:"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"` // Specifies the OTLP transport protocol to be used for metric data.
	MetricsIntervalCfg         float64 `env:"OTEL_METRICS_INTERVAL_SECONDS"`       // Specifies the interval at which metrics are exported.
}

func NewMonitoringConfig(opts ...Option) Monitoring {
	monitoringConfigOnce.Do(func() {
		cfg := Monitoring{}
		if err := envParse(&cfg, opts...); err != nil {
			panic(err)
		}
		monitoringConfigInstance = cfg
	})
	return monitoringConfigInstance
}

// Service returns the service name for application tagging.
func (d Monitoring) Service() string {
	return d.ServiceCfg
}

// IsTestExporter returns whether the OTLP exporter should be used in test mode.
func (d Monitoring) IsTestExporter() bool {
	return d.TestExporterCfg
}

// ExporterTracesProtocol returns the protocol used by the OTLP Trace exporter.
// If both `OTEL_EXPORTER_OTLP_PROTOCOL` and `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL` are present,
// `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL` takes higher precedence.
func (d Monitoring) ExporterTracesProtocol() string {
	if d.ExporterTraceProtocolCfg != "" {
		return d.ExporterTraceProtocolCfg
	}

	// if the trace protocol is not set, use the general exporter protocol
	return d.ExporterProtocolCfg
}

// ExporterMetricsProtocol returns the protocol used by the OTLP Metrics exporter.
// If both `OTEL_EXPORTER_OTLP_PROTOCOL` and `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL` are present,
// `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL` takes higher precedence.
func (d Monitoring) ExporterMetricsProtocol() string {
	if d.ExporterMetricsProtocolCfg != "" {
		return d.ExporterMetricsProtocolCfg
	}

	// if the metrics protocol is not set, use the general exporter protocol
	return d.ExporterProtocolCfg
}

// MetricsInterval returns the interval at which metrics are exported.
// If the metrics interval is not set, it returns the default metrics interval.
// If the metrics interval is set to a negative value, it returns the default metrics interval.
func (d Monitoring) MetricsInterval() time.Duration {
	interval := d.MetricsIntervalCfg
	if interval <= 0.0 {
		interval = defaultMetricsInterval
	}

	return time.Duration(interval * float64(time.Second))
}
