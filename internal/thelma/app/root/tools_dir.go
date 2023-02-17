package root

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/helm"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"path/filepath"
)

const toolsdirname = "tools"
const bindirname = "bin"

// PathToRunningThelmaExecutable returns the path to the currently-running
// Thelma binary executable.
// Note that this could be _outside_ Thelma's configured root directory
// (i.e., not ~/.thelma/releases/current/bin).
// For example:
//   - During initial installation, Thelma is run out of Thelma release archive
//     that is unpacked into a temp directory.
//   - In CI pipelines, Thelma is run out of a well-known path on it's Docker image
//     /thelma/bin/thelma
//   - When Thelma is built locally during development, it is run out of the build
//     output directory, ./output/bin/thelma
func PathToRunningThelmaExecutable() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("error finding path to currently running executable: %v", err)
	}

	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return "", fmt.Errorf("error finding path to currently running executable: %v", err)
	}

	return executable, nil
}

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

	log.Warn().Err(err).Msgf("error identifying path to thelma executable; will search %s for bundled tools", releasesDir.ReleasesRoot())

	toolsdir = path.Join(releasesDir.ForCurrentVersion(), toolsdirname)
	if err = validateToolsDir(toolsdir); err != nil {
		return "", err
	}
	return toolsdir, nil
}

// try to expand ../../tools relative to $0
func findToolsDirRelativeToThelmaExecutable() (string, error) {
	exepath, err := PathToRunningThelmaExecutable()
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
