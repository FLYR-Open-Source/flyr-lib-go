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

package pubsub // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/pubsub"

import (
	"context"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

// NewClient initializes a new GCP PubSub client with OpenTelemetry tracing enabled.
//
// This function creates and returns a *pubsub.Client configured to enable the OpenTelemetry
// feature for both publishing and subscribing to PubSub messages. When enabled, the client
// will automatically trace all PubSub operations using OpenTelemetry.
//
// Returns a new *pubsub.Client with OpenTelemetry tracing configured and an error if any.
func NewClient(ctx context.Context, projectID string, config *pubsub.ClientConfig, opts ...option.ClientOption) (*pubsub.Client, error) {
	if config == nil {
		config = &pubsub.ClientConfig{}
	}

	config.EnableOpenTelemetryTracing = true
	return pubsub.NewClientWithConfig(ctx, projectID, config, opts...)
}
