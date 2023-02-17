// Package installer implements Thelma's self-install and self-update features
package installer

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/installer/bootstrap"
	"github.com/broadinstitute/thelma/internal/thelma/app/installer/spawn"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/app/scratch"
	"github.com/broadinstitute/thelma/internal/thelma/clients/api"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/broadinstitute/thelma/internal/thelma/utils/flock"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
	"golang.org/x/mod/semver"
	"io"
	"k8s.io/utils/strings/slices"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const configKey = "autoupdate"
const tagsFile = "tags.json"
const buildManifest = "build.json"
const releasesObjectPrefix = "releases"
const updateCommandName = "update"

type updateConfig struct {
	Enabled bool   `default:"true"`            // if false, do not perform automatic updates
	Tag     string `default:"latest"`          // which Thelma build tag to follow for auto-updates
	Bucket  string `default:"thelma-releases"` // name of GCS bucket that contains Thelma releases
}

type Installer interface {
	// Update performs a foreground update of Thelma to the current version of its configured tag
	Update() error
	// UpdateTo performs a foreground update of Thelma to a specific version
	UpdateTo(version string) error
	// StartBackgroundUpdateIfEnabled start a new process that will update Thelma in the background
	StartBackgroundUpdateIfEnabled() error
	// Bootstrap set up a new Thelma installation, including generating shell scripts
	Bootstrap() error
}

func New(thelmaConfig config.Config, bucketFactory api.BucketFactory, root root.Root, runner shell.Runner, scratch scratch.Scratch) (Installer, error) {
	var cfg updateConfig

	if err := thelmaConfig.Unmarshal(configKey, &cfg); err != nil {
		return nil, err
	}

	_bucket, err := bucketFactory.Bucket(cfg.Bucket)
	if err != nil {
		return nil, err
	}

	releasesDir := root.ReleasesDir()

	locker := flock.NewLocker(releasesDir.LockFile(), func(options *flock.Options) {
		options.Timeout = 5 * time.Minute
		options.RetryInterval = 10 * time.Second
	})

	bootstrapper := bootstrap.New(root, thelmaConfig, runner)

	return &installer{
		config:       cfg,
		bucket:       _bucket,
		releasesDir:  root.ReleasesDir(),
		shellRunner:  runner,
		locker:       locker,
		scratch:      scratch,
		bootstrapper: bootstrapper,
	}, nil
}

type installer struct {
	config       updateConfig
	bucket       bucket.Bucket
	shellRunner  shell.Runner
	releasesDir  root.ReleasesDir
	locker       flock.Locker
	scratch      scratch.Scratch
	bootstrapper bootstrap.Bootstrapper
}

func (a *installer) Update() error {
	needsUpdate, err := a.isUpdateNeeded()
	if err != nil {
		return err
	}

	if !needsUpdate {
		currentVersion, err := a.currentVersion()
		if err != nil {
			return err
		}
		log.Info().Msgf("Thelma is up-to-date (current version %s == %s)", currentVersion, a.config.Tag)
		return nil
	}

	return a.updateThelma(a.config.Tag)
}

func (a *installer) UpdateTo(versionOrTag string) error {
	if a.config.Enabled {
		return fmt.Errorf("auto-update is enabled; please disable by " +
			"setting THELMA_AUTOUPDATE_ENABLED=false or adding installer.enabled=false " +
			"to ~/.thelma/config.yaml and re-run (otherwise this change will be overwritten!)")
	}

	return a.updateThelma(versionOrTag)
}

func (a *installer) StartBackgroundUpdateIfEnabled() error {
	if !a.config.Enabled {
		return nil
	}

	_spawn := spawn.New()
	if _spawn.CurrentProcessIsSpawn() {
		// never start a background process from a background process (infinite loops are bad)
		return nil
	}

	if currentProcessIsThelmaUpdateCommand() {
		// don't trigger a background update from a manually-run `thelma update` command
		return nil
	}

	needsUpdate, err := a.isUpdateNeeded()
	if err != nil {
		return fmt.Errorf("error starting background Thelma updates: %v", err)
	}

	if !needsUpdate {
		return nil
	}

	// launch "thelma update" in the background
	return _spawn.Spawn(updateCommandName)
}

func (a *installer) Bootstrap() error {
	return a.bootstrapper.Bootstrap()
}

