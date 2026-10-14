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

package tracer

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	collector "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"github.com/FLYR-Open-Source/flyr-lib-go/internal/version"
)

// restoreGlobals restores the global tracer provider, the text map propagator and the
// default tracer when the test ends. It also resets the monitoring config singleton.
func restoreGlobals(t *testing.T) {
	t.Helper()

	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()
	previousTracer := defaultTracer
	config.ResetMonitoringConfig()

	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
		defaultTracer = previousTracer
		config.ResetMonitoringConfig()
	})
}

// fakeTraceServer is an in-process OTLP/gRPC trace collector.
type fakeTraceServer struct {
	collector.UnimplementedTraceServiceServer

	mu       sync.Mutex
	received []*collector.ExportTraceServiceRequest
}

func (s *fakeTraceServer) Export(_ context.Context, req *collector.ExportTraceServiceRequest) (*collector.ExportTraceServiceResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.received = append(s.received, req)
	return &collector.ExportTraceServiceResponse{}, nil
}

func (s *fakeTraceServer) requests() []*collector.ExportTraceServiceRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*collector.ExportTraceServiceRequest(nil), s.received...)
}

func startFakeTraceServer(t *testing.T) (*fakeTraceServer, string) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	fake := &fakeTraceServer{}
	srv := grpc.NewServer()
	collector.RegisterTraceServiceServer(srv, fake)

	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	return fake, "http://" + lis.Addr().String()
}

func TestInitializeTracerProvider_Branches(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.Monitoring
		env         map[string]string
		expectedErr error
		// expectErr is set when an error is expected that is not a package sentinel error.
		expectErr bool
		// expectProvider is the check applied to the global provider afterwards.
		expectProvider func(t *testing.T, provider oteltrace.TracerProvider, original oteltrace.TracerProvider)
	}{
		{
			name: "empty service name installs the noop provider",
			cfg:  config.Monitoring{ExporterProtocolCfg: "grpc"},
			expectProvider: func(t *testing.T, provider, _ oteltrace.TracerProvider) {
				assert.IsType(t, noop.TracerProvider{}, provider)
			},
		},
		{
			name:        "unsupported protocol fails and leaves the global provider untouched",
			cfg:         config.Monitoring{ServiceCfg: "svc", ExporterProtocolCfg: "carrier-pigeon"},
			expectedErr: ErrExporterClientNotSupported,
			expectProvider: func(t *testing.T, provider, original oteltrace.TracerProvider) {
				assert.Same(t, original, provider)
			},
		},
		{
			name:        "traces protocol takes precedence over the general protocol",
			cfg:         config.Monitoring{ServiceCfg: "svc", ExporterProtocolCfg: "grpc", ExporterTraceProtocolCfg: "nope"},
			expectedErr: ErrExporterClientNotSupported,
			expectProvider: func(t *testing.T, provider, original oteltrace.TracerProvider) {
				assert.Same(t, original, provider)
			},
		},
		{
			name:      "malformed resource attributes fail and leave the global provider untouched",
			cfg:       config.Monitoring{ServiceCfg: "svc", TestExporterCfg: true},
			env:       map[string]string{"OTEL_RESOURCE_ATTRIBUTES": "missing-equals-sign"},
			expectErr: true,
			expectProvider: func(t *testing.T, provider, original oteltrace.TracerProvider) {
				assert.Same(t, original, provider)
			},
		},
		{
			name: "valid config installs the sdk provider",
			cfg:  config.Monitoring{ServiceCfg: "svc", TestExporterCfg: true},
			expectProvider: func(t *testing.T, provider, _ oteltrace.TracerProvider) {
				sdkProvider, ok := provider.(*sdktrace.TracerProvider)
				require.True(t, ok, "unexpected provider %T", provider)
				require.NoError(t, sdkProvider.Shutdown(context.Background()))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restoreGlobals(t)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			original := sdktrace.NewTracerProvider()
			otel.SetTracerProvider(original)

			err := initializeTracerProvider(context.Background(), tt.cfg)

			switch {
			case tt.expectedErr != nil:
				require.ErrorIs(t, err, tt.expectedErr)
			case tt.expectErr:
				require.Error(t, err)
			default:
				require.NoError(t, err)
			}
			tt.expectProvider(t, otel.GetTracerProvider(), original)
		})
	}
}

func TestInitializeTracerProvider_SetsCompositePropagator(t *testing.T) {
	restoreGlobals(t)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

	cfg := config.Monitoring{ServiceCfg: "svc", TestExporterCfg: true}
	require.NoError(t, initializeTracerProvider(context.Background(), cfg))
	t.Cleanup(func() {
		//nolint:errcheck
		ShutdownTracerProvider(context.Background())
	})

	assert.ElementsMatch(t, []string{"traceparent", "tracestate", "baggage"}, otel.GetTextMapPropagator().Fields())
}

