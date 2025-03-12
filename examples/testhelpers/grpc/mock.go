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

package grpc

import (
	"context"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	testhelpers "github.com/FLYR-Open-Source/flyr-lib-go/pkg/testhelpers/grpc"
)

// MockLoggingServiceServer is a mock implementation of the LoggingServiceV2Server interface.
type MockLoggingServiceServer struct {
	loggingpb.UnimplementedLoggingServiceV2Server
	Entries           []*loggingpb.LogEntry
	ListLogEntriesErr error
}

func (m *MockLoggingServiceServer) WithEntries(entries []*loggingpb.LogEntry) {
	m.Entries = entries
}

func (m *MockLoggingServiceServer) WithListLogEntriesError(err error) {
	m.ListLogEntriesErr = err
}

// ListLogEntries implements the ListLogEntries method of the LoggingServiceV2Server interface.
func (m *MockLoggingServiceServer) ListLogEntries(ctx context.Context, req *loggingpb.ListLogEntriesRequest) (*loggingpb.ListLogEntriesResponse, error) {
	if m.ListLogEntriesErr != nil {
		return nil, m.ListLogEntriesErr
	}
	return &loggingpb.ListLogEntriesResponse{
		Entries: m.Entries,
	}, nil
}

func NewLoggingServiceServer(mockServer loggingpb.LoggingServiceV2Server) (*grpc.Server, []option.ClientOption, error) {
	cb := func(grpcServer *grpc.Server) error {
		loggingpb.RegisterLoggingServiceV2Server(grpcServer, mockServer)
		return nil
	}

	return testhelpers.SetupMockGrpcServer(cb)
}