// perform a Thelma update, obtaining a file lock first so multiple processes don't
// step on each other.
func (a *installer) updateThelma(versionOrTag string) error {
	needsUpdate, err := a.isUpdateNeededFor(versionOrTag)
	if err != nil {
		return err
	}

	if !needsUpdate {
		currentVersion, err := a.currentVersion()
		if err != nil {
			return err
		}
		log.Info().Msgf("Thelma is already up-to-date (current version %q matches %q)", currentVersion, versionOrTag)
		return nil
	}

	return a.withInstallerLock(func() error {
		return a.updateThelmaUnsafe(versionOrTag)
	})
}

// obtain the installer lock and write pid to lockfile,
func (a *installer) withInstallerLock(inner func() error) error {
	return a.locker.WithLock(func() error {
		// write pid to lock file after we obtain it, to help with debugging
		pidStr := fmt.Sprintf("%d", os.Getpid())
		if err := os.WriteFile(a.releasesDir.LockFile(), []byte(pidStr), 0644); err != nil {
			return fmt.Errorf("error updating lock file %s: %v", a.releasesDir.LockFile(), err)
		}

		return inner()
	})
}

// perform a Thelma update (without a lock)
func (a *installer) updateThelmaUnsafe(versionOrTag string) error {
	// check once more to make sure an update is actually needed, in case someone else
	// finished an update while we were waiting for the lock
	needsUpdate, err := a.isUpdateNeededFor(versionOrTag)
	if err != nil {
		return err
	}

	if !needsUpdate {
		currentVersion, err := a.currentVersion()
		if err != nil {
			return err
		}
		log.Info().Msgf("Thelma was updated to %s in the background; nothing to do", currentVersion)
		return nil
	}

	targetVersion, err := a.resolveTagToVersion(versionOrTag)
	if err != nil {
		return err
	}

	priorVersion, err := a.currentVersion()
	if err != nil {
		log.Debug().Err(err).Msgf("Could not identify current installed version of Thelma; is this a fresh install?")
		priorVersion = "unknown"
	}
	log.Info().Msgf("Updating Thelma from %s to %s...", priorVersion, targetVersion)
	time.Sleep(30 * time.Second)
	scratchDir, err := a.scratch.Mkdir("installer")
	if err != nil {
		return err
	}

	if err = a.verifyReleaseArchiveExists(targetVersion); err != nil {
		return err
	}

	if err = a.installThelmaReleaseArchive(targetVersion, scratchDir); err != nil {
		return fmt.Errorf("error installing Thelma version %s to release directory: %v", targetVersion, err)
	}

	if err = a.updateCurrentReleaseSymlink(targetVersion, scratchDir); err != nil {
		return err
	}

	log.Info().Msgf("Thelma has been updated to %s", targetVersion)
	return nil
}

// atomically update the ~/.thelma/releases/current symlink to point to the new version
func (a *installer) updateCurrentReleaseSymlink(version string, scratchDir string) error {
	tmpLink := path.Join(scratchDir, "current.tmp")

	// create a new symlink in a tempdir pointing at the release version directory
	if err := os.Symlink(a.releasesDir.ForVersion(version), tmpLink); err != nil {
		return fmt.Errorf("error updating %s symbolic link: %v", a.releasesDir.CurrentSymlink(), err)
	}

	// then rename it on top of the existing symlink, which is atomic
	if err := os.Rename(tmpLink, a.releasesDir.CurrentSymlink()); err != nil {
		return fmt.Errorf("error updating %s symbolic link: %v", a.releasesDir.CurrentSymlink(), err)
	}

	return nil
}

