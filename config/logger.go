package config // import "github.com/FlyrInc/flyr-lib-go/config"

type LoggerConfig interface {
	LogLevel() string
	Service() string
	Env() string
	Version() string
	Tenant() string
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

// Env returns the environment name for application tagging.
// Possible values could be local, dev, int, qa, testing, traning, prd etc...
func (l Logger) Env() string {
	return l.EnvCfg
}

// Version returns the version name for application tagging.
// Possible values could be v1.2.3, v2023-01-23-0456, etc...
func (l Logger) Version() string {
	return l.VersionCfg
}

// Tenant returns the customer code for application tagging.
func (l Logger) Tenant() string {
	return l.FlyrTenantCfg
}
