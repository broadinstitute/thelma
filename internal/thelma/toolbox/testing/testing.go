package testing

import (
	"github.com/broadinstitute/thelma/internal/thelma/toolbox"
	"github.com/pkg/errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// testsToolsDirEnvVar name of the environment variable that is set by the Makefile to point to the local tools directory
const testsToolsBinDirEnvVar = "TOOLS_BIN_DIR"

// toolsSubDir is the subdirectory of the project root that contains the tools
const toolsSubdir = "output/tools/bin"

// makefile name of the Makefile that is used as a signal to identify the project root
const makefile = "Makefile"

// NewToolFinderForTests returns a toolbox.ToolFinder for use in unit tests, pointing at the
// output/tools/bin directory that is created by `make build`
//
// If this fails when run in Goland, make sure that you have run `make build` at least once.
//
// You can manually specify the tools bin dir by setting the TOOLS_BIN_DIR environment variable
// to the desired path before running your tests
func NewToolFinderForTests() (toolbox.ToolFinder, error) {
	envVar := os.Getenv(testsToolsBinDirEnvVar)
	if envVar != "" {
		return toolbox.NewToolFinderWithDir(envVar)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, errors.Errorf("error getting cwd: %v", err)
	}

	filepath.SplitList(workingDir)
	dir := path.Clean(workingDir)

	max := strings.Count(dir, string(os.PathSeparator))
	for i := 0; i < max; i++ { // prevent infinite loop by only iterating number of components in path

		if _, err = os.Stat(path.Join(dir, makefile)); err == nil {
			return toolbox.NewToolFinderWithDir(path.Join(dir, toolsSubdir))
		} else if !os.IsNotExist(err) {
			return nil, errors.Errorf("error checking if %s exists: %v", makefile, err)
		}

		parent := path.Dir(dir)
		if parent == "/" || parent == "" {
			break
		}
		dir = parent
	}

	return nil, errors.Errorf("could not find project root")
}
