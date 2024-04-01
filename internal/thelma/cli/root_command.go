package cli

import (
	"bytes"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/app/name"
	"github.com/broadinstitute/thelma/internal/thelma/cli/environmentflags"
	"github.com/broadinstitute/thelma/internal/thelma/cli/printing"
	"github.com/broadinstitute/thelma/internal/thelma/cli/printing/format"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

// thelmaCommandName command name for the thelma tool
const thelmaCommandName = name.Name

// globalUsage common usage string printed for all subcommands
const globalUsage = `CLI tools for interacting with Terra's Helm charts`

// rootCommand is the root command for Thelma. It configures global flags and their related features.
type rootCommand struct {
	printer printing.Printer
}

func newRootCommand() ThelmaCommand {
	return &rootCommand{
		printer: printing.NewPrinter(),
	}
}

func (r *rootCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	// Configure Cobra with thelma defaults
	cobraCommand.Use = thelmaCommandName
	cobraCommand.Short = "CLI tools for Terra Helm"
	cobraCommand.Long = globalUsage
	cobraCommand.SilenceUsage = true  // Only print out usage error when user supplies -h/--help
	cobraCommand.SilenceErrors = true // Don't print errors, we do it ourselves using a logging library

	// Add output formatting flags to Cobra command (e.g. "--output-format=yaml")
	r.printer.AddFlags(cobraCommand.PersistentFlags())

	// Add flag for control if other flags should be read from the environment (e.g. "--flags-from-environment-prefix=PARAM_")
	// See execution.setFlagsFromEnvironment for where this option gets picked up (and why there, rather than in PreRun here)
	environmentflags.AddFlag(cobraCommand.PersistentFlags())
}

func (r *rootCommand) PreRun(_ app.ThelmaApp, _ RunContext) error {
	log.Debug().Strs("argv", os.Args).Msgf("Starting new thelma run")
	// check that output format flags were used correctly
	if err := r.printer.VerifyFlags(); err != nil {
		return err
	}
	return nil
}

func (r *rootCommand) Run(_ app.ThelmaApp, _ RunContext) error {
	panic("Run() is only executed for leaf commands")
}

func (r *rootCommand) PostRun(_ app.ThelmaApp, ctx RunContext) error {
	if ctx.HasOutput() {
		// write YAML-formatted output to debug log
		buf := &bytes.Buffer{}
		if err := format.Yaml.Format(ctx.Output(), buf); err != nil {
			log.Warn().Err(err).Msgf("error writing output to debug log")
		}
		log.Debug().Str("output", buf.String()).Msgf("Writing thelma output")

		// write user-formatted output to stdout (or file, as configured)
		if err := r.printer.PrintOutput(ctx.Output(), ctx.CobraCommand().OutOrStdout()); err != nil {
			return err
		}
	}

	return nil
}
