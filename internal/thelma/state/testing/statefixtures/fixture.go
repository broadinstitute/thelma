package statefixtures

import (
	"embed"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"path"
)

//go:embed fixtures/*.yaml
var fixturesFS embed.FS

// Fixture is a convenience interface for retrieving environments, clusters and from state by name
type Fixture interface {
	Environment(name string) *statemocks.Environment
	Cluster(name string) *statemocks.Cluster
	Release(name string, dest string) terra.Release
	AllReleases() []terra.Release
	Mocks() *Mocks
}

// LoadFixture load a state fixture for use in tests
//
// Deprecated: this package was added for backwards compatibility when we
// deleted gitops state; new tests that depend on state should set up their own mocks.
//
// Old tests should also ideally be refactored to mock their own state.
func LoadFixture(name FixtureName) (Fixture, error) {
	content, err := fixturesFS.ReadFile(path.Join("fixtures", name.String()+".yaml"))
	if err != nil {
		return nil, errors.Errorf("error loading %s fixture: %v", name.String(), err)
	}

	var data FixtureData
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		return nil, errors.Errorf("error loading %s fixture: %v", name.String(), err)
	}

	mocks := newBuilder(&data).buildMocks()
	return NewFixture(mocks), nil
}

func NewFixture(mocks *Mocks) Fixture {
	return &fixture{
		mocks: mocks,
	}
}

type fixture struct {
	mocks *Mocks
}

func (f *fixture) Mocks() *Mocks {
	return f.mocks
}

func (f *fixture) Environment(name string) *statemocks.Environment {
	return f.mocks.Items.Environments[name]
}

func (f *fixture) Cluster(name string) *statemocks.Cluster {
	return f.mocks.Items.Clusters[name]
}

func (f *fixture) Release(name string, dest string) terra.Release {
	key := name + "-" + dest
	r, exists := f.mocks.Items.ClusterReleases[key]
	if exists {
		return r
	}
	return f.mocks.Items.AppReleases[key]
}

func (f *fixture) AllReleases() []terra.Release {
	var allReleases []terra.Release

	for _, r := range f.mocks.Items.ClusterReleases {
		allReleases = append(allReleases, r)
	}

	for _, r := range f.mocks.Items.AppReleases {
		allReleases = append(allReleases, r)
	}

	return allReleases
}
