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

package pubsub

import (
	"context"
	"testing"
	"time"

	gpubsub "cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_ReturnsWorkingInProcessServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv, client, err := NewClient(ctx)
	require.NoError(t, err)
	require.NotNil(t, srv)
	require.NotNil(t, client)
	t.Cleanup(func() {
		assert.NoError(t, client.Close())
		assert.NoError(t, srv.Close())
	})

	assert.Equal(t, "test-project", client.Project())
	assert.NotEmpty(t, srv.Addr)

	// The client must talk to the in-process pstest server, not real GCP.
	topicName := "projects/test-project/topics/test-topic"
	topic, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{Name: topicName})
	require.NoError(t, err)
	assert.Equal(t, topicName, topic.GetName())

	publisher := client.Publisher(topicName)
	t.Cleanup(publisher.Stop)

	msgID, err := publisher.Publish(ctx, &gpubsub.Message{Data: []byte("hello")}).Get(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, msgID)

	msgs := srv.Messages()
	require.Len(t, msgs, 1)
	assert.Equal(t, "hello", string(msgs[0].Data))
}
