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
	name     string
	ifExists bool
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	name     string
	ifExists string
}{
	name:     "name",
	ifExists: "if-exists",
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
	cobraCommand.Flags().BoolVar(&cmd.options.ifExists, flagNames.ifExists, false, "Do not return an error if the BEE does not exist")
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

	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}

	env, err := state.Environments().Get(cmd.options.name)
	if err != nil {
		return err
	}
	if env == nil {
		if cmd.options.ifExists {
			log.Warn().Msgf("Could not unpin %s, no BEE by that name exists", cmd.options.name)
			return nil
		}
		return fmt.Errorf("--%s: unknown bee %q", flagNames.name, cmd.options.name)
	}

	removed, err := state.Environments().UnpinVersions(cmd.options.name)
	if err != nil {
		return err
	}
	log.Info().Msgf("Removed all version overrides for %s", cmd.options.name)

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
