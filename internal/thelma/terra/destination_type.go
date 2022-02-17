package terra

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

// DestinationType is an enum type referring to the two types of destinations supported by terra-helmfile.
type DestinationType int

const (
	EnvironmentDestination DestinationType = iota
	ClusterDestination
)

// CompareReleases Returns 0 if t == other, -1 if t < other, or +1 if t > other.
func (t DestinationType) Compare(other DestinationType) int {
	if t == other {
		return 0
	}
	if t == EnvironmentDestination {
		return -1
	}
	return 1
}

// UnmarshalYAML is a custom unmarshaler so that the string "environment" or "cluster" in a
// yaml file can be unmarshaled into a DestinationType
func (t *DestinationType) UnmarshalYAML(value *yaml.Node) error {
	switch value.Value {
	case "environment":
		*t = EnvironmentDestination
		return nil
	case "cluster":
		*t = ClusterDestination
		return nil
	}

	return fmt.Errorf("unknown destination type: %v", value.Value)
}

func (t DestinationType) String() string {
	switch t {
	case EnvironmentDestination:
		return "environment"
	case ClusterDestination:
		return "cluster"
	}
	return "unknown"
}
