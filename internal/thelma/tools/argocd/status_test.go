package argocd

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func Test_AppUnmarshalling(t *testing.T) {
	content, err := os.ReadFile("testdata/application.yaml")
	require.NoError(t, err)

	var app application
	err = yaml.Unmarshal(content, &app)
	require.NoError(t, err)

	assert.Equal(t, "HEAD", app.Spec.Source.TargetRevision)

	assert.Equal(t, Degraded, app.Status.Health.Status)
	assert.Equal(t, 18, len(app.Status.Resources))

	var deployment Resource
	for _, r := range app.Status.Resources {
		if r.Name == "workspacemanager-deployment" {
			deployment = r
			break
		}
	}

	assert.Equal(t, "apps", deployment.Group)
	assert.Equal(t, "Deployment", deployment.Kind)
	assert.Equal(t, "v1", deployment.Version)
	assert.Equal(t, "workspacemanager-deployment", deployment.Name)
	assert.Equal(t, "terra-staging", deployment.Namespace)
	assert.Equal(t, OutOfSync, deployment.Status)
	assert.Equal(t, Degraded, deployment.Health.Status)
	assert.Equal(t, `Deployment "workspacemanager-deployment" exceeded its progress deadline`, deployment.Health.Message)

	assert.Equal(t, OutOfSync, app.Status.Sync.Status)
}

func Test_EnumMarshallers(t *testing.T) {
	out, err := yaml.Marshal(Healthy)
	require.NoError(t, err)
	assert.Equal(t, "Healthy\n", string(out))

	var health HealthStatus
	err = yaml.Unmarshal(out, &health)
	require.NoError(t, err)
	assert.Equal(t, Healthy, health)

	out, err = yaml.Marshal(OutOfSync)
	require.NoError(t, err)
	assert.Equal(t, "OutOfSync\n", string(out))

	var syncstatus SyncStatus
	err = yaml.Unmarshal(out, &syncstatus)
	require.NoError(t, err)
	assert.Equal(t, OutOfSync, syncstatus)
}
