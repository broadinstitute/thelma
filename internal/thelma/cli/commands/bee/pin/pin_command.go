package pin

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/pinflags"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/argocd"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

const helpMessage = `Override the version of a service that is deployed to a BEE.

Examples:

# Pin leonardo application image to tag v100
thelma bee pin -n swat-grungy-puma sam --app-version=v100

# Pin sam helm chart to version 0.10.3
thelma bee pin -n swat-grungy-puma sam --chart-version=0.10.3

# Pin sam to the terra-helmfile PR branch my-pr-1
thelma bee pin -n swat-grungy-puma sam --terra-helmfile-ref=my-pr-1

# Pin all services in a BEE to the terra-helmfile PR branch my-pr-1
thelma bee pin -n swat-grungy-puma ALL --terra-helmfile-ref=my-pr-1

# Pin services in a BEE to specific versions described in the given file, with a format like:
#   {
#      "sam": {
#        "appVersion": "my-image-tag",
#        "terraHelmfileRef": "my-terra-helmfile-branch"
#      },
#      "leonardo": {
#        ...
#      },
#      ...
#   }
thelma bee pin -n swat-grungy-puma sam --versions-file=/tmp/version.json --versions-format=json
`

type options struct {
	name        string
	sync        bool
	waitHealthy bool
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	name        string
	sync        string
	waitHealthy string
}{
	name:        "name",
	sync:        "sync",
	waitHealthy: "wait-healthy",
}

type pinCommand struct {
	options    options
	pinOptions pinflags.PinFlags
}

func NewBeePinCommand() cli.ThelmaCommand {
	return &pinCommand{
		pinOptions: pinflags.NewPinFlags(),
	}
}

func (cmd *pinCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "pin [SERVICE] [options]"
	cobraCommand.Short = "Pin a BEE to specific version"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "", "Required. Name of the BEE to pin")

	cmd.pinOptions.AddFlags(cobraCommand)

	cobraCommand.Flags().BoolVar(&cmd.options.sync, flagNames.sync, true, "Sync all services in BEE after updating versions")
	cobraCommand.Flags().BoolVar(&cmd.options.waitHealthy, flagNames.waitHealthy, true, "Wait for BEE's Argo apps to become healthy after syncing")
}

func (cmd *pinCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	flags := ctx.CobraCommand().Flags()

	// validate --name
	if !flags.Changed(flagNames.name) {
		return errors.Errorf("no environment name specified; --%s is required", flagNames.name)
	}
	if strings.TrimSpace(cmd.options.name) == "" {
		log.Warn().Msg("Is Thelma running in CI? Check that you're setting the name of your environment when running your job")
		return errors.Errorf("no environment name specified; --%s was passed but no name was given", flagNames.name)
	}

	return nil
}

func (cmd *pinCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
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

	pinOptions, err := cmd.pinOptions.GetPinOptions(app, ctx)
	if err != nil {
		return err
	}

	env, err = bees.PinVersions(env, pinOptions)
	if err != nil {
		return err
	}
	ctx.SetOutput(pinOptions)

	if err = bees.RefreshBeeGenerator(); err != nil {
		return err
	}
	if err = bees.SyncEnvironmentGenerator(env); err != nil {
		return err
	}
	if !cmd.options.sync {
		return nil
	}
	_, err = bees.SyncArgoAppsIn(env, func(options *argocd.SyncOptions) {
		options.WaitHealthy = cmd.options.waitHealthy
	})
	return err
}

func (cmd *pinCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}
