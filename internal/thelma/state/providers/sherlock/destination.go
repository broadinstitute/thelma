package sherlock

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

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
	panic("abstract mehtod implemented by children")
}

func (t *destination) RequireSuitable() bool {
	return t.requireSuitable
}

// TODO
func (t *destination) TerraHelmfileRef() string {
	panic("TODO, not yet implemented on Sherlock backend")
}
