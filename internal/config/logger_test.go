package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerConfig(t *testing.T) {
	cfg := NewLoggerConfig()

	assert.Equalf(t, "", cfg.Service(), "default Service() return value is not correct")
	assert.Equalf(t, "info", cfg.LogLevel(), "default LogLevel() return value is not correct")
}

func TestLoggerConfigWithEnvVars(t *testing.T) {
	en := map[string]string{
		"OTEL_SERVICE_NAME": "test-service",
		"LOG_LEVEL":         "error",
	}

	cfg := NewLoggerConfig(withEnvironment(en))

	assert.Equalf(t, "test-service", cfg.Service(), "default Service() return value is not correct")
	assert.Equalf(t, "error", cfg.LogLevel(), "default LogLevel() return value is not correct")
}
