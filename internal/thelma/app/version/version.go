package version

import "runtime"

// These variables are set using -ldflags at build time (see the build target in the Makefile)

// Version Thelma semantic version
var Version = "unset"

// GitSha git sha that was used to produce this version of Thelma
var GitSha = "unset"

// BuildTimestamp timestamp at which this version of Thelma was built
var BuildTimestamp = "unset"

// Manifest bundles build information into an object that can be rendered as JSON or YAML
type Manifest struct {
	Version        string
	GitSha         string
	Arch           string
	Os             string
	BuildTimestamp string
}

func GetManifest() Manifest {
	return Manifest{
		Version:        Version,
		GitSha:         GitSha,
		Os:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		BuildTimestamp: BuildTimestamp,
	}
}
