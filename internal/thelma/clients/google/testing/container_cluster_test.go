package testing

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	"testing"
)

func Test_ClusterManager(t *testing.T) {
	clusterManagerMocks, client := NewMockClusterManagerServerAndClient(t)

	clusterManagerMocks.ExpectGetCluster("my-project-id", "us-central1-a", "my-cluster", &containerpb.Cluster{
		Name:        "my-cluster",
		Description: "A cluster in my project",
	}, nil)

	cluster, err := client.GetCluster(context.TODO(), &containerpb.GetClusterRequest{
		Name: "projects/my-project-id/locations/us-central1-a/clusters/my-cluster",
	})

	require.NoError(t, err)
	assert.Equal(t, "my-cluster", cluster.Name)
	assert.Equal(t, "A cluster in my project", cluster.Description)
}
