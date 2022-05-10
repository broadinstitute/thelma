package versions

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/views"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// versions pin
// versions unpin
const helpMessage = `Sync a BEE (Branch Engineering Environment)

Examples:

# Update a bee with 
thelma bee update --name=fiab-automation-grungy-puma --from-versions-properties=images.properties

thelma bee pin --versions-override='{"cromwell":""}'
thelma bee unpin --
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
	cobraCommand.Flags().BoolVar(&cmd.options.ifExists, flagNames.ifExists, false, "Do not return an error if the BEE does not exist")
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
		msg := fmt.Sprintf("Could not delete %s, no BEE by that name exists", cmd.options.name)
		if cmd.options.ifExists {
			log.Warn().Msg(msg)
			return nil
		} else {
			return fmt.Errorf(msg)
		}
	}

	if err = state.Environments().Delete(env.Name()); err != nil {
		return err
	}

	log.Info().Msgf("Deleted environment %s", cmd.options.name)

	rc.SetOutput(views.ForTerraEnv(env))

	log.Info().Msgf("Syncing %s to delete applications", bee.GeneratorArgoApp)
	err = _argocd.SyncApp(bee.GeneratorArgoApp, func(options *argocd.SyncOptions) {
		options.WaitHealthy = false
	})
	if err != nil {
		return err
	}

	// Unfortunately we need to double-sync the generator to destroy the project after the applications are deleted
	log.Info().Msgf("Syncing %s to delete projects", bee.GeneratorArgoApp)
	return _argocd.SyncApp(bee.GeneratorArgoApp, func(options *argocd.SyncOptions) {
		options.WaitHealthy = false
		options.HardRefresh = false
	})
}

func (cmd *deleteCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}
