package testing

import (
	"context"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"net"
	"testing"
)

type grpcServer struct {
	server        *grpc.Server
	clientOptions []option.ClientOption
}

// NotFoundError convenience constructor for returning a not found error in mocks
func NotFoundError() error {
	return status.Error(codes.NotFound, "Not found")
}

// construct a new fake grpc server
// reference: https://github.com/googleapis/google-cloud-go/blob/main/testing.md#testing-grpc-services-using-fakes
func newFakeGRPCServer(t *testing.T, registerMockBackends func(server *grpc.Server)) grpcServer {
	// Create a GRPC server
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	gsrv := grpc.NewServer(grpc.UnaryInterceptor(convertPanicIntoErrorInterceptor))
	t.Cleanup(func() {
		gsrv.Stop()
	})

	// register mock backends with the server
	registerMockBackends(gsrv)

	// start server in async goroutine
	go func() {
		if err := gsrv.Serve(listener); err != nil {
			// we can't use t.Fatal err since we're in a separate goroutine
			panic(err)
		}
	}()

	// Create client that is configured to talk to the fake GRPC server
	fakeServerAddr := listener.Addr().String()

	return grpcServer{
		server: gsrv,
		clientOptions: []option.ClientOption{
			option.WithEndpoint(fakeServerAddr),
			option.WithoutAuthentication(),
			option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		},
	}
}

// convertPanicIntoErrorInterceptor detects when a handler panics and returns it as a GRPC error.
// This is useful because we are using mockery-generated Testify mocks as Google API GRPC backends.
// The mocks panic when an unexpected call is made. This interceptor will recover the panic,
// convert it into a GRPC error and return it to the client.
// This is desirable because it causes the individual test case to fail gracefully (a panic seems to kill the entire go
// test run).
func convertPanicIntoErrorInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (result interface{}, err error) {

	// Set up a defer function to intercept a panic
	defer func() {
		if panicErr := recover(); panicErr != nil {
			// convert panic into a GRPC error that will be returned to client
			err = status.Errorf(codes.Internal, "%v", panicErr)
		}
	}()

	result, err = handler(ctx, req)

	return
}
