package toolbox

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/helm"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"path"
	"path/filepath"
)

// toolsDirName name of the root directory that includes Thelma's bundled tools
const toolsDirName = "tools"

// executableDirName name of the directory that includes Thelma's bundled executables (as apposed to config files, etc)
const executableDirName = "bin"

// verifyTool presence of this tool in the tools directory will be checked to ensure that Thelma identified
// the correct path
const verifyTool = helm.ProgName

// FindToolsDir identifies the path to the local directory containing Thelma's bundled tools, and returning an error
// if it can't be found.
// It does this by trying to expand "../../tools" relative to $0 (the currently-running Thelma executable)
// Note that this location is correct:
//   - for installed Thelma instances running out of ~/.thelma/releases/current
//     (thelma is located in ~/.thelma/releases/current/bin, tools are in ~/.thelma/releases/current/tools/bin)
//   - for locally-build Thelma compiled with `make build` (thelma is compiled to ./output/bin,
//     third-party tools are downloaded ./output/tools/bin)
//   - for unpacked Thelma releases (release archives have thelma at ./bin/ and tools at ./tools/bin)
func FindToolsDir() (string, error) {
	exe, err := utils.PathToRunningThelmaExecutable()
	if err != nil {
		return "", fmt.Errorf("error resolving path to Thelma's bundled tools: %v", err)
	}

	toolsDir, err := findToolsDir(exe)
	if err != nil {
		return "", fmt.Errorf("error resolving path to Thelma's bundled tools: %v", err)
	}

	return toolsDir, nil
}

func findToolsDir(thelmaPath string) (string, error) {
	thelmaPath, err := filepath.EvalSymlinks(thelmaPath)
	if err != nil {
		return "", err
	}

	toolsdir := filepath.Clean(path.Join(thelmaPath, "..", "..", toolsDirName))

	if err = validateToolsDir(toolsdir); err != nil {
		return "", err
	}
	return toolsdir, nil
}

func validateToolsDir(toolsdir string) error {
	// make sure a tool executable exists in the tools dir.
	toolexe := path.Join(toolsdir, executableDirName, verifyTool)
	exists, err := utils.FileExists(toolexe)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("tools dir not found; %s does not exist", toolexe)
	}
	return nil
}
