package export

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Exports thelmas internal state to a destination`
const defaultFormat = "sherlock"

type options struct {
	destinationURL string
	format         string
}

var flagNames = struct {
	destinationURL string
	format         string
}{
	destinationURL: "destination",
	format:         "format",
}

type exportCommand struct {
	options *options
}

func NewStateExportCommand() cli.ThelmaCommand {
	return &exportCommand{
		options: &options{},
	}
}

func (cmd *exportCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "export [options]"
	cobraCommand.Short = "exports thelma's internal state"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVar(&cmd.options.destinationURL, flagNames.destinationURL, "http://localhost:8080", "destination to export state to")
	cobraCommand.Flags().StringVar(&cmd.options.format, flagNames.format, "sherlock", "format in which to output state, currently only sherlock is supported")
}

func (cmd *exportCommand) PreRun(app app.ThelmaApp, ctx cli.RunContext) error {
	return nil
}

func (cmd *exportCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	log.Info().Msg("Hello from the new state command")
	return nil
}

func (cmd *exportCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}
