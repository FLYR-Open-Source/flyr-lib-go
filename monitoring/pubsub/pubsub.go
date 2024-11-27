package pubsub // import "github.com/FlyrInc/flyr-lib-go/monitoring/pubsub"

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
