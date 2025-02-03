// Package flyr-lib-go/google/pubsub contains functionality for interacting with Google Pub/Sub.
// The primary mechanism used for this is the GooglePubSubProvider which implements the interface defined in flyr-lib-go/google/core.
package pubsub

import (
	"context"
	"errors"

	flyrGoogle "github.com/FlyrInc/flyr-lib-go/google/core"
)

// A GooglePubSubProvider serves as the primary mechanism for interacting (e.g., receiving and sending messages) with Google Pub/Sub.
// GooglePubSubProvider is safe for concurrent use by multiple goroutines.
type GooglePubSubProvider struct {
	client PubSubClient
}

// NewGooglePubSubProvider creates a new GooglePubSubProvider with the provided client
func NewGooglePubSubProvider(client PubSubClient) *GooglePubSubProvider {
	return &GooglePubSubProvider{
		client: client,
	}
}

// ProcessSubMessages uses the provided function to process messages received by the specified subscription.
// Processing continues until the context is closed or a non-retryable error is encountered.
// The provided function must explicitly call the Ack() and/or Nack() methods on the PubSubMessage in order to send or refuse acknowledgement.
func (provider *GooglePubSubProvider) ProcessSubMessages(ctx context.Context, subscriptionName string, f func(c context.Context, m flyrGoogle.PubSubMessage)) error {
	if provider.client == nil {
		return errors.New("the GooglePubSubProvider does not have a valid client")
	}

	return provider.client.Subscription(subscriptionName).Receive(ctx, f)
}

// SendPubSubMessage sends a message to the specified topic. Any attributes provided will be attached to the message.
func (provider *GooglePubSubProvider) SendPubSubMessage(ctx context.Context, topicName string, message []byte, attributes map[string]string) error {
	if provider.client == nil {
		return errors.New("the GooglePubSubProvider does not have a valid client")
	}

	topic := provider.client.Topic(topicName)

	result := topic.Publish(ctx, message, attributes)

	// Errors can result from failing to publish or from a missing resource (i.e., topic)
	_, err := result.Get(ctx)
	if err != nil {
		return err
	}

	topic.Stop()

	return nil
}
