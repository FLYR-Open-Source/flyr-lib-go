package pubsub

import (
	"context"

	flyrGoogle "github.com/FlyrInc/flyr-lib-go/google/core"

	"cloud.google.com/go/pubsub"
)

// Interfaces abstracting the Google Pub/Sub API

// Interface abstracting the Google pubsub.Client
type PubSubClient interface {
	Topic(string) PubSubTopic
	Subscription(string) PubSubSubscription
	Close() error
}

// Interface abstracting the Google pubsub.Topic
type PubSubTopic interface {
	Publish(context.Context, []byte, map[string]string) PubSubPublishResult
	Stop()
}

// Interface abstracting the Google pubsub.PublishResult
type PubSubPublishResult interface {
	Get(context.Context) (string, error)
}

// Interface abstracting the Google pubsub.Subscription
type PubSubSubscription interface {
	Receive(context.Context, func(context.Context, flyrGoogle.PubSubMessage)) error
}

// Client Implementation

type clientWrapper struct {
	client *pubsub.Client
}

func (c *clientWrapper) Topic(id string) PubSubTopic {
	return &topicWrapper{topic: c.client.Topic(id)}
}

func (c *clientWrapper) Subscription(id string) PubSubSubscription {
	return &subscriptionWrapper{subscription: c.client.Subscription(id)}
}

func (c *clientWrapper) Close() error {
	return c.client.Close()
}

type topicWrapper struct {
	topic *pubsub.Topic
}

func (t *topicWrapper) Publish(ctx context.Context, data []byte, attributes map[string]string) PubSubPublishResult {
	message := &pubsub.Message{Data: data, Attributes: attributes}
	return &publishResultWrapper{result: t.topic.Publish(ctx, message)}
}

func (t *topicWrapper) Stop() {
	t.topic.Stop()
}

type publishResultWrapper struct {
	result *pubsub.PublishResult
}

func (r *publishResultWrapper) Get(ctx context.Context) (string, error) {
	return r.result.Get(ctx)
}

type subscriptionWrapper struct {
	subscription *pubsub.Subscription
}

func (s *subscriptionWrapper) Receive(ctx context.Context, f func(context.Context, flyrGoogle.PubSubMessage)) error {
	function := func(c context.Context, m *pubsub.Message) {
		f(c, &messageWrapper{message: m})
	}
	return s.subscription.Receive(ctx, function)
}

type messageWrapper struct {
	message *pubsub.Message
}

func (m *messageWrapper) Ack() {
	m.message.Ack()
}

func (m *messageWrapper) Nack() {
	m.message.Nack()
}

func (m *messageWrapper) ID() string {
	return m.message.ID
}

func (m *messageWrapper) Data() []byte {
	return m.message.Data
}

func (m *messageWrapper) Attributes() map[string]string {
	return m.message.Attributes
}
