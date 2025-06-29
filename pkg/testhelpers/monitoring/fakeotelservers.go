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

	testhelpers "github.com/FLYR-Open-Source/flyr-lib-go/pkg/testhelpers/grpc"
	collector "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

type MockOtelMetricsServer struct {
	collector.UnimplementedMetricsServiceServer
	received []*collector.ExportMetricsServiceRequest
}

func (m *MockOtelMetricsServer) Export(ctx context.Context, req *collector.ExportMetricsServiceRequest) (*collector.ExportMetricsServiceResponse, error) {
	m.received = append(m.received, req)
	return &collector.ExportMetricsServiceResponse{}, nil
}

func NewOtelMetricsGrpcServer(mockServer collector.MetricsServiceServer) (*grpc.Server, []option.ClientOption, error) {
	cb := func(grpcServer *grpc.Server) error {
		collector.RegisterMetricsServiceServer(grpcServer, mockServer)
		return nil
	}

	return testhelpers.SetupMockGrpcServer(cb)
}
