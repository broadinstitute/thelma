package stateutils

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

// ReleaseFullNames return the full names of a slice of releases
func ReleaseFullNames(releases []terra.Release) []string {
	var names []string
	for _, r := range releases {
		names = append(names, r.FullName())
	}
	return names
}

// BuildReleaseMap build a map of releases keyed by full names
func BuildReleaseMap(releases []terra.Release) map[string]terra.Release {
	m := make(map[string]terra.Release)
	for _, release := range releases {
		m[release.FullName()] = release
	}
	return m
}
