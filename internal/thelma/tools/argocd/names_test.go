package argocd

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statefixtures"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Names(t *testing.T) {
	fixture := statefixtures.LoadFixture(statefixtures.Default, t)
	samDev := fixture.Release("sam", "dev")
	yaleTerraDev := fixture.Release("yale", "terra-dev")

	assert.Equal(t, "sam-dev", ApplicationName(samDev))
	assert.Equal(t, "yale-terra-dev", ApplicationName(yaleTerraDev))

	assert.Equal(t, "sam-configs-dev", LegacyConfigsApplicationName(samDev))

	assert.Equal(t, "terra-dev", ProjectName(samDev.Destination()))
	assert.Equal(t, "cluster-terra-dev", ProjectName(yaleTerraDev.Destination()))

	assert.Equal(t, map[string]string{"app": "sam", "env": "dev"}, releaseSelector(samDev))
	assert.Equal(t, map[string]string{"release": "yale", "cluster": "terra-dev", "type": "cluster"}, releaseSelector(yaleTerraDev))
}
