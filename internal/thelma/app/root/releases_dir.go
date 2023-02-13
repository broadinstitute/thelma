package root

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/version"
	"path"
)

const currentSymlink = "current"

type ReleasesDir interface {
	// Root return path of root of releases dir (~/.thelma/releases)
	Root() string
	// CurrentSymlink return path of current release symlink (~/.thelma/releases/current)
	CurrentSymlink() string
	// ForVersion return path of release dir for specific version (~/.thelma/releases/v1.2.3)
	ForVersion(version string) string
	// ForCurrentVersion return path of release dir for current running thelma version (~/.thelma/releases/v1.2.3)
	ForCurrentVersion() string
}

type releasesDir struct {
	dir string
}

func (r releasesDir) Root() string {
	return r.dir
}

func (r releasesDir) CurrentSymlink() string {
	return path.Join(r.dir, currentSymlink)
}

func (r releasesDir) ForVersion(version string) string {
	return path.Join(r.dir, version)
}

func (r releasesDir) ForCurrentVersion() string {
	return path.Join(r.dir, version.Version)
}
