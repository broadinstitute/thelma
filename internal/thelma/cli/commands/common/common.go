package common

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

// ReleaseMapToStructuredView Given a map where releases are keys, return a data structure organizing the releases by name
// eg.
//
//	{
//	  "dev":
//	    "leonardo": {
//	    }
//	}
func ReleaseMapToStructuredView[T any](input map[terra.Release]T) map[string]map[string]T {
	output := make(map[string]map[string]T)
	for release, _status := range input {
		destName := release.Destination().Name()
		destMap, exists := output[destName]
		if !exists {
			destMap = make(map[string]T)
		}
		destMap[release.Name()] = _status
		output[destName] = destMap
	}
	return output
}
