package argocd

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"strings"
)

const delimiter = "-"
const configsName = "configs"

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

// releaseSelector returns set of selectors for all argo apps associated with a release
// (often just the primary application, but can include the legacy configs application as well)
func releaseSelector(release terra.Release) map[string]string {
	if release.IsAppRelease() {
		return map[string]string{
			"app": release.Name(),
			"env": release.Destination().Name(),
		}
	} else {
		return map[string]string{
			"release": release.Name(),
			"cluster": release.Destination().Name(),
			"type":    "cluster",
		}
	}
}

// EnvironmentSelector returns set of selectors for all argo apps associated with an environment
func EnvironmentSelector(env terra.Environment) map[string]string {
	return map[string]string{
		"env": env.Name(),
	}
}

// joinSelector join map of label key-value pairs {"a":"b", "c":"d"} into selector string "a=b,c=d"
func joinSelector(labels map[string]string) string {
	var list []string
	for name, value := range labels {
		list = append(list, fmt.Sprintf("%s=%s", name, value))
	}
	return strings.Join(list, ",")
}
