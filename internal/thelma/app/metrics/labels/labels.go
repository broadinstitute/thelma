// Package labels contains utility functions for generating a standard set of labels for terra.State objects
// Note that you cannot record two metrics with the same name and a different set of labels; if you do, the
// prometheus client library will panic.
// For that reason, we supply empty labels ("env": "") where labels don't apply.
package labels

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
)

var reservedLabelNames = set.NewStringSet("job")

// ForReleaseOrDestination is for metrics that could apply to either a release or a destination.
// (manifest rendering jobs are probably the only use case for this).
func ForReleaseOrDestination(value interface{}, extra ...map[string]string) map[string]string {
	var labels map[string]string

	switch t := value.(type) {
	case terra.Release:
		labels = ForRelease(t)
	case terra.Destination:
		labels = Merge(
			map[string]string{"release": ""},
			ForDestination(t),
		)
	default:
		panic(fmt.Errorf("unexpected type: %#v", t))
	}

	var all []map[string]string
	all = append(all, labels)
	all = append(all, extra...)
	return Merge(all...)
}

// ForRelease returns a standard set of labels for a chart release.
// For example:
//
//	{
//	  "release": "leonardo",
//	  "env": "dev",
//	  "cluster": "terra-dev",
//	}
func ForRelease(release terra.Release) map[string]string {
	labels := make(map[string]string)
	labels["release"] = release.Name()
	labels["cluster"] = release.Cluster().Name()
	if release.Destination().IsEnvironment() {
		labels["env"] = release.Destination().Name()
	} else {
		labels["env"] = ""
	}
	return labels
}

// ForDestination returns a standard set of labels for a destination.
// For example, for the dev env:
//
//	{
//	  "env": "dev",
//	  "cluster": "",
//	}
//
// For the terra-dev-cluster:
//
//	{
//	  "env": "",
//	  "cluster": "terra-dev",
//	}
func ForDestination(dest terra.Destination) map[string]string {
	labels := make(map[string]string)
	if dest.IsCluster() {
		labels["cluster"] = dest.Name()
		labels["env"] = ""
	} else if dest.IsEnvironment() {
		labels["cluster"] = ""
		labels["env"] = dest.Name()
	}
	return labels
}

func Normalize(labels map[string]string) map[string]string {
	normalized := make(map[string]string)
	for k, v := range labels {
		if reservedLabelNames.Exists(k) {
			k = "_" + k
		}
		normalized[k] = v
	}
	return normalized
}

// Merge N maps into a single map (last takes precedence)
func Merge(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		if m == nil {
			// ignore nil maps
			continue
		}
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
