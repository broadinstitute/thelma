package export

import (
	"fmt"
	"strings"

	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	sherlock_client "github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
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
	options        *options
	sherlockClient *sherlock_client.Client
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
	// construct a sherlock client to export state to, this is different than the app level
	// sherlock client to support use cases such as exporting state from prod to a local sherlock for debugging

	iapToken, err := app.Clients().IAPToken()
	if err != nil {
		return fmt.Errorf("error retrieving iap token for exporter client: %v", err)
	}
	client, err := sherlock_client.NewWithHostnameOverride(cmd.options.destinationURL, iapToken)
	if err != nil {
		return fmt.Errorf("error building exporter sherlock client")
	}
	cmd.sherlockClient = client
	return nil
}

func (cmd *exportCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	// check to make sure destination is not prod sherlock, this should not be allowed
	if strings.Contains(cmd.options.destinationURL, prodSherlockHostName) {
		log.Warn().Msgf("exporting to destination: %s is forbidden", prodSherlockHostName)
		return ErrExportDestinationForbidden
	}

	log.Info().Msgf("exporting state to: %s using format: %s", cmd.options.destinationURL, cmd.options.format)
	state, err := app.State()
	if err != nil {
		return fmt.Errorf("error retrieving Thelma state: %v", err)
	}

	stateExporter := sherlock.NewSherlockStateWriter(state, cmd.sherlockClient)

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
