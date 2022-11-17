package argocd

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func Test_Marshallers(t *testing.T) {
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
