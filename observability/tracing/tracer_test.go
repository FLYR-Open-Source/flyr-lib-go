package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_tracerStartOptions(t *testing.T) {
	service := "test-service"
	version := "test-version"
	env := "test-env"
	tenant := "test-tenant"
	team := "test-team"

	options := tracerStartOptions(service, version, env, tenant, team)

	// Check that it has correct number of options
	assert.Equal(t, 11, len(options))
}

func TestStartStopTracer(t *testing.T) {
	service := "test-service"
	version := "test-version"
	env := "test-env"
	tenant := "test-tenant"
	team := "test-team"

	StartTracer(service, version, env, tenant, team)
	StopTracer()
}
