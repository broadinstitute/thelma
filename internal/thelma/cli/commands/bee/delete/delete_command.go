package delete

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/views"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Destroy a BEE (Branch Engineering Environment)

Examples:

# Create a hybrid BEE from the swatomation template
thelma bee delete \
  --name=swat-grungy-puma
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

type deleteCommand struct {
	options options
}

func NewBeeDeleteCommand() cli.ThelmaCommand {
	return &deleteCommand{}
}

func (cmd *deleteCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "delete"
	cobraCommand.Short = "Destroy a BEE"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "NAME", "Required. Name of the BEE to delete")
}

func (cmd *deleteCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	// validate --name
	if !ctx.CobraCommand().Flags().Changed(flagNames.name) {
		return fmt.Errorf("--%s is required", flagNames.name)
	}

	return nil
}

func (cmd *deleteCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	_argocd, err := app.Clients().ArgoCD()
	if err != nil {
		return err
	}

	state, err := app.State()
	if err != nil {
		return err
	}

	env, err := state.Environments().Get(cmd.options.name)
	if err != nil {
		return err
	}
	if env == nil {
		return fmt.Errorf("could not delete environment %s: no environment by that name exists", cmd.options.name)
	}

	if err = state.Environments().Delete(env.Name()); err != nil {
		return err
	}

	log.Info().Msgf("Deleted environment %s", cmd.options.name)

	rc.SetOutput(views.ForTerraEnv(env))

	log.Info().Msgf("Syncing %s", bee.GeneratorArgoApp)
	return _argocd.SyncApp(bee.GeneratorArgoApp)
}

func (cmd *deleteCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}
