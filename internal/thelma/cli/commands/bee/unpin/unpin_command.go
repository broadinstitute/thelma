package unpin

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/builders"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Remove version overrides from a BEE (Branch Engineering Environment)

Examples:

# Remove all version overrides from the swat-grungy-puma BEE
thelma bee unpin --name=swat-grungy-puma
`

type options struct {
	name string
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	name string
}{
	name: "name",
}

type unpinCommand struct {
	options options
}

func NewBeeUnpinCommand() cli.ThelmaCommand {
	return &unpinCommand{}
}

func (cmd *unpinCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "unpin"
	cobraCommand.Short = "Remove version overrides from a BEE"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "", "Required. Name of the BEE to delete")
}

func (cmd *unpinCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	// validate --name
	if !ctx.CobraCommand().Flags().Changed(flagNames.name) {
		return fmt.Errorf("--%s is required", flagNames.name)
	}

	return nil
}

func (cmd *unpinCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	state, err := app.State()
	if err != nil {
		return err
	}
	removed, err := state.Environments().UnpinVersions(cmd.options.name)
	if err != nil {
		return err
	}
	log.Info().Msgf("Removed all version overrides for %s", cmd.options.name)

	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}
	if err = bees.SyncGeneratorForName(cmd.options.name); err != nil {
		return err
	}

	log.Info().Msgf("The following overrides were removed:")
	rc.SetOutput(removed)
	return nil

}

func (cmd *unpinCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}
