package config // import "github.com/FlyrInc/flyr-lib-go/internal/config"

type MonitoringConfig interface {
	Service() string
}

type Monitoring struct {
	ServiceCfg string `env:"OTEL_SERVICE_NAME"`
}

func NewMonitoringConfig(opts ...Option) Monitoring {
	cfg := Monitoring{}
	if err := envParse(&cfg, opts...); err != nil {
		panic(err)
	}
	return cfg
}

// Service returns the service name for application tagging.
func (d Monitoring) Service() string {
	return d.ServiceCfg
}
