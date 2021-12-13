package cli

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/loader"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
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
	loader      loader.ThelmaLoader
	rootCommand *cobra.Command
	chartsCLI   *chartsCLI
	renderCLI   *renderCLI
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
	cli.loader.SetConfigOverride(config.Keys.Home, thelmaHome)
}

// setLogLevel (for use in tests only) makes it possible to set THELMA_LOGLEVEL to
// a custom value for testing
func (cli *ThelmaCLI) setLogLevel(level string) {
	cli.loader.SetConfigOverride(config.Keys.LogLevel, level)
}

// setShellRunner (for use in tests only) configures this CLI instance to use the given shell runner
func (cli *ThelmaCLI) setShellRunner(runner shell.Runner) {
	cli.loader.SetShellRunner(runner)
}

// newThelmaCLI constructs a new Thelma CLI
func newThelmaCLI() *ThelmaCLI {
	_loader := loader.NewLoader()

	rootCommand := &cobra.Command{
		Use:           commandName,
		Short:         "CLI tools for Terra Helm",
		Long:          globalUsage,
		SilenceUsage:  true, // Only print out usage error when user supplies -h/--help
		SilenceErrors: true, // Don't print errors, we do it ourselves using a logging library
	}

	cli := ThelmaCLI{
		loader:      _loader,
		rootCommand: rootCommand,
		chartsCLI:   newChartsCLI(_loader),
		renderCLI:   newRenderCLI(_loader),
	}

	// Close ThelmaApp if a subcommand initialized it
	rootCommand.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		return _loader.Close()
	}

	// Add subcommands
	rootCommand.AddCommand(
		cli.chartsCLI.cobraCommand,
		cli.renderCLI.cobraCommand,
	)

	return &cli
}
