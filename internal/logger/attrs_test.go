package logger

import (
	"log/slog"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/FlyrInc/flyr-lib-go/config"
	"github.com/stretchr/testify/assert"
)

func getLoggingConfig() config.LoggerConfig {
	serviceName := "test-service"
	env := "test-env"
	flyrTenant := "test-tenant"
	version := "test-version"

	cfg := config.NewLoggerConfig()
	cfg.ServiceCfg = serviceName
	cfg.EnvCfg = env
	cfg.FlyrTenantCfg = flyrTenant
	cfg.VersionCfg = version

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
	cfg := getLoggingConfig()
	output := &logOutput{}

	handler := slog.NewTextHandler(output, nil)
	log := slog.New(InjectRootAttrs(handler, cfg))
	log.Info("Test log message")

	assert.Contains(t, output.log, config.ENV_NAME)
	assert.Equal(t, output.log[config.ENV_NAME], cfg.Env())

	assert.Contains(t, output.log, config.ENV_NAME)
	assert.Equal(t, output.log[config.SERVICE_NAME], cfg.Service())

	assert.Contains(t, output.log, config.VERSION_NAME)
	assert.Equal(t, output.log[config.VERSION_NAME], cfg.Version())

	assert.Contains(t, output.log, config.TENANT_NAME)
	assert.Equal(t, output.log[config.TENANT_NAME], cfg.Tenant())
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
