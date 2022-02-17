package gitops

import "github.com/broadinstitute/thelma/internal/thelma/terra"

type appRelease struct {
	appVersion string
	release
}

func (r *appRelease) AppVersion() string {
	return r.appVersion
}

func (r *appRelease) Environment() terra.Environment {
	return r.destination.(terra.Environment)
}
