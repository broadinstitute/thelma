package statebucket

// Override represents configuration overrides for a release in an environment
type Override struct {
	Enabled             *bool  `json:"enabled,omitempty"`
	AppVersion          string `json:"appVersion"`
	ChartVersion        string `json:"chartVersion"`
	TerraHelmfileRef    string `json:"terraHelmfileRef"`
	FirecloudDevelopRef string `json:"firecloudDevelopRef"`
}

func (o *Override) SetAppVersion(version string) {
	o.AppVersion = version
}

func (o *Override) UnsetAppVersion() {
	o.AppVersion = ""
}

func (o *Override) SetChartVersion(version string) {
	o.ChartVersion = version
}

func (o *Override) UnsetChartVersion() {
	o.ChartVersion = ""
}

func (o *Override) SetTerraHelmfileRef(ref string) {
	o.TerraHelmfileRef = ref
}

func (o *Override) UnsetTerraHelmfileRef() {
	o.TerraHelmfileRef = ""
}

func (o *Override) SetFirecloudDevelopRef(ref string) {
	o.FirecloudDevelopRef = ref
}

func (o *Override) UnsetFirecloudDevelopRef() {
	o.FirecloudDevelopRef = ""
}

func (o *Override) UnsetAll() {
	o.UnsetAppVersion()
	o.UnsetChartVersion()
	o.UnsetTerraHelmfileRef()
	o.UnsetFirecloudDevelopRef()
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
