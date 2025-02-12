package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonitoringConfig(t *testing.T) {
	tests := []struct {
		name                  string
		variables             map[string]string
		expectedService       string
		expectedTraceExporter string
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
			expectedService:       "my-service",
			expectedTraceExporter: "",
		},
		{
			name: "with global trace exporter protocol",
			variables: map[string]string{
				"OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
			},
			expectedTraceExporter: "grpc",
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
			expectedTraceExporter: "http/protobuf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewMonitoringConfig(withEnvironment(tt.variables))

			assert.Equalf(t, tt.expectedService, cfg.Service(), "Service() return value is not correct")
			assert.Equalf(t, tt.expectedTraceExporter, cfg.ExporterTracesProtocol(), "ExporterTracesProtocol() return value is not correct")
		})
	}
}
