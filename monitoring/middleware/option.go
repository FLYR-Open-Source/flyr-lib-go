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

package middleware // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/middleware"

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"

	internalConfig "github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"github.com/FLYR-Open-Source/flyr-lib-go/internal/version"
)

const (
	tracerName = "github.com/flyr-open-source/flyr-lib-go"
)

type config struct {
	monitoringCfg internalConfig.Monitoring
	tracer        oteltrace.Tracer
	propagators   propagation.TextMapPropagator
	filters       []filter
}

// Filter is a predicate used to determine whether a given http.request should
// be traced. A Filter must return true if the request should be traced.
type filter func(*http.Request) bool

// configOption applies a configuration option to the trace middleware config.
type configOption interface {
	apply(config)
}

// Adapter functions to implement these interface for configOption
type optionFunc func(config)

func (fn optionFunc) apply(s config) {
	fn(s)
}

// newConfig returns a new config with default values.
func newConfig() config {
	return config{
		filters:       []filter{},
		monitoringCfg: internalConfig.NewMonitoringConfig(),
		tracer:        otel.GetTracerProvider().Tracer(tracerName, oteltrace.WithInstrumentationVersion(version.Version())),
		propagators:   otel.GetTextMapPropagator(),
	}
}