// download Thelma release and install to ~/.thelma/releases/<version>
func (a *installer) installThelmaReleaseArchive(version string, scratchDir string) error {
	releaseDir := a.releasesDir.ForVersion(version)
	exists, err := utils.FileExists(releaseDir)
	if err != nil {
		return fmt.Errorf("error downloading release to %s: %v", releaseDir, err)
	}
	if exists {
		if err = a.verifyBuildManifestMatches(releaseDir, version); err != nil {
			log.Warn().Msgf("Release directory %s exists but has incorrect build manifest; will re-install version %s of Thelma", releaseDir, version)
			if err = os.RemoveAll(releaseDir); err != nil {
				return fmt.Errorf("error cleaning up broken release directory %s: %v", releaseDir, err)
			}
		} else {
			log.Info().Msgf("Not downloading Thelma version %s; it already exists at %s", version, releaseDir)
			return nil
		}
	}

	localArchive := path.Join(scratchDir, releaseArchiveFilename(version))
	objectName := releaseArchiveObject(version)

	if err = a.bucket.Download(objectName, localArchive); err != nil {
		return fmt.Errorf("error downloading release archive gs://%s/%s to %s: %v", a.bucket.Name(), objectName, localArchive, err)
	}

	if err = a.verifySha256Sum(localArchive, version); err != nil {
		return err
	}

	unpackDir := path.Join(scratchDir, version)
	if err = os.MkdirAll(unpackDir, 0755); err != nil {
		return fmt.Errorf("error unpacking release archive %s: %v", localArchive, err)
	}

	err = a.shellRunner.Run(shell.Command{
		Prog: "tar",
		Args: []string{
			"-xz",
			"-C",
			unpackDir,
			"-f",
			localArchive,
		},
	})
	if err != nil {
		return fmt.Errorf("error unpacking release archive %s: %v", localArchive, err)
	}

	if err = a.verifyBuildManifestMatches(unpackDir, version); err != nil {
		return err
	}

	if err = os.Rename(unpackDir, a.releasesDir.ForVersion(version)); err != nil {
		return fmt.Errorf("error moving unpacked release archive to releases dir: %v", err)
	}

	return nil
}

func (a *installer) verifyReleaseArchiveExists(version string) error {
	exists, err := a.bucket.Exists(releaseArchiveObject(version))
	if err != nil {
		return fmt.Errorf("error validating release version %s: %v", version, err)
	}
	if !exists {
		return fmt.Errorf("%q does not match a known Thelma release", version)
	}
	return nil
}

func (a *installer) verifyBuildManifestMatches(releaseUnpackDir string, version string) error {
	manifestFile := path.Join(releaseUnpackDir, buildManifest)
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return fmt.Errorf("error reading build manifest %s: %v", manifestFile, err)
	}

	type manifest struct {
		Version string `json:"version"`
	}
	var m manifest
	if err = json.Unmarshal(content, &m); err != nil {
		return fmt.Errorf("error parsing build manifest %s: %v", manifestFile, err)
	}
	if m.Version == "" {
		return fmt.Errorf("error parsing build manifest %s (version not found): %v", manifestFile, err)
	}

	if m.Version != version {
		return fmt.Errorf("release verification error: %s: build manifest version %s does not match desired Thelma version %s", manifestFile, m.Version, version)
	}

	return nil
}

func (a *installer) verifySha256Sum(file string, version string) error {
	archiveSha256Sum, err := a.getReleaseArchiveSha256Sum(version)
	if err != nil {
		return fmt.Errorf("error identifying sha256sum for Thelma version %s: %v", version, err)
	}

	localSha256Sum, err := computeSha256Sum(file)
	if err != nil {
		return err
	}
	if localSha256Sum != archiveSha256Sum {
		return fmt.Errorf("downloaded release archive %s has incorrect sha256sum (has %s, should be %s)", file, localSha256Sum, archiveSha256Sum)
	}
	return nil
}

