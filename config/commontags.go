package config

const (
	// The version of the service
	VERSION_NAME = "service.version"
	// The environment of the service
	ENV_NAME = "deployment.environment.name"
	// The name of the service
	SERVICE_NAME = "service.name"
	// The name of the tenant (multi-tenancy)
	TENANT_NAME = "flyr_tenant"
)

const (
	LOG_MESSAGE_KEY  = "message"
	LOG_ERROR_KEY    = "error"
	LOG_METADATA_KEY = "metadata"
)

const (
	// The file path of the code that generated the log or the span
	CODE_PATH = "code.filepath"
	// The line number of the code that generated the log or the span
	CODE_LINE = "code.lineno"
	// The function name of the code that generated the log or the span
	CODE_FUNC = "code.function"
	// The namespace of the code that generated the log or the span
	CODE_NS = "code.namespace"
)
