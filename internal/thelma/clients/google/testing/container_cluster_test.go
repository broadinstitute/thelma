package testing

import (
	"context"
	"testing"

	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
