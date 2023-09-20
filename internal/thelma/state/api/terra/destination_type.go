package terra

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// DestinationType is an enum type referring to the two types of destinations supported by terra-helmfile.
type DestinationType int

const (
	EnvironmentDestination DestinationType = iota
	ClusterDestination
)

// DestinationTypes returns all destination types
func DestinationTypes() []DestinationType {
	return []DestinationType{
		EnvironmentDestination,
		ClusterDestination,
	}
}

// DestinationTypeNames returns a list of destination type names as strings
func DestinationTypeNames() []string {
	var names []string
	for _, dt := range DestinationTypes() {
		names = append(names, dt.String())
	}
	return names
}

// Compare returns 0 if t == other, -1 if t < other, or +1 if t > other.
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
	return t.FromString(value.Value)
}

// FromString will set the receiver's value to the one denoted by the given string
func (t *DestinationType) FromString(value string) error {
	switch value {
	case "environment":
		*t = EnvironmentDestination
		return nil
	case "cluster":
		*t = ClusterDestination
		return nil
	}

	return errors.Errorf("unknown destination type: %q", value)
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
