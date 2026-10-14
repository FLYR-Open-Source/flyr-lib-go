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
	"strings"
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

	// When PUBSUB_EMULATOR_HOST is set, the SDK's emulator hook appends
	// option.WithTelemetryDisabled(), which suppresses the gRPC (otelgrpc)
	// spans these tests assert on. Clear it so results do not depend on the
	// environment (e.g. CI).
	t.Setenv("PUBSUB_EMULATOR_HOST", "")

	srv := pstest.NewServer()
	t.Cleanup(func() { _ = srv.Close() })

	return []option.ClientOption{
		option.WithEndpoint(srv.Addr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}
}

func TestDefaultOptions(t *testing.T) {
	o := defaultOptions()

	require.NotNil(t, o)
	assert.Empty(t, o.clientOpts)
	assert.Nil(t, o.psCfg)
}

func TestOptions(t *testing.T) {
	customCfg := &pubsub.ClientConfig{EnableOpenTelemetryTracing: false}

	tests := []struct {
		name         string
		opts         []Option
		wantClient   int
		wantCfg      *pubsub.ClientConfig
		wantCfgIsNil bool
	}{
		{
			name:         "no options",
			opts:         nil,
			wantClient:   0,
			wantCfgIsNil: true,
		},
		{
			name:         "WithClientOptions stores the options",
			opts:         []Option{WithClientOptions(option.WithoutAuthentication(), option.WithEndpoint("localhost:1"))},
			wantClient:   2,
			wantCfgIsNil: true,
		},
		{
			name: "WithClientOptions overwrites previous calls",
			opts: []Option{
				WithClientOptions(option.WithoutAuthentication(), option.WithEndpoint("localhost:1")),
				WithClientOptions(option.WithoutAuthentication()),
			},
			wantClient:   1,
			wantCfgIsNil: true,
		},
		{
			name:         "WithDisabledGrpcTracing appends telemetry option",
			opts:         []Option{WithDisabledGrpcTracing()},
			wantClient:   1,
			wantCfgIsNil: true,
		},
		{
			name: "WithDisabledGrpcTracing keeps preceding client options",
			opts: []Option{
				WithClientOptions(option.WithoutAuthentication()),
				WithDisabledGrpcTracing(),
			},
			wantClient:   2,
			wantCfgIsNil: true,
		},
		{
			name: "WithClientOptions after WithDisabledGrpcTracing overwrites it",
			opts: []Option{
				WithDisabledGrpcTracing(),
				WithClientOptions(option.WithoutAuthentication()),
			},
			wantClient:   1,
			wantCfgIsNil: true,
		},
		{
			name:       "WithConfig stores the config",
			opts:       []Option{WithConfig(customCfg)},
			wantClient: 0,
			wantCfg:    customCfg,
		},
		{
			name:         "WithConfig nil resets to default",
			opts:         []Option{WithConfig(customCfg), WithConfig(nil)},
			wantClient:   0,
			wantCfgIsNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := defaultOptions()
			for _, opt := range tt.opts {
				opt(o)
			}

			assert.Len(t, o.clientOpts, tt.wantClient)
			if tt.wantCfgIsNil {
				assert.Nil(t, o.psCfg)
			} else {
				assert.Same(t, tt.wantCfg, o.psCfg)
			}
		})
	}
}

func TestNewClient_EmptyProjectID(t *testing.T) {
	client, err := NewClient(context.Background(), "", WithClientOptions(localOptions(t)...))

	require.ErrorIs(t, err, pubsub.ErrEmptyProjectID)
	assert.Nil(t, client)
}

