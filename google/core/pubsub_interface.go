package core

import (
	"context"
)

// The key used to store and retrieve a Pub/Sub provider from a context.
const PubSubProviderKey = "pubsubProvider"

// Interface abstracting the FLYR Google Pub/Sub provider.
type PubSubProvider interface {
	ProcessSubMessages(ctx context.Context, subscriptionName string, f func(c context.Context, m PubSubMessage)) error
	SendPubSubMessage(ctx context.Context, topicName string, message []byte, attributes map[string]string) error
}

// Interface abstracting the message passed to PubSubProvider.ProcessSubMessages().
type PubSubMessage interface {
	Ack()
	Nack()
	ID() string
	Data() []byte
	Attributes() map[string]string
}
