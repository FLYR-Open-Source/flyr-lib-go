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
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
)

func setOsEnvironments(testExporter bool, protocol string) {
	os.Setenv("OTEL_SERVICE_NAME", "test-service")
	os.Setenv("OTEL_EXPORTER_OTLP_TEST", strconv.FormatBool(testExporter))
	os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", protocol)
}

func deleteOsEnvironments() {
	os.Unsetenv("OTEL_SERVICE_NAME")
	os.Unsetenv("OTEL_EXPORTER_OTLP_TEST")
	os.Unsetenv("OTEL_EXPORTER_OTLP_PROTOCOL")
}

// TestStartDefaultMeter tests the StartDefaultMeter function.
func TestStartDefaultMeter(t *testing.T) {
	ctx := context.Background()

	t.Run("successful initialization", func(t *testing.T) {
		setOsEnvironments(true, "grpc")
		defer config.ResetMonitoringConfig()
		defer deleteOsEnvironments()
		// Reset the global defaultMeter
		defaultMeter = nil

		// Call the function
		meter, err := StartDefaultMeter(ctx)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, meter)
		assert.Equal(t, defaultMeter, meter)
	})

	t.Run("initializeMeterProvider fails", func(t *testing.T) {
		setOsEnvironments(false, "invalid")
		defer config.ResetMonitoringConfig()
		defer deleteOsEnvironments()

		// Reset the global defaultMeter
		defaultMeter = nil

		// Call the function
		meter, err := StartDefaultMeter(ctx)

		// Assertions
		require.Error(t, err)
		assert.Equal(t, noop.Meter{}, meter)
		assert.Nil(t, defaultMeter)
	})
}
