package flyr_pubsub_test

import (
	"context"
	"errors"
	"testing"

	flyrGoogle "github.com/FlyrInc/flyr-lib-go/google"
	flyrPubSub "github.com/FlyrInc/flyr-lib-go/google/pubsub"
	"github.com/stretchr/testify/assert"
)

// Setup

type mockPubSubClient struct{}

func (c *mockPubSubClient) Topic(id string) flyrPubSub.PubSubTopic {
	return &mockTopic{id: id}
}

func (c *mockPubSubClient) Subscription(id string) flyrPubSub.PubSubSubscription {
	return &mockSubscription{id: id}
}

func (c *mockPubSubClient) Close() error {
	return nil
}

type mockTopic struct {
	id string
}

func (t *mockTopic) Publish(ctx context.Context, data []byte, attributes map[string]string) flyrPubSub.PubSubPublishResult {
	failOnGet := false
	if t.id == "error" {
		failOnGet = true
	}
	return &mockPublishResult{failOnGet: failOnGet}
}

func (t *mockTopic) Stop() {}

type mockPublishResult struct {
	failOnGet bool
}

func (r *mockPublishResult) Get(ctx context.Context) (string, error) {
	if r.failOnGet {
		return "", errors.New("test error")
	}
	return "", nil
}

type mockSubscription struct {
	id string
}

func (s *mockSubscription) Receive(ctx context.Context, f func(context.Context, flyrGoogle.PubSubMessage)) error {
	if s.id == "error" {
		return errors.New("test error")
	}
	return nil
}

// Tests

func TestProcessSubMessagesRunsSuccessfully(t *testing.T) {
	ctx := context.Background()

	provider := flyrPubSub.NewGooglePubSubProvider(&mockPubSubClient{})

	processingFunction := func(c context.Context, m flyrGoogle.PubSubMessage) {}

	err := provider.ProcessSubMessages(ctx, "", processingFunction)

	assert.NoError(t, err)
}

func TestProcessSubMessagesReturnsErrorOnMissingClient(t *testing.T) {
	ctx := context.Background()

	provider := flyrPubSub.GooglePubSubProvider{}

	processingFunction := func(c context.Context, m flyrGoogle.PubSubMessage) {}

	err := provider.ProcessSubMessages(ctx, "", processingFunction)

	assert.Error(t, err)
}

func TestProcessSubMessagesReturnsErrorOnProcessingFailure(t *testing.T) {
	ctx := context.Background()

	provider := flyrPubSub.NewGooglePubSubProvider(&mockPubSubClient{})

	processingFunction := func(c context.Context, m flyrGoogle.PubSubMessage) {}

	err := provider.ProcessSubMessages(ctx, "error", processingFunction)

	assert.Error(t, err)
}

func TestSendPubSubMessageSendsSuccessfully(t *testing.T) {
	ctx := context.Background()

	provider := flyrPubSub.NewGooglePubSubProvider(&mockPubSubClient{})

	err := provider.SendPubSubMessage(ctx, "", []byte("test"), nil)

	assert.NoError(t, err)
}

func TestSendPubSubMessageReturnsErrorOnMissingClient(t *testing.T) {
	ctx := context.Background()

	provider := flyrPubSub.GooglePubSubProvider{}

	err := provider.SendPubSubMessage(ctx, "", []byte("test"), nil)

	assert.Error(t, err)
}

func TestSendPubSubMessageReturnsErrorOnSendFailure(t *testing.T) {
	ctx := context.Background()

	provider := flyrPubSub.GooglePubSubProvider{}

	err := provider.SendPubSubMessage(ctx, "error", []byte("test"), nil)

	assert.Error(t, err)
}
