package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonitoringConfig(t *testing.T) {
	cfg := NewMonitoringConfig()

	assert.Equalf(t, "", cfg.Service(), "default Service() return value is not correct")
	assert.Equalf(t, "", cfg.Env(), "default Env() return value is not correct")
	assert.Equalf(t, "", cfg.Version(), "default Version() return value is not correct")
	assert.Equalf(t, "", cfg.Tenant(), "default Tenant() return value is not correct")
	assert.Falsef(t, cfg.TracerEnabled(), "default TracerEnabled() return value is not correct")
}

func TestMonitoringConfigWithEnvVars(t *testing.T) {
	en := map[string]string{
		"OBSERVABILITY_SERVICE":                "test-service",
		"OBSERVABILITY_ENV":                    "test-env",
		"OBSERVABILITY_VERSION":                "test-version",
		"OBSERVABILITY_FLYR_TENANT":            "test-tenant",
		"OBSERVABILITY_TRACER_ENABLED":         "true",
		"OBSERVABILITY_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317",
	}

	cfg := NewLoggerConfig(withEnvironment(en))

	assert.Equalf(t, "test-service", cfg.Service(), "environment override Service() return value is not correct")
	assert.Equalf(t, "test-env", cfg.Env(), "environment override Env() return value is not correct")
	assert.Equalf(t, "test-version", cfg.Version(), "environment override Version() return value is not correct")
	assert.Equalf(t, "test-tenant", cfg.Tenant(), "environment override Tenant() return value is not correct")
	assert.Truef(t, cfg.TracerEnabled(), "environment override TracerEnabled() return value is not correct")
	assert.Equalf(t, "http://localhost:4317", cfg.ExporterEndpoint(), "environment override ExporterEndpoint() return value is not correct")
}

func TestMonitoringConfigWithBadEnvVars(t *testing.T) {
	en := map[string]string{
		"OBSERVABILITY_TRACER_ENABLED": "not-a-bool",
	}

	expectPanicFunc := func() { NewLoggerConfig(withEnvironment(en)) }

	assert.Panicsf(t, expectPanicFunc, "environment override TracerEnabled() should panic")
}
