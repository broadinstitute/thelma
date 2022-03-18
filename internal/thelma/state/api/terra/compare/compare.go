package compare

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"strings"
)

// Comparators for infrastructure types

// Releases Compares two releases.
// Returns 0 if r1 == r2, -1 if r1 < r2, or +1 if r1 > r2.
// Compares by type, then by name, then by destination
func Releases(r1 terra.Release, r2 terra.Release) int {
	byType := r1.Type().Compare(r2.Type())
	if byType != 0 {
		return byType
	}
	byName := strings.Compare(r1.Name(), r2.Name())
	if byName != 0 {
		return byName
	}
	byDestination := Destinations(r1.Destination(), r2.Destination())
	return byDestination
}

// Destinations compares two release destinations.
// Returns 0 if t1 == t2, -1 if t1 < t2, or +1 if t1 > t2.
// Compares lexicographically by type, by base, and then by name.
func Destinations(t1 terra.Destination, t2 terra.Destination) int {
	byType := t1.Type().Compare(t2.Type())
	if byType != 0 {
		return byType
	}
	byBase := strings.Compare(t1.Base(), t2.Base())
	if byBase != 0 {
		return byBase
	}
	byName := strings.Compare(t1.Name(), t2.Name())
	return byName
}
