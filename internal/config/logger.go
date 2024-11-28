package config // import "github.com/FlyrInc/flyr-lib-go/internal/config"

type LoggerConfig interface {
	LogLevel() string
	Service() string
}
type Logger struct {
	LogLevelCfg string `env:"LOG_LEVEL" envDefault:"info"`

	Monitoring
}

// LoggerConfig returns the Logger configuration from the environment.
func NewLoggerConfig(opts ...Option) Logger {
	cfg := Logger{}
	if err := envParse(&cfg, opts...); err != nil {
		panic(err)
	}
	return cfg
}

// LogLevel returns the minimum log level for the logger.
// Possible values could be error, warn, info, debug
func (l Logger) LogLevel() string {
	return l.LogLevelCfg
}

// Service returns the service name for application tagging.
func (l Logger) Service() string {
	return l.ServiceCfg
}
