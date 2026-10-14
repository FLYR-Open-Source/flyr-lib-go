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

package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	collector "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	transport "google.golang.org/api/transport/grpc"
)

func TestMockOtelMetricsServer_Export(t *testing.T) {
	m := &MockOtelMetricsServer{}
	req := &collector.ExportMetricsServiceRequest{}

	resp, err := m.Export(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	require.Len(t, m.received, 1)
	assert.Same(t, req, m.received[0])
}

func TestNewOtelMetricsGrpcServer_ReceivesExportedMetrics(t *testing.T) {
	mock := &MockOtelMetricsServer{}

	srv, opts, err := NewOtelMetricsGrpcServer(mock)
	require.NoError(t, err)
	require.NotNil(t, srv)
	require.NotEmpty(t, opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := transport.Dial(ctx, opts...)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, conn.Close()) })

	req := &collector.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricspb.ResourceMetrics{{
			Resource: &resourcepb.Resource{
				Attributes: []*commonpb.KeyValue{{
					Key:   "service.name",
					Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: "svc"}},
				}},
			},
			ScopeMetrics: []*metricspb.ScopeMetrics{{
				Metrics: []*metricspb.Metric{{Name: "requests_total"}},
			}},
		}},
	}
	_, err = collector.NewMetricsServiceClient(conn).Export(ctx, req)
	require.NoError(t, err)

	// Stopping the server waits for in-flight handlers, which orders the read of mock.received after the write.
	srv.GracefulStop()

	require.Len(t, mock.received, 1)
	got := mock.received[0].GetResourceMetrics()
	require.Len(t, got, 1)
	assert.Equal(t, "service.name", got[0].GetResource().GetAttributes()[0].GetKey())
	assert.Equal(t, "requests_total", got[0].GetScopeMetrics()[0].GetMetrics()[0].GetName())
}

type unimplementedMetricsServer struct {
	collector.UnimplementedMetricsServiceServer
}

func TestNewOtelMetricsGrpcServer_UsesProvidedServer(t *testing.T) {
	// A server that does not implement Export answers with Unimplemented, proving the provided
	// implementation (not MockOtelMetricsServer) is the one registered.
	srv, opts, err := NewOtelMetricsGrpcServer(&unimplementedMetricsServer{})
	require.NoError(t, err)
	t.Cleanup(srv.GracefulStop)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := transport.Dial(ctx, opts...)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, conn.Close()) })

	_, err = collector.NewMetricsServiceClient(conn).Export(ctx, &collector.ExportMetricsServiceRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unimplemented")
}
