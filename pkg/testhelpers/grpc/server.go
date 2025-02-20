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
	"errors"
	"net"

	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ErrInvalidServerType is returned when the server type is invalid.
var ErrInvalidServerType = errors.New("invalid server type")

type RegisterServerFunc = func(*grpc.Server) error

// setupMockGrpcServer starts a mock gRPC server and returns the options to be passed on the client.
//
// The function is generic and expects as a type, the type of the mock server.
//
// For more information, see: https://github.com/googleapis/google-cloud-go/blob/main/testing.md#testing-grpc-services-using-fakes
// Also see the examples in the `examples/testhelpers/grpc` package.
func SetupMockGrpcServer(registerServer RegisterServerFunc) (*grpc.Server, []option.ClientOption, error) {
	// Create a gRPC server and register the mock service.
	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return grpcServer, []option.ClientOption{}, err
	}

	// Register the mock server to the GRPC server.
	err = registerServer(grpcServer)
	if err != nil {
		return grpcServer, []option.ClientOption{}, err
	}

	// Use a channel to communicate errors from the goroutine.
	errChan := make(chan error, 1)

	fakeServerAddr := lis.Addr().String()
	// Start the gRPC server in a goroutine.
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- err // Send the error to the channel.
		}
	}()

	// Check if the server started successfully.
	select {
	case err := <-errChan:
		return grpcServer, []option.ClientOption{}, err
	default:
		// No error, continue.
	}

	return grpcServer, []option.ClientOption{
		option.WithEndpoint(fakeServerAddr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}, nil
}

// shutdownGrpcServer shuts down the gRPC server.
func ShutdownGrpcServer(grpcServer *grpc.Server) {
	grpcServer.GracefulStop()
}
