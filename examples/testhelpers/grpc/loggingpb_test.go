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
	"testing"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	"cloud.google.com/go/logging/logadmin"
	testhelpers "github.com/FlyrInc/flyr-lib-go/examples/testhelpers/grpc"
	"google.golang.org/api/iterator"
	logtypepb "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLogGRPC(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"

	expectedEntry := &loggingpb.LogEntry{
		Labels:    map[string]string{"key": "value"},
		Severity:  logtypepb.LogSeverity_DEFAULT,
		Timestamp: timestamppb.New(time.Now()),
		InsertId:  "some-insert-id",
		Payload:   &loggingpb.LogEntry_TextPayload{TextPayload: "Some message"},
	}

	mockLogServer := &testhelpers.MockLoggingServiceServer{
		Entries:           []*loggingpb.LogEntry{expectedEntry},
		ListLogEntriesErr: nil,
	}
	grpcServer, logOpts, err := testhelpers.NewLoggingServiceServer(mockLogServer)
	if err != nil {
		t.Fatalf("failed to create mock logging service server: %v", err)
		return
	}
	defer grpcServer.GracefulStop()

	logClient, err := logadmin.NewClient(ctx, projectID, logOpts...)
	if err != nil {
		t.Fatalf("failed to create log client: %v", err)
		return
	}
	defer logClient.Close()

	it := logClient.Entries(ctx)
	for {
		entry, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}

			t.Fatalf("failed to get next log entry: %v", err)
			return
		}

		if entry.Payload != nil && entry.Payload.(string) != expectedEntry.GetTextPayload() {
			t.Fatalf("unexpected log entry: got %v, want %v", entry.Payload, expectedEntry.GetTextPayload())
			return
		}
	}
}
