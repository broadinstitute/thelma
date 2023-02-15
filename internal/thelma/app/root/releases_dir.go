package root

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/version"
	"path"
)

const currentSymlink = "current"
const lockFile = ".installer.lk"

type ReleasesDir interface {
	// ReleasesRoot return path of root of releases dir (~/.thelma/releases)
	ReleasesRoot() string
	// CurrentSymlink return path of current release symlink (~/.thelma/releases/current)
	CurrentSymlink() string
	// ForVersion return path of release dir for specific version (~/.thelma/releases/v1.2.3)
	ForVersion(version string) string
	// ForCurrentVersion return path of release dir for current running thelma version (~/.thelma/releases/v1.2.3)
	ForCurrentVersion() string
	// LockFile return path to the release update lock file
	LockFile() string
}

type releasesDir struct {
	dir string
}

func (r releasesDir) ReleasesRoot() string {
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

func (r releasesDir) LockFile() string {
	return path.Join(r.dir, lockFile)
}
