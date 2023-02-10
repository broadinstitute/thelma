package root

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/tools/helm"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"path/filepath"
)

const toolsdirname = "tools"
const bindirname = "bin"

// ToolsDir path on disk where Thelma's bundled tools, such as `kubectl`, `helm`, etc live
type ToolsDir interface {
	// Bin subdir within tools dir that includes tool binaries
	Bin() string
}

type toolsDir struct {
	dir string
}

func (t toolsDir) Bin() string {
	return path.Join(t.dir, bindirname)
}

func findToolsDir(releasesDir ReleasesDir) (string, error) {
	toolsdir, err := findToolsDirRelativeToThelmaExecutable()
	if err == nil {
		return toolsdir, nil
	}

	log.Warn().Err(err).Msgf("error identifying path to thelma executable; will search %s for bundled tools", releasesDir.Root())

	toolsdir = path.Join(releasesDir.ForCurrentVersion(), toolsdirname)
	if err = validateToolsDir(toolsdir); err != nil {
		return "", err
	}
	return toolsdir, nil
}

// try to expand ../../tools relative to $0
// we make it possible to pass in a fake path for testing
func findToolsDirRelativeToThelmaExecutable() (string, error) {
	// use os lib to find executable that launched the process.
	// Note that in the real world, this should always be the THelma binary, but in
	// tests it will be a `go test` command
	exepath, err := os.Executable()
	if err != nil {
		return "", err
	}

	exepath, err = filepath.EvalSymlinks(exepath)
	if err != nil {
		return "", err
	}

	toolsdir := filepath.Clean(path.Join(exepath, "..", "..", toolsdirname))

	if err = validateToolsDir(toolsdir); err != nil {
		return "", err
	}
	return toolsdir, nil
}

func validateToolsDir(toolsdir string) error {
	// make sure a helm executable exists in the tools dir.
	helmexe := path.Join(toolsdir, bindirname, helm.ProgName)
	exists, err := utils.FileExists(helmexe)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("tools dir not found; %s does not exist", helmexe)
	}
	return nil
}
