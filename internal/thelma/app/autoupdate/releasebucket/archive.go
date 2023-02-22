package releasebucket

import (
	"fmt"
	"path"
	"runtime"
)

const releaseArchiveObjectPrefix = "releases"

// Archive represents a release archive in the Thelma releases bucket
type Archive struct {
	version string
	os      string
	arch    string
}

func NewArchive(version string, osArch ...string) Archive {
	os := runtime.GOOS
	arch := runtime.GOARCH

	if len(osArch) >= 1 {
		os = osArch[0]
	}
	if len(osArch) >= 2 {
		arch = osArch[1]
	}

	return Archive{
		version: version,
		os:      os,
		arch:    arch,
	}
}

// Version return the semantic version for this archive
// eg. "v1.2.3"
func (a *Archive) Version() string {
	return a.version
}

// ObjectPath returns the path to this release archive object in the bucket.
// eg. "releases/v.1.2.3/thelma_v1.2.3_linux_amd64.tar.gz"
func (a *Archive) ObjectPath() string {
	return path.Join(releaseArchiveObjectPrefix, a.version, a.Filename())
}

// Filename returns the filename for this release archive
// eg. "thelma_v1.2.3_linux_amd64.tar.gz"
func (a *Archive) Filename() string {
	return fmt.Sprintf("thelma_%s_%s_%s.tar.gz", a.version, a.os, a.arch)
}

// Sha256SumObjectPath returns the path to a sha256sum file for this release archive
// eg. "releases/v.1.2.3/thelma_v1.2.3_SHA256SUMS"
func (a *Archive) Sha256SumObjectPath() string {
	return path.Join(releaseArchiveObjectPrefix, a.version, fmt.Sprintf("thelma_%s_SHA256SUMS", a.version))
}
