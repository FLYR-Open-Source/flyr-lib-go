package config // import "github.com/FlyrInc/flyr-lib-go/internal/config"

type MonitoringConfig interface {
	Service() string
	ExporterProtocol() string
}

type Monitoring struct {
	ServiceCfg          string `env:"OTEL_SERVICE_NAME"`
	ExporterProtocolCfg string `env:"OTEL_EXPORTER_OTLP_PROTOCOL"` // http or grpc
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

// ExporterProtocol returns the protocol used by the OTLP exporter.
func (d Monitoring) ExporterProtocol() string {
	return d.ExporterProtocolCfg
}
