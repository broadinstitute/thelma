package statefixtures

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

// FixtureData root type for a fixture definition file
type FixtureData struct {
	Clusters     []Cluster
	Environments []Environment
	Charts       []Chart
	Releases     []Release
}

type Cluster struct {
	Name             string
	Base             string
	Address          string
	Project          string
	Location         string
	RequireSuitable  bool
	TerraHelmfileRef string
}

type Environment struct {
	Name                 string
	Base                 string
	Template             string
	Lifecycle            terra.Lifecycle
	UniqueResourcePrefix string
	DefaultCluster       string
	RequireSuitable      bool
	TerraHelmfileRef     string
	Owner                string
	EnableJanitor        bool
}

type Chart struct {
	Name string
	Repo string
}

type Release struct {
	FullName         string
	Name             string
	Repo             string
	Chart            string
	Cluster          string
	Namespace        string
	Environment      string
	AppVersion       string
	ChartVersion     string
	TerraHelmfileRef string
	Port             int
	Protocol         string
	Subdomain        string
}

func (r Release) name() string {
	if r.Name != "" {
		return r.Name
	}
	return r.Chart
}

func (r Release) key() string {
	var dest string
	if r.Environment != "" {
		dest = r.Environment
	} else {
		dest = r.Cluster
	}
	return r.Chart + "-" + dest
}
