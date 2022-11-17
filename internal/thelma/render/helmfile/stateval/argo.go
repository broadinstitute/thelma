package stateval

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/argocd"
)

// ArgoApp -- information about the Argo application that will be used to deploy this release
type ArgoApp struct {
	// ProjectName name of the ArgoCD project the release's Argo app will belong to
	ProjectName string `yaml:"ProjectName"`
	// ClusterName name of the cluster this release is being deployed to
	ClusterName string `yaml:"ClusterName"`
	// ClusterAddress address of the cluster this release is being deployed to
	ClusterAddress string `yaml:"ClusterAddress"`
	// TerraHelmfileRef override terra-helmfile ref
	TerraHelmfileRef string `yaml:"TerraHelmfileRef"`
	// FirecloudDevelopRef override firecloud-develop ref
	FirecloudDevelopRef string `yaml:"FirecloudDevelopRef"`
}

// ArgoProject -- information about the ArgoProject that will be used to deploy this release
type ArgoProject struct {
	// ProjectName name of the ArgoCD project that is being rendered
	ProjectName string `yaml:"ProjectName"`
	// Generator settings for the project's application generator, if enabled
	Generator Generator `yaml:"Generator"`
}

// Generator information about an Argo project's application generator
type Generator struct {
	// Name name of the project's application generator
	Name string `yaml:"Name"`
	// TerraHelmfileRef override terra-helmfile ref for the project's app generator
	TerraHelmfileRef string `yaml:"TerraHelmfileRef"`
}

func forArgoApp(r terra.Release) ArgoApp {
	return ArgoApp{
		ProjectName:         argocd.ProjectName(r.Destination()),
		ClusterName:         r.ClusterName(),
		ClusterAddress:      r.ClusterAddress(),
		TerraHelmfileRef:    r.TerraHelmfileRef(),
		FirecloudDevelopRef: r.FirecloudDevelopRef(),
	}
}

func forArgoProject(d terra.Destination) ArgoProject {
	return ArgoProject{
		ProjectName: argocd.ProjectName(d),
		Generator: Generator{
			Name:             argocd.GeneratorName(d),
			TerraHelmfileRef: d.TerraHelmfileRef(),
		},
	}
}
