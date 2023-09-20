// Package autoupdate implements Thelma's self-install and self-update features
package autoupdate

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/bootstrap"
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/installer"
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/releasebucket"
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/releases"
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/spawn"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/app/scratch"
	"github.com/broadinstitute/thelma/internal/thelma/clients/api"
	"github.com/broadinstitute/thelma/internal/thelma/utils/lazy"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"
	"os"
	"strings"
)

const configKey = "autoupdate"
const updateCommandName = "update"

type updateConfig struct {
	Enabled      bool   `default:"true"`            // if false, do not perform automatic updates
	Tag          string `default:"latest"`          // which Thelma build tag to follow for auto-updates
	Bucket       string `default:"thelma-releases"` // name of GCS bucket that contains Thelma releases
	KeepReleases int    `default:"10"`              // number of old Thelma releases to keep around in the ~/.releases dir
}

type AutoUpdate interface {
	// Update performs a foreground update of Thelma to the current version of its configured tag
	Update() error
	// UpdateTo performs a foreground update of Thelma to a specific version
	UpdateTo(version string) error
	// StartBackgroundUpdateIfEnabled start a new process that will update Thelma in the background
	StartBackgroundUpdateIfEnabled() error
	// Bootstrap set up a new Thelma installation, including generating shell scripts
	Bootstrap() error
}

func New(thelmaConfig config.Config, bucketFactory api.BucketFactory, root root.Root, runner shell.Runner, scratch scratch.Scratch) (AutoUpdate, error) {
	var cfg updateConfig

	if err := thelmaConfig.Unmarshal(configKey, &cfg); err != nil {
		return nil, err
	}

	bootstrapper := bootstrap.New(root, thelmaConfig, runner)

	_installer := lazy.NewLazyE[installer.Installer](func() (installer.Installer, error) {
		gcsBucket, err := bucketFactory.Bucket(cfg.Bucket)
		if err != nil {
			return nil, err
		}
		releasesBucket := releasebucket.New(gcsBucket, runner, scratch)

		releasesDir := releases.NewDir(root.ReleasesDir(), scratch)

		return installer.New(releasesDir, releasesBucket, func(options *installer.Options) {
			options.KeepReleases = cfg.KeepReleases
		}), nil
	})

	_spawn := spawn.New(root, func(options *spawn.Options) {
		// write spawned update process logs to ~/.thelma/logs/update.out and ~/.thelma/logs/update.err
		options.LogFileName = updateCommandName
	})

	return &autoupdate{
		config:       cfg,
		installer:    _installer,
		bootstrapper: bootstrapper,
		spawn:        _spawn,
	}, nil
}

type autoupdate struct {
	config       updateConfig
	installer    lazy.LazyE[installer.Installer]
	bootstrapper bootstrap.Bootstrapper
	spawn        spawn.Spawn
}

func (a *autoupdate) Update() error {
	return a.updateTo(a.config.Tag)
}

func (a *autoupdate) UpdateTo(versionOrTag string) error {
	if a.config.Enabled {
		return errors.Errorf("auto-update is enabled; please disable by " +
			"setting THELMA_AUTOUPDATE_ENABLED=false or adding autoupdate.enabled=false " +
			"to ~/.thelma/config.yaml and re-run (otherwise this change will be overwritten!)")
	}

	return a.updateTo(versionOrTag)
}

func (a *autoupdate) StartBackgroundUpdateIfEnabled() error {
	if !a.config.Enabled {
		return nil
	}

	if a.spawn.CurrentProcessIsSpawn() {
		// never start a background process from a background process (infinite loops are bad)
		return nil
	}

	if currentProcessIsThelmaUpdateCommand() {
		// don't trigger a background update from a manually-run `thelma update` command
		return nil
	}

	_installer, err := a.installer.Get()
	if err != nil {
		return errors.Errorf("error initializing installer: %v", err)
	}
	resolved, err := _installer.ResolveVersions(a.config.Tag)
	if err != nil {
		return errors.Errorf("error preparing background Thelma updates: %v", err)
	}
	if !resolved.UpdateNeeded() {
		return nil
	}

	// launch "thelma update" in the background
	return a.spawn.Spawn(updateCommandName)
}

func (a *autoupdate) Bootstrap() error {
	return a.bootstrapper.Bootstrap()
}

// updateTo will create a new installer instance via the lazy initializer
// and then use it to perform an update
func (a *autoupdate) updateTo(versionOrTag string) error {
	_installer, err := a.installer.Get()
	if err != nil {
		return errors.Errorf("error initializing installer: %v", err)
	}
	return _installer.UpdateThelma(versionOrTag)
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
