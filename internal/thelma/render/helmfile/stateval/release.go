package stateval

import "github.com/broadinstitute/thelma/internal/thelma/terra"

// Release -- information related to the chart release that is being rendered
type Release struct {
	// Name of this release
	Name string `yaml:"Name"`
	// Type of this release
	Type string `yaml:"Type"`
	// Namespace this release is being deployed to
	Namespace string `yaml:"Namespace"`
	// AppVersion version of the application that's being deployed (only included for app releases)
	AppVersion string `yaml:"AppVersion,omitempty"`
}

func forRelease(release terra.Release) Release {
	// app version is omitted for cluster releases
	var appVersion string
	if release.IsAppRelease() {
		appVersion = release.(terra.AppRelease).AppVersion()
	}

	return Release{
		Name:       release.Name(),
		Type:       release.Type().String(),
		Namespace:  release.Namespace(),
		AppVersion: appVersion,
	}
}
