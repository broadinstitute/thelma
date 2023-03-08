package kubecfg

import (
	"encoding/base64"
	"github.com/broadinstitute/thelma/internal/thelma/utils/testutils"
	"os"
	"path"
	"testing"

	"cloud.google.com/go/container/apiv1/containerpb"
	googletesting "github.com/broadinstitute/thelma/internal/thelma/clients/google/testing"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Kubecfg(t *testing.T) {
	cluster1 := mocks.NewCluster(t)
	cluster1.EXPECT().Name().Return("cluster1")
	cluster1.EXPECT().Project().Return("cluster1-project")
	cluster1.EXPECT().Location().Return("us-central1-a")
	cluster1.EXPECT().Address().Return("https://cluster1-address/")

	cluster2 := mocks.NewCluster(t)
	cluster2.EXPECT().Name().Return("cluster2")
	cluster2.EXPECT().Project().Return("cluster2-project")
	cluster2.EXPECT().Location().Return("us-central1")
	cluster2.EXPECT().Address().Return("https://cluster2-address/")

	env1 := mocks.NewEnvironment(t)
	env1.EXPECT().Name().Return("env1")
	env1.EXPECT().DefaultCluster().Return(cluster1)
	env1.EXPECT().Namespace().Return("env1-namespace")

	// release1 - deployed to env1's default cluster (cluster1)
	release1 := mocks.NewAppRelease(t)
	release1.EXPECT().Name().Return("release1")
	release1.EXPECT().IsClusterRelease().Return(false)
	release1.EXPECT().Environment().Return(env1)
	release1.EXPECT().Cluster().Return(cluster1)
	release1.EXPECT().Namespace().Return("env1-namespace")

	// release2 - deployed to non-default cluster (cluster2)
	release2 := mocks.NewAppRelease(t)
	release2.EXPECT().Name().Return("release2")
	release2.EXPECT().IsClusterRelease().Return(false)
	release2.EXPECT().Environment().Return(env1)
	release2.EXPECT().Cluster().Return(cluster2)
	release2.EXPECT().Namespace().Return("env1-namespace")

	// release3 - deployed to env1's default cluster (cluster1)
	release3 := mocks.NewAppRelease(t)
	release3.EXPECT().Name().Return("release3")
	release3.EXPECT().IsClusterRelease().Return(false)
	release3.EXPECT().Environment().Return(env1)
	release3.EXPECT().Cluster().Return(cluster1)
	release3.EXPECT().Namespace().Return("env1-namespace")

	env1.EXPECT().Releases().Return([]terra.Release{release1, release2})

	// release4 - cluster release deployed in cluster1
	release4 := mocks.NewClusterRelease(t)
	release4.EXPECT().Name().Return("release4")
	release4.EXPECT().IsClusterRelease().Return(true)
	release4.EXPECT().Cluster().Return(cluster1)
	release4.EXPECT().Namespace().Return("release4-namespace")

	// create a mock gke server and connected client
	gkeMock, gkeClient := googletesting.NewMockClusterManagerServerAndClient(t)
	gkeMock.ExpectGetCluster("cluster1-project", "us-central1-a", "cluster1", &containerpb.Cluster{
		Name: "cluster1",
		MasterAuth: &containerpb.MasterAuth{
			ClusterCaCertificate: base64.StdEncoding.EncodeToString([]byte("fake-cluster1-cert")),
		},
	}, nil)

	gkeMock.ExpectGetCluster("cluster2-project", "us-central1", "cluster2", &containerpb.Cluster{
		Name: "cluster2",
		MasterAuth: &containerpb.MasterAuth{
			ClusterCaCertificate: base64.StdEncoding.EncodeToString([]byte("fake-cluster2-cert")),
		},
	}, nil)

	// write kubecfg to temporary file
	file := path.Join(t.TempDir(), "kubecfg")
	tokenSource := testutils.NewFakeTokenSource("fake-token")
	kubecfg := New(file, gkeClient, tokenSource)

	// get kubecfg for release1
	// because release1 is deployed to env1's default cluster, we should be using the environment's default context
	kubectx, err := kubecfg.ForRelease(release1)
	require.NoError(t, err)
	assert.Equal(t, "env1", kubectx.ContextName())
	assert.Equal(t, "env1-namespace", kubectx.Namespace())
	assertFilesHaveSameContent(t, "testdata/kubecfg-01.yaml", file)

	// release3 is ALSO deployed to env1's default cluster, so should be using the environment's default context.
	// the kubecfg should have no changes, since no new context needs to be added.
	kubectx, err = kubecfg.ForRelease(release3)
	require.NoError(t, err)
	assert.Equal(t, "env1", kubectx.ContextName())
	assert.Equal(t, "env1-namespace", kubectx.Namespace())
	assertFilesHaveSameContent(t, "testdata/kubecfg-01.yaml", file)

	// now we will exercise ForEnvironment and verify that a new context for release2 is generated
	kubectxs, err := kubecfg.ForEnvironment(env1)
	require.NoError(t, err)
	assert.Equal(t, 2, len(kubectxs))
	assert.Equal(t, "env1", kubectxs[0].ContextName())
	assert.Equal(t, "env1-namespace", kubectxs[0].Namespace())
	assert.Equal(t, "env1_release2", kubectxs[1].ContextName())
	assert.Equal(t, "env1-namespace", kubectxs[1].Namespace())

	assertFilesHaveSameContent(t, "testdata/kubecfg-02.yaml", file)

	// finally we will exercise ForReleases and verify that a new context for the cluster release release4 is generated
	releasektxs, err := kubecfg.ForReleases(release1, release2, release3, release4)
	require.NoError(t, err)
	assert.Equal(t, 4, len(releasektxs))

	assert.Equal(t, "release1", releasektxs[0].Release.Name())
	assert.Equal(t, "env1", releasektxs[0].Kubectx.ContextName())
	assert.Equal(t, "env1-namespace", releasektxs[0].Kubectx.Namespace())

	assert.Equal(t, "release2", releasektxs[1].Release.Name())
	assert.Equal(t, "env1_release2", releasektxs[1].Kubectx.ContextName())
	assert.Equal(t, "env1-namespace", releasektxs[1].Kubectx.Namespace())

	assert.Equal(t, "release3", releasektxs[2].Release.Name())
	assert.Equal(t, "env1", releasektxs[2].Kubectx.ContextName())
	assert.Equal(t, "env1-namespace", releasektxs[2].Kubectx.Namespace())

	assert.Equal(t, "release4", releasektxs[3].Release.Name())
	assert.Equal(t, "cluster1_release4", releasektxs[3].Kubectx.ContextName())
	assert.Equal(t, "release4-namespace", releasektxs[3].Kubectx.Namespace())

	assertFilesHaveSameContent(t, "testdata/kubecfg-03.yaml", file)
}

func assertFilesHaveSameContent(t *testing.T, expectedFile string, actualFile string) {
	expectedBytes, err := os.ReadFile(expectedFile)
	require.NoError(t, err)

	actualBytes, err := os.ReadFile(actualFile)
	require.NoError(t, err)

	assert.Equal(t, string(expectedBytes), string(actualBytes), "expected contents of %s to match contents of %s", actualFile, expectedFile)
}
