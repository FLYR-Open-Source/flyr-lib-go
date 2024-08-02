package tracing

import (
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	apmAgentAddr      = "localhost:8126"
	apmReceiverSocket = "/var/run/datadog/apm.socket"
	dogstatsdAddr     = "unix:///var/run/datadog/dsd.socket"
)

func StartTracer(service, version, env, tenant, team string) {
	options := tracerStartOptions(service, version, env, tenant, team)
	tracer.Start(options...)
}

func StopTracer() {
	tracer.Stop()
}

func tracerStartOptions(service, version, env, tenant, team string) []tracer.StartOption {
	return []tracer.StartOption{
		tracer.WithService(service),
		tracer.WithServiceVersion(version),
		tracer.WithEnv(env),
		tracer.WithAgentAddr(apmAgentAddr),
		tracer.WithUDS(apmReceiverSocket),
		tracer.WithTraceEnabled(true),
		tracer.WithLogStartup(false),
		tracer.WithGlobalTag("flyr_tenant", tenant),
		tracer.WithGlobalTag("flyr_team", team),
		tracer.WithRuntimeMetrics(),
		tracer.WithDogstatsdAddress(dogstatsdAddr),
	}
}
