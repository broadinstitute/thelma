package testing

import (
	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/testing/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
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
	// Note - we intentionally do NOT pass in the test object to the generated mock, because it
	// will end up calling t.FailNow() in the grpc server's goroutine instead of main,
	// causing tests to hang in some situations. (It's a no-no to call t.FailNow() outside of the test's main goroutine)
	mockServer := &mocks.ClusterManagerServer{}
	t.Cleanup(func() {
		mockServer.AssertExpectations(t)
	})

	server := newFakeGRPCServer(t, func(s *grpc.Server) {
		containerpb.RegisterClusterManagerServer(s, mockServer)
	})

	client, err := container.NewClusterManagerClient(context.TODO(), server.clientOptions...)
	require.NoError(t, err)

	return &ContainerClusterManagerMocks{mockServer: mockServer}, client
}
