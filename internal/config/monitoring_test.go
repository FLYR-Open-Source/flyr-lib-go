package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonitoringConfig(t *testing.T) {
	cfg := NewMonitoringConfig()

	assert.Equalf(t, "", cfg.Service(), "default Service() return value is not correct")
}

func TestMonitoringConfigWithEnvVars(t *testing.T) {
	en := map[string]string{
		"OTEL_SERVICE_NAME": "test-service",
	}

	cfg := NewMonitoringConfig(withEnvironment(en))

	assert.Equalf(t, "test-service", cfg.Service(), "environment override Service() return value is not correct")
}
