package config

import "github.com/caarlos0/env/v10"

type parseConfig struct {
	environment map[string]string
}

// Option is a function that modifies the parseConfig.
// The only implemented option is WithEnvironment, which is used for testing.
type Option func(*parseConfig)

// envParse is a wrapper around env.Parse that allows for passing in
// an environment map for testing.
func envParse(v interface{}, opts ...Option) error {
	if len(opts) == 0 {
		return env.Parse(v)
	}
	// allow for optional passed in environment, used for testing
	cfg := parseConfig{}
	opts[0](&cfg)
	options := env.Options{
		Environment: cfg.environment,
	}
	return env.ParseWithOptions(v, options)
}
