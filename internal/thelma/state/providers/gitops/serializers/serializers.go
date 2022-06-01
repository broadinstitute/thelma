package serializers

type Environment struct {
	DefaultCluster  string                `yaml:"defaultCluster"`
	Releases        map[string]AppRelease `yaml:"releases"`
	Lifecycle       string                `yaml:"lifecycle"`
	RequireSuitable bool                  `yaml:"requireSuitable"`
}

type Cluster struct {
	Address         string                    `yaml:"address"`
	Project         string                    `yaml:"project"`
	Location        string                    `yaml:"location"`
	RequireSuitable bool                      `yaml:"requireSuitable"`
	Releases        map[string]ClusterRelease `yaml:"releases"`
}

type ClusterRelease struct {
	Namespace    string `yaml:"namespace"`
	Enabled      bool   `yaml:"enabled"`
	ChartName    string `yaml:"chartName"`
	ChartVersion string `yaml:"chartVersion"`
	Repo         string `yaml:"repo"`
}

type AppRelease struct {
	AppVersion   string `yaml:"appVersion"`
	Cluster      string `yaml:"cluster"`
	Enabled      bool   `yaml:"enabled"`
	ChartName    string `yaml:"chartName"`
	ChartVersion string `yaml:"chartVersion"`
	Repo         string `yaml:"repo"`
}
