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

// Service returns the service name for application tagging.
func (d Monitoring) Service() string {
	return d.ServiceCfg
}

// Env returns the environment name for application tagging.
// Possible values could be local, dev, int, qa, testing, traning, prd etc...
func (d Monitoring) Env() string {
	return d.EnvCfg
}

// Version returns the version name for application tagging.
// Possible values could be v1.2.3, v2023-01-23-0456, etc...
func (d Monitoring) Version() string {
	return d.VersionCfg
}

// Tenant returns the customer code for application tagging.
func (d Monitoring) Tenant() string {
	return d.FlyrTenantCfg
}

// TracerEnabled returns the whether the tracer is enabled or not.
func (d Monitoring) TracerEnabled() bool {
	return d.EnableTracer
}

// ExporterEndpoint returns the exporter endpoint for the monitoring.
func (d Monitoring) ExporterEndpoint() string {
	return d.ExporterEndpointCfg
}
