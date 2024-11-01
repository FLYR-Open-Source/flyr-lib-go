package config

type MonitoringConfig interface {
	Service() string
	Env() string
	Version() string
	Tenant() string
	TracerEnabled() bool
	ExporterEndpoint() string
}

type Monitoring struct {
	ServiceCfg          string `env:"OBSERVABILITY_SERVICE"`
	EnvCfg              string `env:"OBSERVABILITY_ENV"`
	VersionCfg          string `env:"OBSERVABILITY_VERSION"`
	FlyrTenantCfg       string `env:"OBSERVABILITY_FLYR_TENANT"`
	EnableTracer        bool   `env:"OBSERVABILITY_TRACER_ENABLED" envDefault:"false"`
	ExporterEndpointCfg string `env:"OBSERVABILITY_EXPORTER_OTLP_ENDPOINT"`
}

func NewMonitoringConfig(opts ...Option) Monitoring {
	cfg := Monitoring{}
	if err := envParse(&cfg, opts...); err != nil {
		panic(err)
	}
	return cfg
}

func (d Monitoring) Service() string {
	return d.ServiceCfg
}

func (d Monitoring) Env() string {
	return d.EnvCfg
}

func (d Monitoring) Version() string {
	return d.VersionCfg
}

func (d Monitoring) Tenant() string {
	return d.FlyrTenantCfg
}

func (d Monitoring) TracerEnabled() bool {
	return d.EnableTracer
}

func (d Monitoring) ExporterEndpoint() string {
	return d.ExporterEndpointCfg
}