func TestStartDefaultTracer(t *testing.T) {
	t.Run("empty service name starts a noop tracer", func(t *testing.T) {
		restoreGlobals(t)
		t.Setenv("OTEL_SERVICE_NAME", "")

		require.NoError(t, StartDefaultTracer(context.Background()))

		assert.Equal(t, noop.Tracer{}, defaultTracer.tracer)

		ctx, span := StartSpan(context.Background(), "noop-span", oteltrace.SpanKindInternal)
		defer span.End()
		assert.False(t, span.IsRecording())
		assert.False(t, oteltrace.SpanContextFromContext(ctx).IsValid())
	})

	t.Run("unsupported protocol returns the error and keeps the default tracer", func(t *testing.T) {
		restoreGlobals(t)
		t.Setenv("OTEL_SERVICE_NAME", "tracer-test")
		t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "carrier-pigeon")
		t.Setenv("OTEL_EXPORTER_OTLP_TEST", "false")
		before := defaultTracer

		err := StartDefaultTracer(context.Background())

		require.ErrorIs(t, err, ErrExporterClientNotSupported)
		assert.Same(t, before, defaultTracer)
	})

	t.Run("exports spans to the OTLP endpoint on shutdown", func(t *testing.T) {
		restoreGlobals(t)
		fake, endpoint := startFakeTraceServer(t)
		t.Setenv("OTEL_SERVICE_NAME", "tracer-e2e-service")
		t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
		t.Setenv("OTEL_EXPORTER_OTLP_TEST", "false")
		t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", endpoint)

		ctx := context.Background()
		require.NoError(t, StartDefaultTracer(ctx))

		spanCtx, span := StartSpan(ctx, "e2e-span", oteltrace.SpanKindServer)
		assert.True(t, span.IsRecording())
		assert.True(t, GetSpanFromContext(spanCtx).SpanContext().IsValid())
		span.End()

		// the batch processor only flushes on its schedule or on shutdown
		require.NoError(t, ShutdownTracerProvider(ctx))

		var spans []*tracepb.Span
		resourceAttrs := map[string]string{}
		scopeNames := map[string]string{}
		for _, req := range fake.requests() {
			for _, rs := range req.GetResourceSpans() {
				for _, kv := range rs.GetResource().GetAttributes() {
					resourceAttrs[kv.GetKey()] = kv.GetValue().GetStringValue()
				}
				for _, ss := range rs.GetScopeSpans() {
					scopeNames[ss.GetScope().GetName()] = ss.GetScope().GetVersion()
					spans = append(spans, ss.GetSpans()...)
				}
			}
		}

		require.Len(t, spans, 1)
		assert.Equal(t, "e2e-span", spans[0].GetName())
		assert.Equal(t, tracepb.Span_SPAN_KIND_SERVER, spans[0].GetKind())
		assert.Equal(t, "tracer-e2e-service", resourceAttrs["service.name"])
		assert.Equal(t, "grpc", resourceAttrs[config.EXPORTER_PROTOCOL])
		assert.Equal(t, map[string]string{"tracer-e2e-service": version.Version()}, scopeNames)

		callerAttrs := map[string]bool{}
		for _, kv := range spans[0].GetAttributes() {
			callerAttrs[kv.GetKey()] = true
		}
		for _, key := range []string{config.FILE_PATH, config.LINE_NUMBER, config.FUNCTION_NAME, config.FUNCTION_PACKAGE_NAME} {
			assert.True(t, callerAttrs[key], "missing caller attribute %q", key)
		}
	})
}

func TestShutdownTracerProvider(t *testing.T) {
	t.Run("noop provider is a no-op", func(t *testing.T) {
		restoreGlobals(t)
		otel.SetTracerProvider(noop.NewTracerProvider())

		require.NoError(t, ShutdownTracerProvider(context.Background()))
	})

	t.Run("sdk provider is shut down", func(t *testing.T) {
		restoreGlobals(t)
		exporter := tracetest.NewInMemoryExporter()
		tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
		otel.SetTracerProvider(tp)

		_, span := tp.Tracer("shutdown-test").Start(context.Background(), "before-shutdown")
		span.End()
		require.Len(t, exporter.GetSpans(), 1)

		require.NoError(t, ShutdownTracerProvider(context.Background()))

		// after shutdown the provider hands out non-recording spans
		_, span = tp.Tracer("shutdown-test").Start(context.Background(), "after-shutdown")
		assert.False(t, span.IsRecording())
		span.End()
	})
}

func TestTracer_StartSpan_NilTracer(t *testing.T) {
	type ctxKey struct{}
	parent := context.WithValue(context.Background(), ctxKey{}, "value")

	tr := &Tracer{}
	ctx, span := tr.StartSpan(parent, "ignored", oteltrace.SpanKindInternal)

	assert.Equal(t, parent, ctx)
	assert.Nil(t, span.Span)
	// the zero-value span wrapper must stay safe to use
	assert.NotPanics(t, span.EndSuccessfully)
}

func TestGetSpanFromContext(t *testing.T) {
	t.Run("returns the active span", func(t *testing.T) {
		restoreGlobals(t)
		sr := tracetest.NewSpanRecorder()
		tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
		t.Cleanup(func() {
			//nolint:errcheck
			tp.Shutdown(context.Background())
		})
		StarCustomTracer(tp.Tracer("get-span-test"))

		ctx, started := StartSpan(context.Background(), "active-span", oteltrace.SpanKindClient)
		defer started.End()

		got := GetSpanFromContext(ctx)
		assert.True(t, got.IsRecording())
		assert.Equal(t, started.SpanContext(), got.SpanContext())
		assert.True(t, got.SpanContext().HasTraceID())
		assert.True(t, got.SpanContext().HasSpanID())
	})

	t.Run("returns a non-recording span when the context has none", func(t *testing.T) {
		got := GetSpanFromContext(context.Background())

		assert.False(t, got.IsRecording())
		assert.False(t, got.SpanContext().IsValid())
	})
}
