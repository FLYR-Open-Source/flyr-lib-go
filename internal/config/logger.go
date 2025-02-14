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

type LoggerConfig interface {
	LogLevel() string
	Service() string
}
type Logger struct {
	LogLevelCfg string `env:"LOG_LEVEL" envDefault:"info"`

	Monitoring
}

// LoggerConfig returns the Logger configuration from the environment.
func NewLoggerConfig(opts ...Option) Logger {
	cfg := Logger{}
	if err := envParse(&cfg, opts...); err != nil {
		panic(err)
	}
	return cfg
}

// LogLevel returns the minimum log level for the logger.
// Possible values could be error, warn, info, debug
func (l Logger) LogLevel() string {
	return l.LogLevelCfg
}

// Service returns the service name for application tagging.
func (l Logger) Service() string {
	return l.ServiceCfg
}
