package sherlock

import (
	"fmt"
	"strings"

	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

type release struct {
	name                string
	enabled             bool
	releaseType         terra.ReleaseType
	chartVersion        string
	chartName           string
	repo                string
	namespace           string
	cluster             terra.Cluster
	destination         terra.Destination
	helmfileRef         string
	firecloudDevelopRef string
	helmfileOverlays    []string
	appVersion          string
	subdomain           string
	protocol            string
	port                int
}

// FullName provides the entire name of the chart release, globally unique as enforced by Sherlock. Name provides
// a truncated name used by Thelma for brevity that is only unique across a destination.
func (r *release) FullName() string {
	return r.name
}

func (r *release) Name() string {
	// sherlock requires unique release names so they are of the from RELEASE_NAME-(ENV_NAME | CLUSTER_NAME)
	// depending on release type. For compatibility with terra-helmfile values file structure and mimicking behavior in the
	// gitops provider, the env or cluster suffix must be stripped here

	// if there are multiple '-' separators we only want to trim off the final one
	releaseName := strings.TrimSuffix(r.name, fmt.Sprintf("-%s", r.destination.Name()))
	return releaseName
}

func (r *release) Type() terra.ReleaseType {
	return r.releaseType
}

func (r *release) IsAppRelease() bool {
	return r.Type() == terra.AppReleaseType
}

func (r *release) IsClusterRelease() bool {
	return r.Type() == terra.ClusterReleaseType
}

func (r *release) ChartName() string {
	return r.chartName
}

func (r *release) ChartVersion() string {
	return r.chartVersion
}

func (r *release) Repo() string {
	return r.repo
}

func (r *release) Namespace() string {
	return r.namespace
}

func (r *release) Cluster() terra.Cluster {
	return r.cluster
}

func (r *release) ClusterName() string {
	return r.cluster.Name()
}

func (r *release) ClusterAddress() string {
	return r.cluster.Address()
}

func (r *release) Destination() terra.Destination {
	return r.destination
}

func (r *release) TerraHelmfileRef() string {
	return r.helmfileRef
}

func (r *release) FirecloudDevelopRef() string {
	return r.firecloudDevelopRef
}

func (r *release) HelmfileOverlays() []string {
	return r.helmfileOverlays
}

func (r *release) AppVersion() string {
	return r.appVersion
}

func (r *release) Environment() terra.Environment {
	if !r.IsAppRelease() {
		return nil
	}
	return r.destination.(terra.Environment)
}

func (r *release) Subdomain() string {
	if r.subdomain == "" {
		return r.chartName
	}
	return r.subdomain
}

func (r *release) Protocol() string {
	if r.protocol == "" {
		return "https"
	}
	return r.protocol
}

func (r *release) Port() int {
	if r.port == 0 {
		return 443
	}
	return r.port
}

func (r *release) Host() string {
	var components []string
	components = append(components, r.Subdomain())
	if r.Environment().NamePrefixesDomain() {
		components = append(components, r.Environment().Name())
	}

	if r.Environment().BaseDomain() != "" {
		components = append(components, r.Environment().BaseDomain())
	}
	return strings.Join(components, ".")
}

func (r *release) URL() string {
	return fmt.Sprintf("%s://%s", r.Protocol(), r.Host())
}
