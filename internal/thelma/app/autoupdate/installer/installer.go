package installer

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/releasebucket"
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/releases"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
)

// ResolvedVersions contains resolved version information for Thelma releases
type ResolvedVersions struct {
	// VersionAlias tag or version string that was resolved (eg. "latest", "1.2.3", "v1.2.3")
	VersionAlias string
	// TargetVersion the resolved canonical version of Thelma that should be installed (eg. "v1.2.3")
	TargetVersion string
	// CurrentVersion locally-installed version of Thelma; may be "" if this is
	// a bootstrap/fresh install
	CurrentVersion string
}

// UpdateNeeded return true if Thelma should be updated
func (v ResolvedVersions) UpdateNeeded() bool {
	return v.CurrentVersion != v.TargetVersion
}

type Installer interface {
	// ResolveVersions expand a version or tag to a canonical Thelma version and identify
	// the currently-installed Thelma version
	ResolveVersions(versionOrTag string) (ResolvedVersions, error)
	// UpdateThelma perform a Thelma update, obtaining a file lock first so multiple processes don't
	// step on each other
	UpdateThelma(versionOrTag string) error
}

type Options struct {
	// KeepReleases number of old releases to keep in ~/.thelma/releases directory
	KeepReleases int
}

func New(releasesDir releases.Dir, bucket releasebucket.ReleaseBucket, options ...func(*Options)) Installer {
	return &installer{
		dir:     releasesDir,
		bucket:  bucket,
		options: utils.CollateOptions(options...),
	}
}

type installer struct {
	dir     releases.Dir
	bucket  releasebucket.ReleaseBucket
	options Options
}

func (i *installer) ResolveVersions(versionOrTag string) (ResolvedVersions, error) {
	targetVersion, err := i.bucket.ResolveTagOrVersion(versionOrTag)
	if err != nil {
		return ResolvedVersions{}, err
	}

	currentVersion, err := i.dir.CurrentVersion()
	if err != nil {
		log.Debug().Err(err).Msgf("Could not identify current installed version of Thelma; is this a fresh install?")
	}

	return ResolvedVersions{
		VersionAlias:   versionOrTag,
		CurrentVersion: currentVersion,
		TargetVersion:  targetVersion,
	}, nil
}

func (i *installer) UpdateThelma(versionOrTag string) error {
	resolved, err := i.ResolveVersions(versionOrTag)
	if err != nil {
		return fmt.Errorf("error updating Thelma: %v", err)
	}

	if !resolved.UpdateNeeded() {
		log.Info().Msgf("Thelma is already up-to-date (current version %q matches %q)", resolved.CurrentVersion, resolved.VersionAlias)
		return nil
	}

	return i.dir.WithInstallerLock(func() error {
		return i.updateThelmaUnsafe(versionOrTag)
	})
}

// perform a Thelma update (without a lock)
func (i *installer) updateThelmaUnsafe(versionOrTag string) error {
	// resolve versions once more to make sure an update is actually needed, in case
	// (a) someone else finished an update while we were waiting for the lock
	// (b) a new version of Thelma was released
	resolved, err := i.ResolveVersions(versionOrTag)
	if err != nil {
		return fmt.Errorf("error updating Thelma: %v", err)
	}

	if !resolved.UpdateNeeded() {
		log.Info().Msgf("Thelma was updated to %s in the background; nothing to do", resolved.CurrentVersion)
		return nil
	}

	targetVersion := resolved.TargetVersion
	priorVersion := resolved.CurrentVersion
	if priorVersion == "" {
		log.Debug().Err(err).Msgf("Could not identify current installed version of Thelma; is this a fresh install?")
		priorVersion = "unknown"
	}
	log.Info().Msgf("Updating Thelma from %s to %s...", priorVersion, targetVersion)

	targetArchive := releasebucket.NewArchive(targetVersion)
	unpackDir, err := i.bucket.DownloadAndUnpack(targetArchive)
	if err != nil {
		return fmt.Errorf("error installing Thelma %s: %v", targetVersion, err)
	}

	if err = i.dir.CopyUnpackedArchive(unpackDir); err != nil {
		return fmt.Errorf("error installing Thelma %s: %v", targetVersion, err)
	}

	if err = i.dir.UpdateCurrentReleaseSymlink(targetVersion); err != nil {
		return fmt.Errorf("error installing Thelma %s: %v", targetVersion, err)
	}

	if err = i.dir.CleanupOldReleases(i.options.KeepReleases); err != nil {
		return fmt.Errorf("error cleaning up release directory: %v", err)
	}

	log.Info().Msgf("Thelma has been updated to %s", targetVersion)
	return nil
}
