package seed

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/seedflags"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
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
	options   options
	seedflags seedflags.SeedFlags
}

func NewBeeSeedCommand() cli.ThelmaCommand {
	return &seedCommand{
		seedflags: seedflags.NewSeedFlags(),
	}
}

func (cmd *seedCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "seed [options]"
	cobraCommand.Short = "Configure a BEE with permissions and test data"
	cobraCommand.Long = `Configure a BEE with permissions and test data.

Designed for parity with the older FiaB population scripts from firecloud-develop.

Individual steps can be disabled with --flag=false, or you can use --no-steps and just enable the ones you want.
Steps generally aren't safe to run multiple times: the BEE should be unseeded and reset before running steps again.

For convenience, the seed process can register extra users with the BEE, with --step-6-extra-user (-u).
  - Set like --step-6-extra-user set-adc, it will run a command to set new GCP Application Default Credentials,
    then use them to register. These credentials will persist on your computer after the command is complete.
  - Set like --step-6-extra-user use-adc, it will register your computer's existing local Application Default Credentials.
    If the ADC lacks permission to access your first/last name, Thelma will try to find it on your computer, falling
    back to placeholders.
  - Set like --step-6-extra-user <email>, it will try to use Firecloud Orchestration's service account with
    domain-wide delegation to register the user. This will generally only work if the email is in the
    test.firecloud.org domain (if a dev BEE) or the quality.firecloud.org domain (if a QA BEE).
This flag can be set multiple times. You can run it on its own if you also pass --no-steps.
This step is safe to run repeatedly, too, so long as it is registering different users each time.

Examples (you'd need to set the --name of your environment):
  - To run the plain, FiaB-like population process:
      $ thelma bee seed
  - To skip normal seeding (suppose it's already run) and set an ADC user to register:
      $ thelma bee seed --no-steps -u set-adc
  - To run the plain, FiaB-like population process, and also register your existing ADC user, this time with the long form:
      $ thelma bee seed --step-6-extra-user use-adc
  - To run the plain, FiaB-like population process, and also register two test firecloud users:
      $ thelma bee seed --step-6-extra-user somebody@test.firecloud.org --step-6-extra-user someone@test.firecloud.org
  - To skip step 5:
      $ thelma bee seed --step-5-create-agora=false
  - To run just steps 1 and 2
      $ thelma bee seed --no-steps --step-1-create-elasticsearch --step-2-register-sa-profiles
`

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "", "required; name of the BEE to seed")

	cmd.seedflags.AddFlags(cobraCommand)
}

func (cmd *seedCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	flags := ctx.CobraCommand().Flags()

	// validate --name
	if !flags.Changed(flagNames.name) {
		return fmt.Errorf("no environment name specified; --%s is required", flagNames.name)
	}
	if strings.TrimSpace(cmd.options.name) == "" {
		log.Warn().Msg("Is Thelma running in CI? Check that you're setting the name of your environment when running your job")
		return fmt.Errorf("no environment name specified; --%s was passed but no name was given", flagNames.name)
	}

	return nil
}

func (cmd *seedCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	seedOptions, err := cmd.seedflags.GetOptions(rc.CobraCommand())
	if err != nil {
		return err
	}

	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}

	env, err := bees.GetBee(cmd.options.name)
	if err != nil {
		return err
	}

	return bees.Seeder().Seed(env, seedOptions)
}

func (cmd *seedCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}
