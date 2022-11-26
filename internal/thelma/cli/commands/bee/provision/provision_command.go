package provision

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/pinflags"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/seedflags"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/views"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

const helpMessage = `Provision the resources for a newly created BEE (Branch Engineering Environment)

Examples:

thelma bee provision --name=bee-swat-ecstatic-spider
`

var flagNames = struct {
	name                      string
	generatorOnly             string
	waitHealthy               string
	waitHealthyTimeoutSeconds string
	seed                      string
	notify                    string
	exportLogsOnFailure       string
}{
	name:                      "name",
	generatorOnly:             "generator-only",
	waitHealthy:               "wait-healthy",
	waitHealthyTimeoutSeconds: "wait-healthy-timeout-seconds",
	seed:                      "seed",
	notify:                    "notify",
	exportLogsOnFailure:       "export-logs-on-failure",
}

type provisionCommand struct {
	options   bee.ProvisionOptions
	pinFlags  pinflags.PinFlags
	seedFlags seedflags.SeedFlags
}

func NewBeeProvisionCommand() cli.ThelmaCommand {
	return &provisionCommand{
		pinFlags: pinflags.NewPinFlags(),
		seedFlags: seedflags.NewSeedFlags(func(options *seedflags.Options) {
			options.Prefix = "seed-"
			options.NoShortHand = true
			options.Hidden = false
		}),
	}
}

func (cmd *provisionCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "provision"
	cobraCommand.Short = "Provision and sync the resources for a newly-created BEE"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.Name, flagNames.name, "n", "NAME", "Name of the newly-created BEE to provision")
	cobraCommand.Flags().BoolVar(&cmd.options.SyncGeneratorOnly, flagNames.generatorOnly, false, "Sync the BEE generator but not the BEE's Argo apps")
	cobraCommand.Flags().BoolVar(&cmd.options.WaitHealthy, flagNames.waitHealthy, true, "Wait for BEE's Argo apps to become healthy after syncing")
	cobraCommand.Flags().IntVar(&cmd.options.WaitHealthTimeoutSeconds, flagNames.waitHealthyTimeoutSeconds, 1200, "How long to wait for BEE's Argo apps to become healthy after syncing")
	cobraCommand.Flags().BoolVar(&cmd.options.Seed, flagNames.seed, true, `Seed BEE after creation (run "thelma bee seed -h" for more info)`)
	cobraCommand.Flags().BoolVar(&cmd.options.ExportLogsOnFailure, flagNames.exportLogsOnFailure, true, `Export container logs to GCS if BEE creation fails)`)
	cobraCommand.Flags().BoolVar(&cmd.options.Notify, flagNames.notify, true, "Attempt to notify the owner via Slack upon success")

	cmd.pinFlags.AddFlags(cobraCommand)
	cmd.seedFlags.AddFlags(cobraCommand)
}

func (cmd *provisionCommand) PreRun(app app.ThelmaApp, ctx cli.RunContext) error {
	if !ctx.CobraCommand().Flags().Changed(flagNames.name) {
		return fmt.Errorf("no environment name specified; --%s is required", flagNames.name)
	}
	if strings.TrimSpace(cmd.options.Name) == "" {
		log.Warn().Msg("Is Thelma running in CI? Check that you're setting the name of your environment when running your job")
		return fmt.Errorf("no environment name specified; --%s was passed but no name was given", flagNames.name)
	}
	// validate/load pin and seed options
	pinOptions, err := cmd.pinFlags.GetPinOptions(app, ctx)
	if err != nil {
		return err
	}
	cmd.options.PinOptions = pinOptions

	seedOptions, err := cmd.seedFlags.GetOptions(ctx.CobraCommand())
	if err != nil {
		return err
	}
	cmd.options.SeedOptions = seedOptions

	return nil
}

func (cmd *provisionCommand) Run(thelmaApp app.ThelmaApp, ctx cli.RunContext) error {
	bees, err := builders.NewBees(thelmaApp)
	if err != nil {
		return err
	}
	_bee, err := bees.ProvisionWith(cmd.options.Name, cmd.options)
	if _bee != nil {
		ctx.SetOutput(views.DescribeBee(_bee))
	}
	return err
}

func (cmd *provisionCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}
