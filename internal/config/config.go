// MIT License
//
// Copyright (c) 2025 FLYR, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config // import "github.com/FlyrInc/flyr-lib-go/internal/config"

import "github.com/caarlos0/env/v10"

// withEnvironment allows for passing in a map of environment variables
// to be used instead of the actual environment. This is mostly for testing.
func withEnvironment(environ map[string]string) Option {
	return func(c *parseConfig) {
		c.environment = environ
	}
}

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
