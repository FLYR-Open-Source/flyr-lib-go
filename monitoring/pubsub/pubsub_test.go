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
	"testing"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"cloud.google.com/go/pubsub/v2/pstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	testProject = "test-project"
	testTopic   = "test-topic"
	pubsubScope = "cloud.google.com/go/pubsub/v2"
)

// installRecorder registers an in-memory tracer provider as the OTel global
// and restores the previous one when the test finishes.
func installRecorder(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()

	previous := otel.GetTracerProvider()
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(tp)

	t.Cleanup(func() {
		otel.SetTracerProvider(previous)
		_ = tp.Shutdown(context.Background())
	})

	return recorder
}

// localOptions points the client at an in-process fake Pub/Sub server.
func localOptions(t *testing.T) []option.ClientOption {
	t.Helper()

	srv := pstest.NewServer()
	t.Cleanup(func() { _ = srv.Close() })

	return []option.ClientOption{
		option.WithEndpoint(srv.Addr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}
}

func TestNewClient_EmptyProjectID(t *testing.T) {
	client, err := NewClient(context.Background(), "", nil, localOptions(t)...)

	require.ErrorIs(t, err, pubsub.ErrEmptyProjectID)
	assert.Nil(t, client)
}

func TestNewClient_ForcesTracingOnConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  *pubsub.ClientConfig
	}{
		{name: "nil config", cfg: nil},
		{name: "empty config", cfg: &pubsub.ClientConfig{}},
		{name: "config explicitly disabling tracing", cfg: &pubsub.ClientConfig{EnableOpenTelemetryTracing: false}},
		{name: "config already enabling tracing", cfg: &pubsub.ClientConfig{EnableOpenTelemetryTracing: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := installRecorder(t)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			t.Cleanup(cancel)

			client, err := NewClient(ctx, testProject, tt.cfg, localOptions(t)...)
			require.NoError(t, err)
			require.NotNil(t, client)
			t.Cleanup(func() { _ = client.Close() })

			// A caller-provided config is mutated in place.
			if tt.cfg != nil {
				assert.True(t, tt.cfg.EnableOpenTelemetryTracing)
			}
			assert.Equal(t, testProject, client.Project())

			topicName := "projects/" + testProject + "/topics/" + testTopic
			_, err = client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{Name: topicName})
			require.NoError(t, err)

			publisher := client.Publisher(topicName)
			t.Cleanup(publisher.Stop)

			_, err = publisher.Publish(ctx, &pubsub.Message{Data: []byte("hello")}).Get(ctx)
			require.NoError(t, err)
			publisher.Stop()

			var createSpans int
			for _, s := range recorder.Ended() {
				if s.InstrumentationScope().Name == pubsubScope && s.Name() == testTopic+" create" {
					createSpans++
				}
			}
			assert.Equal(t, 1, createSpans)
		})
	}
}
