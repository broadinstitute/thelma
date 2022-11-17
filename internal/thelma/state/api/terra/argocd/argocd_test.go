package argocd

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Names(t *testing.T) {
	samDev := mocks.NewRelease(t)
	samDev.On("Name").Return("sam")
	devEnv := mocks.NewEnvironment(t)
	devEnv.On("Name").Return("dev")
	samDev.On("Destination").Return(devEnv)

	yaleTerraDev := mocks.NewRelease(t)
	yaleTerraDev.On("Name").Return("yale")
	devCluster := mocks.NewCluster(t)
	devCluster.On("Name").Return("terra-dev")
	yaleTerraDev.On("Destination").Return(devCluster)

	assert.Equal(t, "sam-dev", ApplicationName(samDev))
	assert.Equal(t, "yale-terra-dev", ApplicationName(yaleTerraDev))

	assert.Equal(t, "sam-configs-dev", LegacyConfigsApplicationName(samDev))

	assert.Equal(t, "terra-dev", ProjectName(samDev.Destination()))
	assert.Equal(t, "cluster-terra-dev", ProjectName(yaleTerraDev.Destination()))
}
