package helmfile

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"strings"
)

// ProgName is the name of the `helmfile` binary
const ProgName = "helmfile"

// Cmd encapsulates low-level parameters for a `helmfile` command
type Cmd struct {
	dir             string
	skipDeps        bool
	skipTests       bool
	logLevel        string
	envVars         []string
	stateValuesFile string
	valuesFiles     []string
	outputDir       string
	stdout          bool
}

// newCmd returns a new Cmd object with all fields initialized
func newCmd() *Cmd {
	return &Cmd{}
}

func (cmd *Cmd) toShellCommand() shell.Command {
	// Convert helmfile parameters into cli arguments
	var cliArgs []string

	if cmd.logLevel != "" {
		cliArgs = append(cliArgs, fmt.Sprintf("--log-level=%s", cmd.logLevel))
	}

	if cmd.stateValuesFile != "" {
		cliArgs = append(cliArgs, fmt.Sprintf("--state-values-file=%s", cmd.stateValuesFile))
	}

	// Append Helmfile command we're running (template)
	cliArgs = append(cliArgs, "template")

	// Append arguments specific to template subcommand
	if cmd.skipDeps {
		// Skip dependencies unless we're rendering a local chart, to save time
		cliArgs = append(cliArgs, "--skip-deps")
	}
	if cmd.skipTests {
		cliArgs = append(cliArgs, "--skip-tests")
	}
	if len(cmd.valuesFiles) > 0 {
		cliArgs = append(cliArgs, fmt.Sprintf("--values=%s", strings.Join(cmd.valuesFiles, ",")))
	}

	if !cmd.stdout {
		outputDirFlag := fmt.Sprintf("--output-dir=%s", cmd.outputDir)
		cliArgs = append(cliArgs, outputDirFlag)
	}

	shellCmd := shell.Command{
		Prog: ProgName,
		Args: cliArgs,
		Dir:  cmd.dir,
		Env:  cmd.envVars,
	}

	return shellCmd
}

func (cmd *Cmd) setStateValuesFile(file string) {
	cmd.stateValuesFile = file
}

func (cmd *Cmd) setDir(dir string) {
	cmd.dir = dir
}

func (cmd *Cmd) setLogLevel(logLevel string) {
	cmd.logLevel = logLevel
}

func (cmd *Cmd) setSkipDeps(skipDeps bool) {
	cmd.skipDeps = skipDeps
}

func (cmd *Cmd) setSkipTests(skipTests bool) {
	cmd.skipTests = skipTests
}

func (cmd *Cmd) setOutputDir(outputDir string) {
	cmd.outputDir = outputDir
}

func (cmd *Cmd) setStdout(stdout bool) {
	cmd.stdout = stdout
}

func (cmd *Cmd) addValuesFiles(valuesFiles ...string) {
	cmd.valuesFiles = append(cmd.valuesFiles, valuesFiles...)
}
