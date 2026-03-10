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

package pubsub // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/pubsub/v2"

import (
	"context"

	"cloud.google.com/go/pubsub/v2"
	"google.golang.org/api/option"
)

// Option is a functional option for configuring the PubSub client.
type Option func(*options)

type options struct {
	noTracing  bool
	clientOpts []option.ClientOption
	psCfg      *pubsub.ClientConfig
}

func defaultOptions() *options {
	return &options{}
}

// WithClientOptions sets additional Google API client options on the PubSub client.
// These are passed directly to the underlying pubsub.NewClientWithConfig call and are
// useful for configuring credentials, endpoints, or transport settings.
// Multiple calls to WithClientOptions overwrite previous ones; use a single call with
// all desired options if multiple are needed.
func WithClientOptions(opts ...option.ClientOption) Option {
	return func(o *options) {
		o.clientOpts = opts
	}
}

// WithNoTracing disables OpenTelemetry tracing on the PubSub client.
// By default, tracing is enabled when no config is provided via WithConfig.
// When a config is provided via WithConfig, WithNoTracing takes precedence and
// overrides the EnableOpenTelemetryTracing field in that config.
func WithNoTracing() Option {
	return func(o *options) {
		o.noTracing = true
	}
}

// WithConfig sets a custom [pubsub.ClientConfig] on the client.
// When provided, the caller controls all fields in the config, including
// EnableOpenTelemetryTracing. However, if WithNoTracing is also specified,
// it takes precedence and disables tracing regardless of the config value.
// Passing nil resets to the default behaviour: an empty config with tracing enabled.
func WithConfig(cfg *pubsub.ClientConfig) Option {
	return func(o *options) {
		o.psCfg = cfg
	}
}

// NewClient initializes a new GCP PubSub client.
//
// By default, OpenTelemetry tracing is enabled automatically. This behaviour can be
// changed in two ways:
//   - Pass WithNoTracing() to disable tracing entirely.
//   - Pass WithConfig with a custom [pubsub.ClientConfig] to take full control over
//     the config, including the EnableOpenTelemetryTracing field. WithNoTracing still
//     takes precedence over the config if both are provided.
//
// Returns a new *pubsub.Client and an error if the client could not be created.
func NewClient(ctx context.Context, projectID string, opts ...Option) (*pubsub.Client, error) {
	cfg := defaultOptions()

	for _, o := range opts {
		o(cfg)
	}

	if cfg.psCfg == nil {
		cfg.psCfg = &pubsub.ClientConfig{EnableOpenTelemetryTracing: true}
	}

	if cfg.noTracing {
		cfg.psCfg.EnableOpenTelemetryTracing = false
	}

	return pubsub.NewClientWithConfig(ctx, projectID, cfg.psCfg, cfg.clientOpts...)
}
