package terra

// VersionOverride represents version overrides for a release in an environment
type VersionOverride struct {
	AppVersion          string `json:"appVersion,omitempty" yaml:"appVersion,omitempty"`
	ChartVersion        string `json:"chartVersion,omitempty" yaml:"chartVersion,omitempty"`
	TerraHelmfileRef    string `json:"terraHelmfileRef,omitempty" yaml:"terraHelmfileRef,omitempty"`
	FirecloudDevelopRef string `json:"firecloudDevelopRef,omitempty" yaml:"firecloudDevelopRef,omitempty"`
}
