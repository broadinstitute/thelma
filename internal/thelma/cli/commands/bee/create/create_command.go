package create

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/pinflags"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/seedflags"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/views"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/validate"
	"github.com/spf13/cobra"
)

const helpMessage = `Create a new BEE (Branch Engineering Environment) from a template

Examples:

# Create a hybrid BEE from the swatomation template
thelma bee create \
  --name=swat-grungy-puma \
  --template=swatomation \
`

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	name                      string
	namePrefix                string
	owner                     string
	template                  string
	generatorOnly             string
	waitHealthy               string
	waitHealthyTimeoutSeconds string
	terraHelmfileRef          string
	seed                      string
	notify                    string
	exportLogsOnFailure       string
}{
	name:                      "name",
	namePrefix:                "name-prefix",
	owner:                     "owner",
	template:                  "template",
	generatorOnly:             "generator-only",
	waitHealthy:               "wait-healthy",
	waitHealthyTimeoutSeconds: "wait-healthy-timeout-seconds",
	terraHelmfileRef:          "terra-helmfile-ref",
	seed:                      "seed",
	notify:                    "notify",
	exportLogsOnFailure:       "export-logs-on-failure",
}

type createCommand struct {
	options   bee.CreateOptions
	pinFlags  pinflags.PinFlags
	seedFlags seedflags.SeedFlags
}

func NewBeeCreateCommand() cli.ThelmaCommand {
	return &createCommand{
		pinFlags: pinflags.NewPinFlags(),
		seedFlags: seedflags.NewSeedFlags(func(options *seedflags.Options) {
			options.Prefix = "seed-"
			options.NoShortHand = true // default short-hand flags could conflict with others in create command
			options.Hidden = false     // we could set to true to hide seed flags in `thelma bee create` output
		}),
	}
}

func (cmd *createCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "create"
	cobraCommand.Short = "Create a new BEE from a template"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.Name, flagNames.name, "n", "NAME", "Name for this BEE. If not given, a name will be generated")
	cobraCommand.Flags().StringVarP(&cmd.options.NamePrefix, flagNames.namePrefix, "p", "bee", "Prefix to use when generating a name for this BEE")
	cobraCommand.Flags().StringVarP(&cmd.options.Owner, flagNames.owner, "o", "", "Email address of the owner of the BEE")
	cobraCommand.Flags().StringVarP(&cmd.options.Template, flagNames.template, "t", "swatomation", "Template to use for this BEE")
	cobraCommand.Flags().BoolVar(&cmd.options.SyncGeneratorOnly, flagNames.generatorOnly, false, "Sync the BEE generator but not the BEE's Argo apps")
	cobraCommand.Flags().BoolVar(&cmd.options.WaitHealthy, flagNames.waitHealthy, true, "Wait for BEE's Argo apps to become healthy after syncing")
	cobraCommand.Flags().IntVar(&cmd.options.WaitHealthTimeoutSeconds, flagNames.waitHealthyTimeoutSeconds, 1200, "How long to wait for BEE's Argo apps to become healthy after syncing")
	cobraCommand.Flags().BoolVar(&cmd.options.Seed, flagNames.seed, true, `Seed BEE after creation (run "thelma bee seed -h" for more info)`)
	cobraCommand.Flags().BoolVar(&cmd.options.ExportLogsOnFailure, flagNames.exportLogsOnFailure, true, `Export container logs to GCS if BEE creation fails)`)
	cobraCommand.Flags().BoolVar(&cmd.options.Notify, flagNames.notify, true, "Attempt to notify the owner via Slack upon success")

	cmd.pinFlags.AddFlags(cobraCommand)
	cmd.seedFlags.AddFlags(cobraCommand)
}

func (cmd *createCommand) PreRun(thelmaApp app.ThelmaApp, ctx cli.RunContext) error {
	// if --name not specified, generate a name for this BEE
	if ctx.CobraCommand().Flags().Changed(flagNames.name) {
		if err := validate.EnvironmentName(cmd.options.Name); err != nil {
			return fmt.Errorf("--%s: %q is not a valid environment name: %v", flagNames.name, cmd.options.Name, err)
		}
		cmd.options.GenerateName = false
	} else {
		if err := validate.EnvironmentNamePrefix(cmd.options.NamePrefix); err != nil {
			return fmt.Errorf("--%s: %q is not a valid environment name prefix: %v", flagNames.namePrefix, cmd.options.NamePrefix, err)
		}
		cmd.options.GenerateName = true
	}

	// validate --template
	bees, err := builders.NewBees(thelmaApp)
	if err != nil {
		return err
	}
	_, err = bees.GetTemplate(cmd.options.Template)
	if err != nil {
		return fmt.Errorf("--%s: %v", flagNames.template, err)
	}

	// validate/load pin and seed options
	pinOptions, err := cmd.pinFlags.GetPinOptions(thelmaApp, ctx)
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

func (cmd *createCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}

	_bee, err := bees.CreateWith(cmd.options)

	if _bee != nil {
		ctx.SetOutput(views.DescribeBee(_bee))
	}
	if err != nil {
		return err
	}
	return nil
}

func (cmd *createCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}
