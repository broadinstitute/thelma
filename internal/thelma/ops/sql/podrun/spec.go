package podrun

import v1 "k8s.io/api/core/v1"

// Spec settings for a pod that will be provisioned by the K8s pod runner
type Spec struct {
	DBMSSpec
	ProviderSpec
}

// DBMSSpec DBMS-specific pod settings
type DBMSSpec struct {
	// ContainerImage container image to use
	ContainerImage string
	// Env environment variables to set (stored in K8s secret)
	Env map[string]string
	// Scripts to mount on the pod (stored in K8s secret)
	Scripts map[string][]byte
	// ScriptsMount mount point for scripts
	ScriptsMount string
}

// ProviderSpec platform-dependent pod settings
type ProviderSpec struct {
	// Sidecar optional sidecar to add to pod
	Sidecar *v1.Container
	// ServiceAccount Kubernetes service account to use for running pod
	ServiceAccount string
}
