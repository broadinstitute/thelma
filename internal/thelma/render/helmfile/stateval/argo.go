package stateval

import (
	"github.com/broadinstitute/thelma/internal/thelma/render/helmfile/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/terra"
)

// ArgoApp -- information about the Argo application that will be used to deploy this release
type ArgoApp struct {
	// ProjectName name of the ArgoCD project the release's Argo app will belong to
	ProjectName string `yaml:"ProjectName"`
	// ClusterName name of the cluster this release is being deployed to
	ClusterName string `yaml:"ClusterName"`
	// ClusterAddress address of the cluster this release is being deployed to
	ClusterAddress string `yaml:"ClusterAddress"`
}

// ArgoProject -- information about the ArgoProject that will be used to deploy this release
type ArgoProject struct {
	// ProjectName name of the ArgoCD project that is being rendered
	ProjectName string `yaml:"ProjectName"`
}

func forArgoApp(r terra.Release) ArgoApp {
	return ArgoApp{
		ProjectName:    argocd.GetProjectName(r.Destination()),
		ClusterName:    r.ClusterName(),
		ClusterAddress: r.ClusterAddress(),
	}
}

func forArgoProject(d terra.Destination) ArgoProject {
	return ArgoProject{
		ProjectName: argocd.GetProjectName(d),
	}
}
