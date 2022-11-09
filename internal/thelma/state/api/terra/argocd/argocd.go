package argocd

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"strings"
)

const delimiter = "-"
const configsName = "configs"
const argocdAppInstanceLabel = "argocd.argoproj.io/instance"

// LegacyConfigsApplicationName name of the firecloud-develop application for a release, eg. cromwell-configs-dev
func LegacyConfigsApplicationName(release terra.Release) string {
	return strings.Join(
		[]string{
			release.Name(),
			configsName,
			release.Destination().Name(),
		}, delimiter)
}

// ApplicationName name of the primary argo application for a release, eg. cromwell-dev
func ApplicationName(release terra.Release) string {
	return strings.Join(
		[]string{
			release.Name(),
			release.Destination().Name(),
		}, delimiter)
}

// ProjectName name of the project for the release. eg. terra-dev, cluster-terra-dev
func ProjectName(destination terra.Destination) string {
	switch t := destination.(type) {
	case terra.Environment:
		return fmt.Sprintf("terra-%s", t.Name())
	case terra.Cluster:
		return fmt.Sprintf("cluster-%s", t.Name())
	default:
		panic(fmt.Errorf("error generating destination values file: unknown destination type %s: %v", destination.Type().String(), destination))
	}
}

// GeneratorName name of a destinations app generator, eg. "terra-dev-generator"
func GeneratorName(destination terra.Destination) string {
	projectName := ProjectName(destination)
	return fmt.Sprintf("%s-generator", projectName)
}

// ApplicationSelector returns a set of Kubernetes labels that selects for resources managed by the given Argo app,
// suitable for use with `kubectl get -l`
func ApplicationSelector(applicationName string) map[string]string {
	return map[string]string{
		argocdAppInstanceLabel: applicationName,
	}
}
