package statebucket

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

// Override represents configuration overrides for a release in an environment
type Override struct {
	Enabled  *bool                 `json:"enabled,omitempty" yaml:",omitempty"`
	Versions terra.VersionOverride `json:"versions,omitempty" yaml:",omitempty"`
}

// PinVersions applies the given VersionOverride to this override, ignoring empty fields in the parameter
func (o *Override) PinVersions(versions terra.VersionOverride) {
	if versions.AppVersion != "" {
		o.Versions.AppVersion = versions.AppVersion
	}
	if versions.ChartVersion != "" {
		o.Versions.ChartVersion = versions.ChartVersion
	}
	if versions.TerraHelmfileRef != "" {
		o.Versions.TerraHelmfileRef = versions.TerraHelmfileRef
	}
	if versions.FirecloudDevelopRef != "" {
		o.Versions.FirecloudDevelopRef = versions.FirecloudDevelopRef
	}
}

// UnpinVersions removes all version overrides
func (o *Override) UnpinVersions() {
	o.Versions = terra.VersionOverride{}
}

func (o *Override) HasEnableOverride() bool {
	return o.Enabled != nil
}

func (o *Override) IsEnabled() bool {
	if !o.HasEnableOverride() {
		return false
	}
	return *o.Enabled
}

func (o *Override) Enable() {
	enabled := true
	o.Enabled = &enabled
}

func (o *Override) Disable() {
	enabled := false
	o.Enabled = &enabled
}
