package config // import "github.com/FlyrInc/flyr-lib-go/internal/config"

type MonitoringConfig interface {
	Service() string
	ExporterTracesProtocol() string
}

type Monitoring struct {
	ServiceCfg               string `env:"OTEL_SERVICE_NAME"`
	ExporterProtocolCfg      string `env:"OTEL_EXPORTER_OTLP_PROTOCOL"`        // Specifies the OTLP transport protocol to be used for all telemetry data.
	ExporterTraceProtocolCfg string `env:"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"` // Specifies the OTLP transport protocol to be used for trace data.
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

// ExporterTracesProtocol returns the protocol used by the OTLP Trace exporter.
func (d Monitoring) ExporterTracesProtocol() string {
	if d.ExporterTraceProtocolCfg != "" {
		return d.ExporterTraceProtocolCfg
	}

	// if the trace protocol is not set, use the general exporter protocol
	return d.ExporterProtocolCfg
}
