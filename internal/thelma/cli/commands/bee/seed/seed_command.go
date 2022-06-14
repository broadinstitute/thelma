package seed

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

type options struct {
	name string
}

var flagNames = struct {
	name string
}{
	name: "name",
}

type seedCommand struct {
	options options
}

func NewBeeSeedCommand() cli.ThelmaCommand {
	return &seedCommand{}
}

func (cmd *seedCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "seed [options]"
	cobraCommand.Short = "Populate the BEE with standard app-to-app permissions and test data"
	cobraCommand.Long = `Populate the BEE with standard app-to-app permissions and test data.

This command aims for parity with the old firecloud-develop FiaB population scripts.

TODO (Jack): write what actually happens here as I figure that out.`

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "", "Required. Name of the BEE to seed")
}

func (cmd *seedCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	flags := ctx.CobraCommand().Flags()

	// validate --name
	if !flags.Changed(flagNames.name) {
		return fmt.Errorf("--%s is required", flagNames.name)
	}

	return nil
}

func (cmd *seedCommand) Run(app app.ThelmaApp, _ cli.RunContext) error {
	state, err := app.State()
	if err != nil {
		return err
	}
	env, err := state.Environments().Get(cmd.options.name)
	if err != nil {
		return err
	}

	if env == nil {
		return fmt.Errorf("BEE '%s' not found, it may be a vanilla FiaB or might not exist at all", cmd.options.name)
	}

	return nil
}

func (cmd *seedCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}
