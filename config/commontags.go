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
	CODE_PATH = "code.filepath"
	CODE_LINE = "code.lineno"
	CODE_FUNC = "code.function"
	CODE_NS   = "code.namespace"
)
