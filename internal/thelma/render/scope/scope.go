package scope

import (
	"github.com/pkg/errors"
)

// Scope is an enum type representing different output formats
type Scope int

const (
	// All renders all resources
	All Scope = iota
	// Release renders release-scoped resources only
	Release
	// Destination renders destination-scoped resources (such as ArgoCD project + generator) only
	Destination
)

// FromString will set the receiver's value to the one denoted by the given string
func FromString(value string) (Scope, error) {
	switch value {
	case "all":
		return All, nil
	case "release":
		return Release, nil
	case "destination":
		return Destination, nil
	}
	return All, errors.Errorf("unknown format: %q", value)
}

// String returns a string representation of this format
func (s Scope) String() string {
	switch s {
	case All:
		return "all"
	case Release:
		return "release"
	case Destination:
		return "destination"
	}
	return "unknown"
}
