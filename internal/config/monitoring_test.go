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

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMonitoringConfig(t *testing.T) {
	tests := []struct {
		name                    string
		variables               map[string]string
		expectedService         string
		expectedTraceExporter   string
		expectedMetricsExporter string
		expectedTestExporter    bool
	}{
		{
			name:                  "with empty values",
			variables:             map[string]string{},
			expectedTraceExporter: "",
		},
		{
			name: "with service",
			variables: map[string]string{
				"OTEL_SERVICE_NAME": "my-service",
			},
			expectedService:         "my-service",
			expectedTraceExporter:   "",
			expectedMetricsExporter: "",
		},
		{
			name: "with global trace exporter protocol",
			variables: map[string]string{
				"OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
			},
			expectedTraceExporter:   "grpc",
			expectedMetricsExporter: "grpc",
		},
		{
			name: "with custom trace exporter protocol",
			variables: map[string]string{
				"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL": "http/protobuf",
			},
			expectedTraceExporter: "http/protobuf",
		},
		{
			name: "custom trace exporter protocol must take precedence over global trace exporter protocol",
			variables: map[string]string{
				"OTEL_EXPORTER_OTLP_PROTOCOL":        "grpc",
				"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL": "http/protobuf",
			},
			expectedTraceExporter:   "http/protobuf",
			expectedMetricsExporter: "grpc",
		},
		{
			name: "with custom metrics exporter protocol",
			variables: map[string]string{
				"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL": "http/protobuf",
			},
			expectedMetricsExporter: "http/protobuf",
		},
		{
			name: "custom metrics exporter protocol must take precedence over global trace exporter protocol",
			variables: map[string]string{
				"OTEL_EXPORTER_OTLP_PROTOCOL":         "grpc",
				"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL": "http/protobuf",
			},
			expectedTraceExporter:   "grpc",
			expectedMetricsExporter: "http/protobuf",
		},
		{
			name: "test flag is enabled",
			variables: map[string]string{
				"OTEL_EXPORTER_OTLP_TEST": "true",
			},
			expectedTestExporter: true,
		},
		{
			name: "test flag is disabled",
			variables: map[string]string{
				"OTEL_EXPORTER_OTLP_TEST": "false",
			},
			expectedTestExporter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the singleton state before each test
			ResetMonitoringConfig()

			cfg := NewMonitoringConfig(withEnvironment(tt.variables))

			assert.Equalf(t, tt.expectedService, cfg.Service(), "Service() return value is not correct")
			assert.Equalf(t, tt.expectedTraceExporter, cfg.ExporterTracesProtocol(), "ExporterTracesProtocol() return value is not correct")
			assert.Equalf(t, tt.expectedMetricsExporter, cfg.ExporterMetricsProtocol(), "ExporterMetricsProtocol() return value is not correct")
			assert.Equalf(t, tt.expectedTestExporter, cfg.IsTestExporter(), "IsTestExporter() return value is not correct")
		})
	}
}

func TestMetricsInterval(t *testing.T) {
	tests := []struct {
		name             string
		variables        map[string]string
		expectedInterval time.Duration
	}{
		{
			name: "with default interval",
			variables: map[string]string{
				"OTEL_METRICS_INTERVAL_SECONDS": "0",
			},
			expectedInterval: 60 * time.Second,
		},
		{
			name: "with default interval",
			variables: map[string]string{
				"OTEL_METRICS_INTERVAL_SECONDS": "60",
			},
			expectedInterval: 60 * time.Second,
		},
		{
			name: "with custom interval",
			variables: map[string]string{
				"OTEL_METRICS_INTERVAL_SECONDS": "10",
			},
			expectedInterval: 10 * time.Second,
		},
		{
			name: "with negative interval",
			variables: map[string]string{
				"OTEL_METRICS_INTERVAL_SECONDS": "-1",
			},
			expectedInterval: 60 * time.Second,
		},
		{
			name: "with zero interval",
			variables: map[string]string{
				"OTEL_METRICS_INTERVAL_SECONDS": "0.1",
			},
			expectedInterval: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the singleton state before each test
			ResetMonitoringConfig()

			cfg := NewMonitoringConfig(withEnvironment(tt.variables))
			assert.Equalf(t, tt.expectedInterval, cfg.MetricsInterval(), "MetricsInterval() return value is not correct")
		})
	}
}
