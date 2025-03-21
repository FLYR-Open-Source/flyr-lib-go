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

type MonitoringConfig interface {
	Service() string
	ExporterTracesProtocol() string
	ExporterMetricsProtocol() string
	IsTestExporter() bool
}

type Monitoring struct {
	ServiceCfg                 string `env:"OTEL_SERVICE_NAME"`
	ExporterProtocolCfg        string `env:"OTEL_EXPORTER_OTLP_PROTOCOL"`         // Specifies the OTLP transport protocol to be used for all telemetry data.
	ExporterTraceProtocolCfg   string `env:"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"`  // Specifies the OTLP transport protocol to be used for trace data.
	ExporterMetricsProtocolCfg string `env:"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"` // Specifies the OTLP transport protocol to be used for metric data.
	TestExporterCfg            bool   `env:"OTEL_EXPORTER_OTLP_TEST"`             // Specifies whether the OTLP exporter should be used in test mode.
}

func NewMonitoringConfig(opts ...Option) Monitoring {
	cfg := Monitoring{}
	if err := envParse(&cfg, opts...); err != nil {
		panic(err)
	}
	return cfg
}

// Service returns the service name for application tagging.
func (d Monitoring) Service() string {
	return d.ServiceCfg
}

// ExporterTracesProtocol returns the protocol used by the OTLP Trace exporter.
func (d Monitoring) ExporterTracesProtocol() string {
	if d.ExporterTraceProtocolCfg != "" {
		return d.ExporterTraceProtocolCfg
	}

	// if the trace protocol is not set, use the general exporter protocol
	return d.ExporterProtocolCfg
}

// ExporterMetricsProtocol returns the protocol used by the OTLP Metrics exporter.
func (d Monitoring) ExporterMetricsProtocol() string {
	if d.ExporterMetricsProtocolCfg != "" {
		return d.ExporterMetricsProtocolCfg
	}

	// if the metrics protocol is not set, use the general exporter protocol
	return d.ExporterProtocolCfg
}

// IsTestExporter returns whether the OTLP exporter should be used in test mode.
func (d Monitoring) IsTestExporter() bool {
	return d.TestExporterCfg
}
