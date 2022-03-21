package argocd

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"sort"
	"strings"
)

//
// Note that Argo has a concept call `destinations` that is similar to terra.Destination
// in that it means "cluster and namespace where my Argo app deploys manifests"
//

type Values struct {
	Destinations []ArgoDestination `yaml:"destinations"`
}

type ArgoDestination struct {
	Server    string `yaml:"server"`
	Namespace string `yaml:"namespace"`
}

// Generate a set of values with a list of ArgoCD project destinations, in the form:
//
// destinations:
//   - server: https://<cluster api address>/
//     namespace: namespace1
//   - server: https://<cluster api address>/
//     namespace: namespace2
func GetDestinationValues(destination terra.Destination) Values {
	destinations := getArgoDestinations(destination)

	// Sort destinations so they always render in a consistent order
	sort.Slice(destinations, func(i, j int) bool {
		return destinations[i].compare(destinations[j]) < 0
	})

	return Values{Destinations: destinations}
}

func GetProjectName(destination terra.Destination) string {
	switch t := destination.(type) {
	case terra.Environment:
		return fmt.Sprintf("terra-%s", t.Name())
	case terra.Cluster:
		return fmt.Sprintf("cluster-%s", t.Name())
	default:
		panic(fmt.Errorf("error generating destination values file: unknown destination type %s: %v", destination.Type().String(), destination))
	}
}

func getArgoDestinations(destination terra.Destination) []ArgoDestination {
	switch t := destination.(type) {
	case terra.Environment:
		return argoDestinationsForEnvironment(t)
	case terra.Cluster:
		return argoDestinationsForCluster(t)
	default:
		panic(fmt.Errorf("error generating destination values file: unknown destination type %s: %v", destination.Type().String(), destination))
	}
}

func argoDestinationsForCluster(cluster terra.Cluster) []ArgoDestination {
	return []ArgoDestination{
		{
			Server:    cluster.Address(),
			Namespace: "*", // Cluster releases can deploy to any namespace
		},
	}
}

func argoDestinationsForEnvironment(environment terra.Environment) []ArgoDestination {
	clusterAddresses := make(map[string]bool)
	for _, release := range environment.Releases() {
		clusterAddresses[release.ClusterAddress()] = true
	}

	var destinations []ArgoDestination
	for address := range clusterAddresses {
		destinations = append(destinations, ArgoDestination{
			Server:    address,
			Namespace: environment.Namespace(),
		})
	}

	return destinations
}

// Return -1 if d < other, 0 if d == other, +1 if d > other
func (d ArgoDestination) compare(other ArgoDestination) int {
	byServer := strings.Compare(d.Server, other.Server)
	if byServer != 0 {
		return byServer
	}

	byNamespace := strings.Compare(d.Namespace, other.Namespace)
	return byNamespace
}
