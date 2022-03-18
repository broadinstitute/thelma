// Package stateval is used for generating Helmfile state values.
// State values are consumed in both helmfile.yaml and values.yaml.gotmpl files.
//
// Note: We serialize yaml keys with upper case name for greater readability in Go templates.
// (so that we can use .Values.Release.ChartPath and not .Values.release.chartPath)
package stateval

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

// AppValues -- the full set of helmfile state values for rendering application manifests
// (used by $THELMA_HOME/helmfile.yaml)
type AppValues struct {
	// Release the release that is being rendered
	Release Release `yaml:"Release"`
	// ChartPath filesystem path for the chart that is being rendered
	ChartPath string `yaml:"ChartPath"`
	// Destination destination where the release is being deployed
	Destination Destination `yaml:"Destination"`
	// Environment environment where the release is being deployed (for app releases only)
	Environment Environment `yaml:"Environment,omitempty"`
	// Cluster cluster where the release is being deployed (for cluster releases only)
	Cluster Cluster `yaml:"Cluster,omitempty"`
}

// ArgoAppValues -- the full set of helmfile state values for rendering argo apps
// (used by $THELMA_HOME/argocd/application/helmfile.yaml)
type ArgoAppValues struct {
	// Release the release this Argo app will deploy
	Release Release `yaml:"Release"`
	// Destination destination where this Argo app will deploy the release to
	Destination Destination `yaml:"Destination"`
	// ArgoApp information about the cluster and project the ArgoApp will deploy to
	ArgoApp ArgoApp `yaml:"ArgoApp"`
	// Environment environment where the release is being deployed (for app releases only)
	Environment Environment `yaml:"Environment,omitempty"`
	// Cluster cluster where the release is being deployed (for cluster releases only)
	Cluster Cluster `yaml:"Cluster,omitempty"`
}

// ArgoProjectValues -- the full set of helmfile state values for rendering argo projects
// (used by $THELMA_HOME/argocd/projects/helmfile.yaml)
type ArgoProjectValues struct {
	// Destination environment or cluster that apps in this project will deploy to
	Destination Destination `yaml:"Destination"`
	// ArgoProject information about the Argo project that is being rendered
	ArgoProject ArgoProject `yaml:"ArgoProject"`
}

// BuildAppValues generates an AppValues for the given release
func BuildAppValues(r terra.Release, chartPath string) AppValues {
	values := AppValues{
		Release:     forRelease(r),
		ChartPath:   chartPath,
		Destination: forDestination(r.Destination()),
	}
	if r.Destination().IsEnvironment() {
		values.Environment = forEnvironment(r.Destination().(terra.Environment))
	}
	if r.Destination().IsCluster() {
		values.Cluster = forCluster(r.Destination().(terra.Cluster))
	}
	return values
}

// BuildArgoAppValues generates an ArgoAppValues for the given release
func BuildArgoAppValues(r terra.Release) ArgoAppValues {
	values := ArgoAppValues{
		Release:     forRelease(r),
		Destination: forDestination(r.Destination()),
		ArgoApp:     forArgoApp(r),
	}
	if r.Destination().IsEnvironment() {
		values.Environment = forEnvironment(r.Destination().(terra.Environment))
	}
	if r.Destination().IsCluster() {
		values.Cluster = forCluster(r.Destination().(terra.Cluster))
	}
	return values
}

// BuildArgoProjectValues genreates an ArgoProjectValues for the given destination
func BuildArgoProjectValues(d terra.Destination) ArgoProjectValues {
	return ArgoProjectValues{
		Destination: forDestination(d),
		ArgoProject: forArgoProject(d),
	}
}
