package labels

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
)

var reservedLabelNames = set.NewStringSet("job")

// ForRelease returns a standard set of labels for a chart release
func ForRelease(release terra.Release) map[string]string {
	labels := make(map[string]string)
	labels["release_name"] = release.Name()
	labels["release_key"] = release.FullName()
	labels["release_type"] = release.Type().String()
	labels["release_chart"] = release.ChartName()
	labels["release_cluster"] = release.Cluster().Name()
	labels["release_namespace"] = release.Namespace()
	if release.Destination().IsEnvironment() {
		labels["release_env"] = release.Destination().Name()
	}
	return Merge(ForDestination(release.Destination()), labels)
}

// ForDestination returns a standard set of labels for a destination
func ForDestination(dest terra.Destination) map[string]string {
	labels := make(map[string]string)
	labels["destination_type"] = dest.Type().String()
	labels["destination_name"] = dest.Name()
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
