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

package logger

import (
	"log/slog"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"github.com/stretchr/testify/assert"
)

func getLoggingConfig() config.LoggerConfig {
	serviceName := "test-service"

	cfg := config.NewLoggerConfig()
	cfg.ServiceCfg = serviceName

	return cfg
}

func parseLogToMap(log string) map[string]string {
	// Adjusted regex to allow dots and underscores in the keys
	re := regexp.MustCompile(`([\w.]+)=("[^"]*"|\S+)`)
	matches := re.FindAllStringSubmatch(log, -1)

	logMap := make(map[string]string)
	for _, match := range matches {
		key := match[1]
		value := match[2]

		// Trim quotes from the value if it's quoted
		if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			value = strings.Trim(value, `"`)
		}

		logMap[key] = value
	}
	return logMap
}

type logOutput struct {
	log map[string]string
}

func (lo *logOutput) Write(p []byte) (n int, err error) {
	lo.log = parseLogToMap(string(p))
	return len(p), nil
}

func TestInjectRootAttrs(t *testing.T) {
	_ = os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "k8s.container.name={some-container},k8s.deployment.name={some-deployment},k8s.deployment.uid={some-uid},k8s.namespace.name={some-namespace},k8s.node.name={some-node},k8s.pod.name={some-pod},k8s.pod.uid={some-uid},k8s.replicaset.name={some-replicaset},k8s.replicaset.uid={some-uid},service.instance.id={some-namespace}.{some-pod}.{some-container},service.version={some-version}")
	defer os.Unsetenv("OTEL_RESOURCE_ATTRIBUTES")

	cfg := getLoggingConfig()
	output := &logOutput{}

	handler := slog.NewTextHandler(output, nil)
	log := slog.New(InjectRootAttrs(handler, cfg))
	log.Info("Test log message")

	assert.Contains(t, output.log, config.SERVICE_NAME)
	assert.Equal(t, output.log[config.SERVICE_NAME], cfg.Service())

	assert.Contains(t, output.log, config.SERVICE_VERSION)
	assert.Equal(t, "{some-version}", output.log[config.SERVICE_VERSION])

	assert.Contains(t, output.log, config.SERVICE_INTANCE_ID)
	assert.Equal(t, "{some-namespace}.{some-pod}.{some-container}", output.log[config.SERVICE_INTANCE_ID])

}

func TestReplaceAttributes(t *testing.T) {
	t.Run("Key is time", func(t *testing.T) {
		attr := slog.Attr{
			Key:   "time",
			Value: slog.AnyValue("original time"),
		}
		result := replaceAttributes(nil, attr)

		assert.Equal(t, "time", result.Key, "Expected Key to remain 'time'")

		// Validate the value is updated to the current time in UTC
		now := time.Now().UTC()
		attrTime, ok := result.Value.Any().(time.Time)
		assert.True(t, ok, "Expected Value to be of type time.Time")
		assert.WithinDuration(t, now, attrTime, time.Second, "Expected Value to be close to current time")
	})

	t.Run("Key is msg", func(t *testing.T) {
		attr := slog.Attr{
			Key:   "msg",
			Value: slog.AnyValue("original message"),
		}
		result := replaceAttributes(nil, attr)

		// Check that the key has changed
		assert.Equal(t, config.LOG_MESSAGE_KEY, result.Key, "Expected Key to change to LOG_MESSAGE_KEY")
		assert.Equal(t, "original message", result.Value.Any(), "Expected Value to remain 'original message'")
	})

	t.Run("Key is neither time nor msg", func(t *testing.T) {
		attr := slog.Attr{
			Key:   "other",
			Value: slog.AnyValue("some value"),
		}
		result := replaceAttributes(nil, attr)

		// Check that the attribute is unchanged
		assert.Equal(t, attr.Key, result.Key, "Expected Key to remain 'other'")
		assert.Equal(t, attr.Value.Any(), result.Value.Any(), "Expected Value to remain 'some value'")
	})
}