// publishAndCountSpans creates the test topic, publishes one message and
// returns how many library ("<topic> create") and otelgrpc spans were recorded.
func publishAndCountSpans(t *testing.T, ctx context.Context, client *pubsub.Client, recorder *tracetest.SpanRecorder) (libSpans, grpcSpans int) {
	t.Helper()

	topicName := "projects/" + testProject + "/topics/" + testTopic
	_, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{Name: topicName})
	require.NoError(t, err)

	publisher := client.Publisher(topicName)
	t.Cleanup(publisher.Stop)

	_, err = publisher.Publish(ctx, &pubsub.Message{Data: []byte("hello")}).Get(ctx)
	require.NoError(t, err)
	publisher.Stop()

	for _, s := range recorder.Ended() {
		scope := s.InstrumentationScope().Name
		switch {
		case scope == pubsubScope && s.Name() == testTopic+" create":
			libSpans++
		case strings.Contains(scope, "otelgrpc"):
			grpcSpans++
		}
	}

	return libSpans, grpcSpans
}

func TestNewClient_Tracing(t *testing.T) {
	tests := []struct {
		name         string
		extraOpts    []Option
		wantLibSpans bool // spans created by the pubsub library (EnableOpenTelemetryTracing)
		wantGrpcSpan bool // spans created by the gRPC transport layer
	}{
		{
			name:         "tracing is enabled by default",
			wantLibSpans: true,
			wantGrpcSpan: true,
		},
		{
			name:         "gRPC tracing disabled keeps library tracing",
			extraOpts:    []Option{WithDisabledGrpcTracing()},
			wantLibSpans: true,
			wantGrpcSpan: false,
		},
		{
			name:         "custom config enabling tracing",
			extraOpts:    []Option{WithConfig(&pubsub.ClientConfig{EnableOpenTelemetryTracing: true})},
			wantLibSpans: true,
			wantGrpcSpan: true,
		},
		{
			name:         "custom config disabling tracing",
			extraOpts:    []Option{WithConfig(&pubsub.ClientConfig{EnableOpenTelemetryTracing: false})},
			wantLibSpans: false,
			wantGrpcSpan: true,
		},
		{
			name:         "nil config falls back to tracing enabled",
			extraOpts:    []Option{WithConfig(nil)},
			wantLibSpans: true,
			wantGrpcSpan: true,
		},
		{
			name: "all tracing disabled",
			extraOpts: []Option{
				WithConfig(&pubsub.ClientConfig{EnableOpenTelemetryTracing: false}),
				WithDisabledGrpcTracing(),
			},
			wantLibSpans: false,
			wantGrpcSpan: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := installRecorder(t)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			t.Cleanup(cancel)

			// WithClientOptions overwrites, and WithDisabledGrpcTracing appends,
			// so the former has to come first.
			opts := append([]Option{WithClientOptions(localOptions(t)...)}, tt.extraOpts...)
			client, err := NewClient(ctx, testProject, opts...)
			require.NoError(t, err)
			require.NotNil(t, client)
			t.Cleanup(func() { _ = client.Close() })

			assert.Equal(t, testProject, client.Project())

			libSpans, grpcSpans := publishAndCountSpans(t, ctx, client, recorder)

			if tt.wantLibSpans {
				assert.Equal(t, 1, libSpans)
			} else {
				assert.Zero(t, libSpans)
			}
			if tt.wantGrpcSpan {
				assert.Positive(t, grpcSpans)
			} else {
				assert.Zero(t, grpcSpans)
			}
			if !tt.wantLibSpans && !tt.wantGrpcSpan {
				assert.Empty(t, recorder.Ended())
			}
		})
	}
}

// TestNewClient_EmulatorHostDisablesGrpcTracing documents SDK behaviour: with
// PUBSUB_EMULATOR_HOST set, the SDK disables client telemetry, so even default
// options yield no otelgrpc spans (the library span is unaffected). The
// endpoint override from localOptions still wins, so traffic reaches pstest.
func TestNewClient_EmulatorHostDisablesGrpcTracing(t *testing.T) {
	recorder := installRecorder(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	opts := localOptions(t)
	t.Setenv("PUBSUB_EMULATOR_HOST", "localhost:1") // after localOptions, which clears it

	client, err := NewClient(ctx, testProject, WithClientOptions(opts...))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	libSpans, grpcSpans := publishAndCountSpans(t, ctx, client, recorder)

	assert.Equal(t, 1, libSpans)
	assert.Zero(t, grpcSpans)
}
