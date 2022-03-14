package statefixtures

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/terra"
	"github.com/broadinstitute/thelma/internal/thelma/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"testing"
)

// Fixture is a convenience interface for retrieving environments, clusters and from state by name
type Fixture interface {
	Environment(name string) terra.Environment
	Cluster(name string) terra.Cluster
	Release(name string, dest string) terra.Release
	AllReleases() []terra.Release
}

func LoadFixture(name FixtureName, t *testing.T) Fixture {
	loader, err := NewFakeStateLoader(name, t, t.TempDir(), shell.NewRunner())
	if err != nil {
		panic(err)
	}
	state, err := loader.Load()
	if err != nil {
		panic(err)
	}
	return NewFixture(state)
}

func NewFixture(state terra.State) Fixture {
	return &fixture{
		state: state,
	}
}

type fixture struct {
	state terra.State
}

func (f *fixture) Environment(name string) terra.Environment {
	e, err := f.state.Environments().Get(name)
	if err != nil {
		panic(err)
	}
	return e
}

func (f *fixture) Cluster(name string) terra.Cluster {
	c, err := f.state.Clusters().Get(name)
	if err != nil {
		panic(err)
	}
	return c
}

func (f *fixture) Release(name string, dest string) terra.Release {
	matches, err := f.state.Releases().Filter(
		filter.Releases().And(
			filter.Releases().HasName(name),
			filter.Releases().HasDestinationName(dest),
		),
	)
	if err != nil {
		panic(err)
	}
	if len(matches) > 1 {
		panic(fmt.Errorf("more than one release matched %s in %s: %v", name, dest, matches))
	}
	if len(matches) == 0 {
		return nil
	}
	return matches[0]
}

func (f *fixture) AllReleases() []terra.Release {
	releases, err := f.state.Releases().All()
	if err != nil {
		panic(err)
	}
	return releases
}