// compute the sha256sum for a local file
func computeSha256Sum(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("error computing sha256sum for %s: %v", file, err)
	}

	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", fmt.Errorf("error computing sha256sum for %s: %v", file, err)
	}

	if err = f.Close(); err != nil {
		return "", fmt.Errorf("error computing sha256sum for %s: %v", file, err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// fetch the published sha256sum for this release archive and return it.
// these are kept in the thelma release archive directory at
//
//	releases/<version>/thelma_<version>_SHA256SUMS
//
// and the file looks like:
//
// 5b0fac41f493099924dbcbcc40ac7b2d61d342e5044b2e2ffd4b771863756a65  thelma_<version>_darwin_amd64.tar.gz
// e50372e48fa3750a917a4e61c67c456015a016f1beac1e55d241349eb44d266d  thelma_<version>_darwin_arm64.tar.gz
// a9afa31857e0b9b8206202a3f31f6e968602e08c324fd270822ae824744cb1c4  thelma_<version>_linux_amd64.tar.gz
func (a *installer) getReleaseArchiveSha256Sum(version string) (string, error) {
	checksumsObject := releaseArchiveSha256SumObject(version)
	checksumsContent, err := a.bucket.Read(releaseArchiveSha256SumObject(version))
	if err != nil {
		return "", fmt.Errorf("error reading checksum object gs://%s/%s: %v", a.bucket.Name(), checksumsObject, err)
	}

	archiveFile := releaseArchiveFilename(version)

	sc := bufio.NewScanner(bytes.NewReader(checksumsContent))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasSuffix(line, archiveFile) {
			return strings.Fields(line)[0], nil
		}
	}
	return "", fmt.Errorf("found no matching checksum for %s in gs://%s/%s", archiveFile, a.bucket.Name(), checksumsObject)
}

// return true if an update is required (i.e., the current installed
// version of Thelma does not match the version of its configured auto-update tag)
func (a *installer) isUpdateNeeded() (bool, error) {
	return a.isUpdateNeededFor(a.config.Tag)
}

// return true if an update is required (i.e., the current installed
// version of Thelma does not match the given version or tag string)
func (a *installer) isUpdateNeededFor(versionOrTag string) (bool, error) {
	_version, err := a.resolveTagToVersion(versionOrTag)
	if err != nil {
		return false, err
	}
	matches, err := a.currentVersionMatches(_version)
	if err != nil {
		return false, err
	}
	return !matches, err
}

// returns the version string that thelma should be updated to, a bool indicating
// whether thelma needs to be updated, and error if one occurred
func (a *installer) currentVersionMatches(version string) (bool, error) {
	currentVersion, err := a.currentVersion()
	if err != nil {
		log.Debug().Err(err).Msgf("error resolving current verison symlink")
		// failed to resolve current version; send false to indicate that we should attempt an upgrade anyway
		// (this happens during bootstrapping when the current symlink does not exist yet)
		return false, nil
	}

	return currentVersion == version, nil
}

// return the Thelma version pointed to by the release directory's "current" symlink
// returns an error if the symlink does not exist, is broken, or another fs error occurs
func (a *installer) currentVersion() (string, error) {
	symlink := a.releasesDir.CurrentSymlink()
	resolved, err := filepath.EvalSymlinks(symlink)
	if err != nil {
		return "", err
	}
	return path.Base(resolved), nil
}

// resolve tag into a version string (eg. "latest" -> "v1.2.3")
func (a *installer) resolveTagToVersion(versionOrTag string) (string, error) {
	tags, err := a.fetchTags()
	if err != nil {
		return "", err
	}
	v, exists := tags[versionOrTag]
	if exists {
		log.Debug().Msgf("Tag %q is an alias for %q", versionOrTag, v)
		return v, nil
	}

	normalizedVersion := versionOrTag
	if !strings.HasPrefix(normalizedVersion, "v") {
		normalizedVersion = fmt.Sprintf("v%s", versionOrTag)
	}
	if !semver.IsValid(normalizedVersion) {
		return "", fmt.Errorf("%q is not a valid Thelma tag or semantic version", versionOrTag)
	}
	return normalizedVersion, nil
}

// fetch list of tags (version aliases) from the thelma releases bucket
func (a *installer) fetchTags() (map[string]string, error) {
	content, err := a.bucket.Read(tagsFile)
	if err != nil {
		return nil, fmt.Errorf("error loading %s from %s: %v", tagsFile, a.bucket.Name(), err)
	}
	tags := make(map[string]string)
	if err = json.Unmarshal(content, &tags); err != nil {
		return nil, fmt.Errorf("error parsing %s from %s: %v", tagsFile, a.bucket.Name(), err)
	}
	return tags, nil
}

func releaseArchiveObject(version string) string {
	return path.Join(releasesObjectPrefix, version, releaseArchiveFilename(version))
}

func releaseArchiveFilename(version string) string {
	return fmt.Sprintf("thelma_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
}

func releaseArchiveSha256SumObject(version string) string {
	return path.Join(releasesObjectPrefix, version, fmt.Sprintf("thelma_%s_SHA256SUMS", version))
}

func currentProcessIsThelmaUpdateCommand() bool {
	args := os.Args
	withoutFlags := slices.Filter(nil, args, func(s string) bool {
		return !strings.HasPrefix(s, "-")
	})
	if len(withoutFlags) < 2 {
		// not sure what we're doing but it ain't "thelma update"
		return false
	}
	return withoutFlags[1] == updateCommandName
}
