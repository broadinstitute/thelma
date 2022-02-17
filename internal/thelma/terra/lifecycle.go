package terra

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

// Lifecycle is an enum type that represents the different types of lifecycles a Terra environment can have
type Lifecycle int

const (
	// Static environments are long-lived environments that are created manually and never destroyed. `prod` is an example of a static environment.
	Static Lifecycle = iota
	// Template environments are never created or destroyed. They exist solely in configuration files, and are used to create dynamic environments.
	Template
	// Dynamic environments are created and destroyed on demand, from a template environment.
	Dynamic
)

// Lifecycles returns a slice of all possible Lifecycles
func Lifecycles() []Lifecycle {
	return []Lifecycle{Static, Template, Dynamic}
}

// Compare returns 0 if r == other, -1 if r < other, or +1 if r > other.
func (l Lifecycle) Compare(other Lifecycle) int {
	diff := l - other
	if diff < 0 {
		return -1
	}
	if diff > 0 {
		return 1
	}
	return 0
}

// UnmarshalYAML is a custom unmarshaler so that the string "app" or "cluster" in a
// yaml file can be unmarshaled into a ReleaseType
func (l *Lifecycle) UnmarshalYAML(value *yaml.Node) error {
	switch value.Value {
	case "static":
		*l = Static
		return nil
	case "template":
		*l = Template
		return nil
	case "dynamic":
		*l = Dynamic
		return nil
	}

	var supported []string
	for _, l := range Lifecycles() {
		supported = append(supported, l.String())
	}
	return fmt.Errorf("unknown lifecycle type %v, supported lifecycles are %s", value.Value, strings.Join(supported, ", "))
}

func (l Lifecycle) String() string {
	switch l {
	case Static:
		return "static"
	case Template:
		return "template"
	case Dynamic:
		return "dynamic"
	}
	return "unknown"
}

func (l Lifecycle) IsStatic() bool {
	return l == Static
}

func (l Lifecycle) IsTemplate() bool {
	return l == Template
}

func (l Lifecycle) IsDynamic() bool {
	return l == Dynamic
}
