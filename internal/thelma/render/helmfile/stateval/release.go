package stateval

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

// Release -- information related to the chart release that is being rendered
type Release struct {
	// Name of this release
	Name string `yaml:"Name"`
	// ChartName name of the chart that is being deployed
	ChartName string `yaml:"ChartName"`
	// Type of this release
	Type string `yaml:"Type"`
	// Namespace this release is being deployed to
	Namespace string `yaml:"Namespace"`
	// AppVersion version of the application that's being deployed (only included for app releases)
	AppVersion string `yaml:"AppVersion,omitempty"`
	// Overlays representing other sets of values files that should be included
	Overlays []string `yaml:"Overlays,omitempty"`
}

func forRelease(release terra.Release) Release {
	return Release{
		Name:       release.Name(),
		ChartName:  release.ChartName(),
		Type:       release.Type().String(),
		Namespace:  release.Namespace(),
		AppVersion: release.AppVersion(),
		Overlays:   release.HelmfileOverlays(),
	}
}
