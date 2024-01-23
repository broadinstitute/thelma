package statefixtures

import (
	_ "embed"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
)

type Mocks struct {
	Clusters     *statemocks.Clusters
	Environments *StubEnvironments
	Releases     *StubReleases
	State        *statemocks.State
	StateLoader  *statemocks.StateLoader

	Items struct {
		Clusters        map[string]*statemocks.Cluster
		Environments    map[string]*statemocks.Environment
		AppReleases     map[string]*statemocks.AppRelease
		ClusterReleases map[string]*statemocks.ClusterRelease
	}
}

type StubReleases struct {
	*statemocks.Releases
}

func (m *StubReleases) Filter(filter terra.ReleaseFilter) ([]terra.Release, error) {
	all, err := m.Releases.All()
	if err != nil {
		return nil, err
	}
	return filter.Filter(all), nil
}

type StubEnvironments struct {
	*statemocks.Environments
}

func (m *StubEnvironments) Filter(filter terra.EnvironmentFilter) ([]terra.Environment, error) {
	all, err := m.Environments.All()
	if err != nil {
		return nil, err
	}
	return filter.Filter(all), nil
}
