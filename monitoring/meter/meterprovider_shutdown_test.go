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

package meter

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	collector "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
)

// restoreGlobalMeterProvider restores the global meter provider and the default meter
// when the test ends.
func restoreGlobalMeterProvider(t *testing.T) {
	t.Helper()

	previousProvider := otel.GetMeterProvider()
	previousMeter := defaultMeter

	t.Cleanup(func() {
		otel.SetMeterProvider(previousProvider)
		defaultMeter = previousMeter
	})
}

func TestInitializeMeterProvider_Branches(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.Monitoring
		env         map[string]string
		expectedErr error
		// expectErr is set when an error is expected that is not a package sentinel error.
		expectErr bool
		// expectNoop is true when the global provider must be the noop provider.
		expectNoop bool
	}{
		{
			name:       "empty service name installs the noop provider",
			cfg:        config.Monitoring{ExporterProtocolCfg: "grpc"},
			expectNoop: true,
		},
		{
			name:        "unsupported protocol installs the noop provider and fails",
			cfg:         config.Monitoring{ServiceCfg: "svc", ExporterProtocolCfg: "carrier-pigeon"},
			expectedErr: ErrExporterProtocolNotSupported,
			expectNoop:  true,
		},
		{
			name:        "metrics protocol takes precedence over the general protocol",
			cfg:         config.Monitoring{ServiceCfg: "svc", ExporterProtocolCfg: "grpc", ExporterMetricsProtocolCfg: "nope"},
			expectedErr: ErrExporterProtocolNotSupported,
			expectNoop:  true,
		},
		{
			name:       "malformed resource attributes install the noop provider and fail",
			cfg:        config.Monitoring{ServiceCfg: "svc", TestExporterCfg: true},
			env:        map[string]string{"OTEL_RESOURCE_ATTRIBUTES": "missing-equals-sign"},
			expectErr:  true,
			expectNoop: true,
		},
		{
			name: "valid config installs the sdk provider",
			cfg:  config.Monitoring{ServiceCfg: "svc", TestExporterCfg: true, MetricsIntervalCfg: 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restoreGlobalMeterProvider(t)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			// make sure a previous provider is never mistaken for the result
			otel.SetMeterProvider(sdkmetric.NewMeterProvider())

			err := initializeMeterProvider(context.Background(), tt.cfg)

			provider := otel.GetMeterProvider()
			if tt.expectNoop {
				assert.IsType(t, noop.MeterProvider{}, provider)
			} else {
				require.NoError(t, err)
				sdkProvider, ok := provider.(*sdkmetric.MeterProvider)
				require.True(t, ok, "unexpected provider %T", provider)
				require.NoError(t, sdkProvider.Shutdown(context.Background()))
				return
			}

			switch {
			case tt.expectedErr != nil:
				require.ErrorIs(t, err, tt.expectedErr)
			case tt.expectErr:
				require.Error(t, err)
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestShutdownMeterProvider(t *testing.T) {
	t.Run("noop provider is a no-op", func(t *testing.T) {
		restoreGlobalMeterProvider(t)
		otel.SetMeterProvider(noop.NewMeterProvider())

		require.NoError(t, ShutdownMeterProvider(context.Background()))
	})

	t.Run("sdk provider is shut down once", func(t *testing.T) {
		restoreGlobalMeterProvider(t)
		reader := sdkmetric.NewManualReader()
		otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))

		require.NoError(t, ShutdownMeterProvider(context.Background()))

		// the reader was shut down together with the provider, so collecting must fail
		var rm metricdata.ResourceMetrics
		require.ErrorIs(t, reader.Collect(context.Background(), &rm), sdkmetric.ErrReaderShutdown)
		// shutting down a second time reports the provider is already shut down
		require.Error(t, ShutdownMeterProvider(context.Background()))
	})
}

// fakeMetricsServer is an in-process OTLP/gRPC metrics collector.
type fakeMetricsServer struct {
	collector.UnimplementedMetricsServiceServer

	mu       sync.Mutex
	received []*collector.ExportMetricsServiceRequest
}

func (s *fakeMetricsServer) Export(_ context.Context, req *collector.ExportMetricsServiceRequest) (*collector.ExportMetricsServiceResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.received = append(s.received, req)
	return &collector.ExportMetricsServiceResponse{}, nil
}

func (s *fakeMetricsServer) requests() []*collector.ExportMetricsServiceRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*collector.ExportMetricsServiceRequest(nil), s.received...)
}

func startFakeMetricsServer(t *testing.T) (*fakeMetricsServer, string) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	fake := &fakeMetricsServer{}
	srv := grpc.NewServer()
	collector.RegisterMetricsServiceServer(srv, fake)

	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	return fake, "http://" + lis.Addr().String()
}

func TestStartDefaultMeter_ExportsMetricsOnShutdown(t *testing.T) {
	restoreGlobalMeterProvider(t)
	config.ResetMonitoringConfig()
	t.Cleanup(config.ResetMonitoringConfig)

	fake, endpoint := startFakeMetricsServer(t)

	t.Setenv("OTEL_SERVICE_NAME", "meter-e2e-service")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	t.Setenv("OTEL_EXPORTER_OTLP_TEST", "false")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", endpoint)

	ctx := context.Background()
	m, err := StartDefaultMeter(ctx)
	require.NoError(t, err)
	require.NotNil(t, m)

	counter, err := IntCounter("e2e.requests", MetricInput{Description: "Requests", Unit: "{request}"})
	require.NoError(t, err)
	counter.Add(ctx, 7)

	// the interval is 60s, so only the final flush on shutdown can deliver the metric
	require.NoError(t, ShutdownMeterProvider(ctx))

	requests := fake.requests()
	require.NotEmpty(t, requests)

	resourceAttrs := map[string]string{}
	var found bool
	for _, req := range requests {
		for _, rm := range req.GetResourceMetrics() {
			for _, kv := range rm.GetResource().GetAttributes() {
				resourceAttrs[kv.GetKey()] = kv.GetValue().GetStringValue()
			}
			for _, sm := range rm.GetScopeMetrics() {
				assert.Equal(t, "meter-e2e-service", sm.GetScope().GetName())
				for _, metric := range sm.GetMetrics() {
					if metric.GetName() != "e2e.requests" {
						continue
					}
					found = true
					assert.Equal(t, `Requests (Measured in "{request}").`, metric.GetDescription())
					assert.Equal(t, "{request}", metric.GetUnit())
					require.Len(t, metric.GetSum().GetDataPoints(), 1)
					assert.Equal(t, int64(7), metric.GetSum().GetDataPoints()[0].GetAsInt())
				}
			}
		}
	}

	assert.True(t, found, "metric e2e.requests was not exported")
	assert.Equal(t, "meter-e2e-service", resourceAttrs["service.name"])
	assert.Equal(t, "grpc", resourceAttrs[config.EXPORTER_PROTOCOL])
}
