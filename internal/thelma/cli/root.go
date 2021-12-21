package cli

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io"
	"os"
)

// commandName name of the thelma tool
const commandName = "thelma"

// globalUsage common usage string printed for all subcommands
const globalUsage = `CLI tools for interacting with Terra's Helm charts

Environment variables:
| Name                               | Description                                                                       |
|------------------------------------|-----------------------------------------------------------------------------------|
| $THELMA_HOME                       | Required. Path to terra-helmfile clone.                                           |
| $THELMA_LOGLEVEL                   | Logging verbosity. One of error, warn, info (default), debug, or trace            |
| $THELMA_TMPDIR                     | Path where Thelma should generate temporary files. Defaults to OS tmp dir.        |
`

func init() {
	// Initialize logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// ThelmaCLI represents a complete command-line interface for Thelma, including subcommands
type ThelmaCLI struct {
	builder     builder.ThelmaBuilder
	rootCommand *cobra.Command
	chartsCLI   *chartsCLI
	renderCLI   *renderCLI
	versionCLI  *versionCLI
}

// Execute is the main method/entrypoint for Thelma
func Execute() {
	cli := newThelmaCLI()

	if err := cli.execute(); err != nil {
		log.Error().Msgf("%v", err)
		os.Exit(1)
	}
}

// execute executes a command
func (cli *ThelmaCLI) execute() error {
	return cli.rootCommand.Execute()
}

// setArgs (for use in tests only) sets command-line arguments on the cobra command
func (cli *ThelmaCLI) setArgs(args []string) {
	cli.rootCommand.SetArgs(args)
}

// setHome (for use in tests only) makes it possible to set THELMA_HOME to
// a custom path for testing
func (cli *ThelmaCLI) setHome(thelmaHome string) {
	cli.builder.SetConfigOverride(config.Keys.Home, thelmaHome)
}

// setLogLevel (for use in tests only) makes it possible to set THELMA_LOGLEVEL to
// a custom value for testing
func (cli *ThelmaCLI) setLogLevel(level string) {
	cli.builder.SetConfigOverride(config.Keys.LogLevel, level)
}

// setShellRunner (for use in tests only) configures this CLI instance to use the given shell runner
func (cli *ThelmaCLI) setShellRunner(runner shell.Runner) {
	cli.builder.SetShellRunner(runner)
}

// setStdout (for use in tests only) configures this CLI instance to use the given shell runner
func (cli *ThelmaCLI) setStdout(stdout io.Writer) {
	cli.rootCommand.SetOut(stdout)
}

// newThelmaCLI constructs a new Thelma CLI
func newThelmaCLI() *ThelmaCLI {
	_builder := builder.NewBuilder()

	rootCommand := &cobra.Command{
		Use:           commandName,
		Short:         "CLI tools for Terra Helm",
		Long:          globalUsage,
		SilenceUsage:  true, // Only print out usage error when user supplies -h/--help
		SilenceErrors: true, // Don't print errors, we do it ourselves using a logging library
	}

	cli := ThelmaCLI{
		builder:     _builder,
		rootCommand: rootCommand,
		chartsCLI:   newChartsCLI(_builder),
		renderCLI:   newRenderCLI(_builder),
		versionCLI:  newVersionCLI(_builder),
	}

	// Close ThelmaApp if a subcommand initialized it
	rootCommand.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		return _builder.Close()
	}

	// Add subcommands
	rootCommand.AddCommand(
		cli.chartsCLI.cobraCommand,
		cli.renderCLI.cobraCommand,
		cli.versionCLI.cobraCommand,
	)

	return &cli
}
