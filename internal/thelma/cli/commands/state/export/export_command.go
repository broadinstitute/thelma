package export

import (
	"fmt"
	"strings"

	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/sherlock"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Exports thelmas internal state to a destination`
const defaultFormat = "sherlock"
const prodSherlockHostName = "sherlock.dsp-devops.broadinstitute.org"

var ErrExportDestinationForbidden = fmt.Errorf("state export to production sherlock: %s is not allowed", prodSherlockHostName)

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
	// check to make sure destination is not prod sherlock, this should not be allowed
	if strings.Contains(cmd.options.destinationURL, prodSherlockHostName) {
		log.Warn().Msgf("exporting to destination: %s is forbidden", prodSherlockHostName)
		return ErrExportDestinationForbidden
	}

	log.Info().Msgf("exporting state to: %s using format: %s", cmd.options.destinationURL, cmd.options.format)
	sherlockClient, err := app.Clients().Sherlock()
	if err != nil {
		return fmt.Errorf("error retrieving sherlock client: %v", err)
	}
	state, err := app.State()
	if err != nil {
		return fmt.Errorf("error retrieving Thelma state: %v", err)
	}

	stateExporter := sherlock.NewSherlockStateWriter(state, sherlockClient)

	if err := stateExporter.WriteClusters(); err != nil {
		return fmt.Errorf("erorr exporting clusters: %v", err)
	}

	if err := stateExporter.WriteEnvironments(); err != nil {
		return fmt.Errorf("erorr exporting environments: %v", err)
	}

	return nil
}

func (cmd *exportCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}
