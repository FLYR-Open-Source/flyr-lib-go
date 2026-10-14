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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	transport "google.golang.org/api/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func TestSetupMockGrpcServer_ServesRegisteredService(t *testing.T) {
	healthServer := health.NewServer()
	healthServer.SetServingStatus("svc", healthpb.HealthCheckResponse_SERVING)

	srv, opts, err := SetupMockGrpcServer(func(s *grpc.Server) error {
		healthpb.RegisterHealthServer(s, healthServer)
		return nil
	})
	require.NoError(t, err)
	require.NotNil(t, srv)
	require.NotEmpty(t, opts)
	t.Cleanup(func() { ShutdownGrpcServer(srv) })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := transport.Dial(ctx, opts...)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, conn.Close()) })

	resp, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{Service: "svc"})
	require.NoError(t, err)
	assert.Equal(t, healthpb.HealthCheckResponse_SERVING, resp.GetStatus())
}

func TestSetupMockGrpcServer_RegisterError(t *testing.T) {
	registerErr := errors.New("register failed")

	srv, opts, err := SetupMockGrpcServer(func(*grpc.Server) error {
		return registerErr
	})

	require.ErrorIs(t, err, registerErr)
	assert.NotNil(t, srv)
	assert.Empty(t, opts)
}

func TestSetupMockGrpcServer_ReturnsClientOptions(t *testing.T) {
	srv, opts, err := SetupMockGrpcServer(func(*grpc.Server) error { return nil })
	require.NoError(t, err)
	t.Cleanup(func() { ShutdownGrpcServer(srv) })

	// endpoint, no-auth and insecure transport credentials.
	assert.Len(t, opts, 3)
}

func TestShutdownGrpcServer_StopsServing(t *testing.T) {
	srv, opts, err := SetupMockGrpcServer(func(s *grpc.Server) error {
		healthpb.RegisterHealthServer(s, health.NewServer())
		return nil
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := transport.Dial(ctx, opts...)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, conn.Close()) })

	client := healthpb.NewHealthClient(conn)
	_, err = client.Check(ctx, &healthpb.HealthCheckRequest{})
	require.NoError(t, err)

	ShutdownGrpcServer(srv)

	// GracefulStop closes the listener, so further calls cannot succeed.
	_, err = client.Check(ctx, &healthpb.HealthCheckRequest{}, grpc.WaitForReady(false))
	require.Error(t, err)
	assert.NotEqual(t, connectivity.Ready, conn.GetState())
}
