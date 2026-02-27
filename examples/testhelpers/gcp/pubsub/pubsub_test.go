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

package grpc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/pubsub/v2"
	pubsubpb "cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"

	testhelpers "github.com/FLYR-Open-Source/flyr-lib-go/pkg/testhelpers/gcp/pubsub"
)

func TestTopic(t *testing.T) {
	// Setup
	ctx := context.Background()
	topicID := "my-topic-id"
	projectID := "test-project"
	expectedMessageID := "m0"

	svc, client, err := testhelpers.NewClient(ctx)
	if err != nil {
		t.Fatalf("pubsub.NewClient: %v", err)
		return
	}
	defer svc.Close()
	defer client.Close()

	topicName := fmt.Sprintf("projects/%s/topics/%s", projectID, topicID)
	topic, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
		Name: topicName,
	})
	if err != nil {
		t.Fatalf("TopicAdminClient.CreateTopic: %v", err)
		return
	}

	// Test
	t.Run("Publish", func(t *testing.T) {
		// Setup
		msg := &pubsub.Message{
			Data: []byte("hello"),
		}

		publisher := client.Publisher(topic.GetName())
		defer publisher.Stop()

		publishResult := publisher.Publish(ctx, msg)
		res, err := publishResult.Get(ctx)
		if err != nil {
			t.Fatalf("topic.Publish: %v", err)
			return
		}

		if res != expectedMessageID {
			t.Fatalf("expected message id: %s, got: %s", expectedMessageID, res)
		}
	})
}

func TestSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := "my-subscription-id"
	topicID := "my-topic-id"
	projectID := "test-project"

	svc, client, err := testhelpers.NewClient(ctx)
	if err != nil {
		t.Fatalf("pubsub.NewClient: %v", err)
		return
	}
	defer svc.Close()
	defer client.Close()

	topicName := fmt.Sprintf("projects/%s/topics/%s", projectID, topicID)
	topic, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
		Name: topicName,
	})
	if err != nil {
		t.Fatalf("TopicAdminClient.CreateTopic: %v", err)
		return
	}

	subscriptionName := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subscriptionID)
	sub, err := client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Name:  subscriptionName,
		Topic: topic.GetName(),
	})
	if err != nil {
		t.Fatalf("SubscriptionAdminClient.CreateSubscription: %v", err)
	}

	// Publish a message
	msg := &pubsub.Message{
		Data: []byte("hello"),
	}

	publisher := client.Publisher(topic.GetName())
	defer publisher.Stop()

	publisher.Publish(ctx, msg)

	// Receive messages
	cctx, cancel := context.WithTimeout(ctx, 1*time.Millisecond) // add a small timeout to avoid blocking
	defer cancel()

	subscriber := client.Subscriber(sub.GetName())
	err = subscriber.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack() // Acknowledge the message

		if string(msg.Data) != "hello" {
			t.Fatalf("expected message data: %s, got: %s", "hello", string(msg.Data))
			return
		}
	})
	if err != nil {
		t.Fatalf("sub.Receive: %v", err)
		return
	}
}
