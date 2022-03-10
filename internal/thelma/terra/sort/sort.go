package sort

import (
	"github.com/broadinstitute/thelma/internal/thelma/terra"
	"github.com/broadinstitute/thelma/internal/thelma/terra/compare"
	"sort"
)

func Releases(releases []terra.Release) {
	sort.Slice(releases, func(i, j int) bool {
		return compare.Releases(releases[i], releases[j]) < 0
	})
}

func Destinations(destinations []terra.Destination) {
	sort.Slice(destinations, func(i, j int) bool {
		return compare.Destinations(destinations[i], destinations[j]) < 0
	})
}

func Environments(environments []terra.Destination) {
	sort.Slice(environments, func(i, j int) bool {
		return compare.Destinations(environments[i], environments[j]) < 0
	})
}

func Clusters(clusters []terra.Destination) {
	sort.Slice(clusters, func(i, j int) bool {
		return compare.Destinations(clusters[i], clusters[j]) < 0
	})
}
