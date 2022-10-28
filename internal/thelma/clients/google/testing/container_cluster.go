package testing

import (
	container "cloud.google.com/go/container/apiv1"
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/testing/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
)

type ContainerClusterManagerMocks struct {
	mockServer *mocks.ClusterManagerServer
}

func (c *ContainerClusterManagerMocks) ClusterManagerServer() *mocks.ClusterManagerServer {
	return c.mockServer
}

func (c *ContainerClusterManagerMocks) ExpectGetCluster(projectId string, location string, clusterName string, cluster *containerpb.Cluster, err error) {
	c.mockServer.On("GetCluster",
		mock.Anything,
		&containerpb.GetClusterRequest{
			Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectId, location, clusterName),
		},
	).Return(cluster, err)
}

// NewMockClusterManagerServerAndClient creates a connected mock cluster manager server and client
func NewMockClusterManagerServerAndClient(t *testing.T) (*ContainerClusterManagerMocks, *container.ClusterManagerClient) {
	mockServer := mocks.NewClusterManagerServer(t)

	// Create a GRPC server
	// reference: https://github.com/googleapis/google-cloud-go/blob/main/testing.md#testing-grpc-services-using-fakes
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	gsrv := grpc.NewServer()
	t.Cleanup(func() {
		gsrv.Stop()
	})

	// Configure the GRPC server to serve responses from our mock cluster manager server
	containerpb.RegisterClusterManagerServer(gsrv, mockServer)
	go func() {
		if err := gsrv.Serve(listener); err != nil {
			panic(err)
		}
	}()

	// Create client that is configured to talk to the fake GRPC server
	fakeServerAddr := listener.Addr().String()
	client, err := container.NewClusterManagerClient(context.TODO(),
		option.WithEndpoint(fakeServerAddr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	require.NoError(t, err)

	return &ContainerClusterManagerMocks{mockServer: mockServer}, client
}
