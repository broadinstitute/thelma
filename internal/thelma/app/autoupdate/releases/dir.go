package releases

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/manifest"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/app/scratch"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/broadinstitute/thelma/internal/thelma/utils/flock"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"path/filepath"
	"time"
)

const currentSymlink = "current"
const lockFile = ".installer.lk"

// CurrentReleaseSymlink returns the path to the current release symlink at ~/.thelma/releases/current
func CurrentReleaseSymlink(root root.Root) string {
	return path.Join(root.ReleasesDir(), currentSymlink)
}

// Dir represents the local directory where Thelma stores its releases ("~/.thelma/releases")
type Dir interface {
	// CurrentVersion returns the currently installed version of Thelma
	// Returns an error if the symlink does not exist, is broken, or another fs error occurs
	CurrentVersion() (string, error)
	// CurrentVersionMatches returns true if the currently installed version of
	// Thelma matches the given version.
	// Returns false if versions don't match, if there is no current version of Thelma (say,
	// during a fresh install), or if another error is encountered identifying the current version
	CurrentVersionMatches(version string) bool
	// UpdateCurrentReleaseSymlink atomically updates the current release symlink at ~/.thelma/releases/current
	// to point at a given installed release version
	UpdateCurrentReleaseSymlink(version string) error
	// CopyUnpackedArchive copies an unpacked release archive into the releases directory
	CopyUnpackedArchive(unpackDir string) error
	// WithInstallerLock executes a callback function while holding the release installer lock
	WithInstallerLock(fn func() error) error
}

func NewDir(dir string, scratch scratch.Scratch) Dir {
	return &releasesDir{
		dir:     dir,
		scratch: scratch,
	}
}

type releasesDir struct {
	dir     string
	scratch scratch.Scratch
}

func (r *releasesDir) CurrentVersion() (string, error) {
	symlink := r.currentSymlinkPath()
	resolved, err := filepath.EvalSymlinks(symlink)
	if err != nil {
		return "", fmt.Errorf("error resolving symlink %s: %v", symlink, err)
	}
	return path.Base(resolved), nil
}

func (r *releasesDir) CurrentVersionMatches(version string) bool {
	currentVersion, err := r.CurrentVersion()
	if err != nil {
		log.Debug().Err(err).Msgf("error resolving current version symlink")
		// failed to resolve current version; send false to indicate that we should attempt an upgrade anyway
		// (this happens during bootstrapping when the current symlink does not exist yet)
		return false
	}

	return currentVersion == version
}

func (r *releasesDir) UpdateCurrentReleaseSymlink(version string) error {
	// following procedure described here:
	// https://stackoverflow.com/a/58148921

	// create tmp dir
	tmpDir, err := r.scratch.Mkdir("releases")
	if err != nil {
		return fmt.Errorf("error updating %s symbolic link: %v", r.currentSymlinkPath(), err)
	}

	releaseDir := r.pathForVersion(version)

	// create a new symlink in tmp dir pointing at the release version directory
	tmpLink := path.Join(tmpDir, "current.tmp")
	if err = os.Symlink(releaseDir, tmpLink); err != nil {
		return fmt.Errorf("error updating %s symbolic link: %v", r.currentSymlinkPath(), err)
	}

	// then rename it on top of the existing symlink, which is atomic
	if err = os.Rename(tmpLink, r.currentSymlinkPath()); err != nil {
		return fmt.Errorf("error updating %s symbolic link: %v", r.currentSymlinkPath(), err)
	}

	return nil
}

func (r *releasesDir) CopyUnpackedArchive(unpackDir string) error {
	version, err := manifest.Version(unpackDir)
	if err != nil {
		return fmt.Errorf("error installing %s to release directory: %v", unpackDir, err)
	}

	targetDir := r.pathForVersion(version)

	exists, err := utils.FileExists(targetDir)
	if err != nil {
		return fmt.Errorf("error installing %s to release directory: %v", unpackDir, err)
	}

	if exists {
		log.Warn().Msgf("Release directory %s already exists; removing it", targetDir)
		if err = os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("error removing existing release directory %s: %v", targetDir, err)
		}
	}

	if err = os.Rename(unpackDir, targetDir); err != nil {
		return fmt.Errorf("error moving unpacked release archive %s to release directory %s: %v", unpackDir, targetDir, err)
	}

	return nil
}

func (r *releasesDir) WithInstallerLock(fn func() error) error {
	_lockFile := path.Join(r.dir, lockFile)
	locker := flock.NewLocker(_lockFile, func(options *flock.Options) {
		options.Timeout = 5 * time.Minute
		options.RetryInterval = 10 * time.Second
	})

	return locker.WithLock(func() error {
		// write pid to lock file after we obtain it, to help with debugging
		pidStr := fmt.Sprintf("%d", os.Getpid())
		if err := os.WriteFile(_lockFile, []byte(pidStr), 0644); err != nil {
			return fmt.Errorf("error updating lock file %s: %v", _lockFile, err)
		}

		return fn()
	})
}

// currentSymlinkPath returns the full path to the current release symlink, eg. ~/.thelma/releases/current
func (r *releasesDir) currentSymlinkPath() string {
	return path.Join(r.dir, currentSymlink)
}

// pathForVersion returns the full path to a given Thelma release, eg ~/.thelma/releases/v1.2.3
func (r *releasesDir) pathForVersion(version string) string {
	return path.Join(r.dir, version)
}
