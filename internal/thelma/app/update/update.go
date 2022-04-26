// Package update is responsible for Thelma's self-updating behavior. See README for more information.
// TODO - this package is a work in progress.
package update

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/app/scratch"
	"time"
)

const configPrefix = "update"
const releaseTagsFile = "tags.json"
const releaseBucket = "thelma-releases"
const LatestTag = "latest"

// Updater handles Thelma updates
type Updater interface {
	// AutoUpdate performs a Thelma update, if auto-update is enabled in Thelma's configuration.
	// Returns true if an update was actually done, false otherwise.
	AutoUpdate() (bool, error)
	// InstallVersion will install the given version of Thelma, even if it is a downgrade.
	// If the target version matches the currently installed version of Thelma, no update is performed.
	InstallVersion(version string) (bool, error)
}

type updateConfig struct {
	Version         string        `default:"latest"`
	AutoUpdate      bool          `default:"false"`
	RefreshInterval time.Duration `default:"15m"`
}

func Load(thelmaConfig config.Config, thelmaRoot root.Root, thelmaScratch scratch.Scratch) (Updater, error) {
	// TODO
	return nil, nil
}

// updater implements the Updater interface
type updater struct {
	releasesDir string
	scratchDir  string
	tagsFile    string
}
