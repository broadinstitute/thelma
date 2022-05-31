package gitops

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

// implements the terra.Destination interface
type destination struct {
	name             string
	base             string
	requireSuitable  bool
	destinationType  terra.DestinationType
	terraHelmfileRef string
}

func (t *destination) Name() string {
	return t.name
}

func (t *destination) Base() string {
	return t.base
}

func (t *destination) Type() terra.DestinationType {
	return t.destinationType
}

func (t *destination) TerraHelmfileRef() string {
	return t.terraHelmfileRef
}

func (t *destination) IsCluster() bool {
	return t.destinationType == terra.ClusterDestination
}

func (t *destination) IsEnvironment() bool {
	return t.destinationType == terra.EnvironmentDestination
}

func (t *destination) Releases() []terra.Release {
	panic("abstract method implemented by children")
}

func (t *destination) ReleaseType() terra.ReleaseType {
	panic("abstract method implemented by children")
}

func (t *destination) RequireSuitable() bool {
	return t.requireSuitable
}
