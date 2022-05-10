package terra

// VersionOverride can be used to set and unset version overrides for a release
type VersionOverride interface {
	SetAppVersion(version string)
	UnsetAppVersion()
	SetChartVersion(version string)
	UnsetChartVersion()
	SetFirecloudDevelopRef(ref string)
	UnsetFirecloudDevelopRef()
	SetTerraHelmfileRef(ref string)
	UnsetTerraHelmfileRef()
	UnsetAll()
}
