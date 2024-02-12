package create

import (
	"github.com/pkg/errors"
	"time"

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
	owner                     string
	template                  string
	generatorOnly             string
	waitHealthy               string
	waitHealthyTimeoutSeconds string
	terraHelmfileRef          string
	seed                      string
	notify                    string
	exportLogsOnFailure       string
	deleteAfter               string
	dailyStopTime             string
	dailyStartTime            string
	dailyStartWeekends        string
}{
	name:                      "name",
	owner:                     "owner",
	template:                  "template",
	generatorOnly:             "generator-only",
	waitHealthy:               "wait-healthy",
	waitHealthyTimeoutSeconds: "wait-healthy-timeout-seconds",
	terraHelmfileRef:          "terra-helmfile-ref",
	seed:                      "seed",
	notify:                    "notify",
	exportLogsOnFailure:       "export-logs-on-failure",
	deleteAfter:               "delete-after",
	dailyStopTime:             "daily-stop-time",
	dailyStartTime:            "daily-start-time",
	dailyStartWeekends:        "daily-start-weekends",
}

type options struct {
	bee.CreateOptions
	deleteAfter        time.Duration
	dailyStopTime      string
	dailyStartTime     string
	dailyStartWeekends bool
}

type createCommand struct {
	options   options
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
	cobraCommand.Flags().StringVarP(&cmd.options.Owner, flagNames.owner, "o", "", "Email address of the owner of the BEE")
	cobraCommand.Flags().StringVarP(&cmd.options.Template, flagNames.template, "t", "swatomation", "Template to use for this BEE")
	cobraCommand.Flags().BoolVar(&cmd.options.SyncGeneratorOnly, flagNames.generatorOnly, false, "Sync the BEE generator but not the BEE's Argo apps")
	cobraCommand.Flags().BoolVar(&cmd.options.WaitHealthy, flagNames.waitHealthy, true, "Wait for BEE's Argo apps to become healthy after syncing")
	cobraCommand.Flags().IntVar(&cmd.options.WaitHealthTimeoutSeconds, flagNames.waitHealthyTimeoutSeconds, 1800, "How long to wait for BEE's Argo apps to become healthy after syncing")
	cobraCommand.Flags().BoolVar(&cmd.options.Seed, flagNames.seed, true, `Seed BEE after creation (run "thelma bee seed -h" for more info)`)
	cobraCommand.Flags().BoolVar(&cmd.options.ExportLogsOnFailure, flagNames.exportLogsOnFailure, true, `Export container logs to GCS if BEE creation fails)`)
	cobraCommand.Flags().BoolVar(&cmd.options.Notify, flagNames.notify, true, "Attempt to notify the owner via Slack upon success")
	cobraCommand.Flags().DurationVar(&cmd.options.deleteAfter, flagNames.deleteAfter, 0, "Automatically delete this BEE after a period of time (eg. 4h)")

	cobraCommand.Flags().StringVar(&cmd.options.dailyStopTime, flagNames.dailyStopTime, "", "An ISO-8601 time (repeating daily) to stop the BEE.")
	cobraCommand.Flags().StringVar(&cmd.options.dailyStartTime, flagNames.dailyStartTime, "", "An ISO-8601 time (repeating weekdays) to stop the BEE.")
	cobraCommand.Flags().BoolVar(&cmd.options.dailyStartWeekends, flagNames.dailyStartWeekends, false, "If the daily start time should also apply on weekend days.")

	cmd.pinFlags.AddFlags(cobraCommand)
	cmd.seedFlags.AddFlags(cobraCommand)
}

func (cmd *createCommand) PreRun(thelmaApp app.ThelmaApp, ctx cli.RunContext) error {
	// Validate name
	if ctx.CobraCommand().Flags().Changed(flagNames.name) {
		if err := validate.EnvironmentName(cmd.options.Name); err != nil {
			return errors.Errorf("--%s: %q is not a valid environment name: %v", flagNames.name, cmd.options.Name, err)
		}
	}

	if ctx.CobraCommand().Flags().Changed(flagNames.deleteAfter) {
		cmd.options.AutoDelete.Enabled = true
		cmd.options.AutoDelete.After = time.Now().Add(cmd.options.deleteAfter)
	}

	if ctx.CobraCommand().Flags().Changed(flagNames.dailyStopTime) {
		cmd.options.StopSchedule.Enabled = true
		t, err := time.Parse(time.RFC3339, cmd.options.dailyStopTime)
		if err != nil {
			return errors.Errorf("%s was an invalid time: %v", cmd.options.dailyStopTime, err)
		}
		cmd.options.StopSchedule.RepeatingTime = t
	}

	if ctx.CobraCommand().Flags().Changed(flagNames.dailyStartTime) {
		cmd.options.StartSchedule.Enabled = true
		t, err := time.Parse(time.RFC3339, cmd.options.dailyStartTime)
		if err != nil {
			return errors.Errorf("%s was an invalid time: %v", cmd.options.dailyStartTime, err)
		}
		cmd.options.StartSchedule.RepeatingTime = t
		cmd.options.StartSchedule.Weekends = cmd.options.dailyStartWeekends
	}

	// validate --template
	bees, err := builders.NewBees(thelmaApp)
	if err != nil {
		return err
	}
	_, err = bees.GetTemplate(cmd.options.Template)
	if err != nil {
		return errors.Errorf("--%s: %v", flagNames.template, err)
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

	_bee, err := bees.CreateWith(cmd.options.CreateOptions)

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
