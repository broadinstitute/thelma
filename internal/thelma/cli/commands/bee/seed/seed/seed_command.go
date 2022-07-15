package seed

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

type options struct {
	name                     string
	force                    bool
	step1CreateElasticsearch bool
	step2RegisterSaProfiles  bool
	step3AddSaSamPermissions bool
	step4RegisterTestUsers   bool
	step5CreateAgora         bool
	step6ExtraUser           []string
	noSteps                  bool
	ifExists                 bool
	registerSelfShortcut     bool
}

var flagNames = struct {
	name                     string
	force                    string
	step1CreateElasticsearch string
	step2RegisterSaProfiles  string
	step3AddSaSamPermissions string
	step4RegisterTestUsers   string
	step5CreateAgora         string
	step6ExtraUser           string
	noSteps                  string
	ifExists                 string
	registerSelfShortcut     string
}{
	name:                     "name",
	force:                    "force",
	step1CreateElasticsearch: "step-1-create-elasticsearch",
	step2RegisterSaProfiles:  "step-2-register-sa-profiles",
	step3AddSaSamPermissions: "step-3-add-sa-sam-permissions",
	step4RegisterTestUsers:   "step-4-register-test-users",
	step5CreateAgora:         "step-5-create-agora",
	step6ExtraUser:           "step-6-extra-user",
	noSteps:                  "no-steps",
	ifExists:                 "if-exists",
	registerSelfShortcut:     "me",
}

type seedCommand struct {
	options options
}

func NewBeeSeedCommand() cli.ThelmaCommand {
	return &seedCommand{}
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
	cobraCommand.Flags().BoolVarP(&cmd.options.force, flagNames.force, "f", false, "attempt to ignore errors during seeding")
	cobraCommand.Flags().BoolVar(&cmd.options.step1CreateElasticsearch, flagNames.step1CreateElasticsearch, true, "create healthy Ontology index with Elasticsearch")
	cobraCommand.Flags().BoolVar(&cmd.options.step2RegisterSaProfiles, flagNames.step2RegisterSaProfiles, true, "register service account profiles with Orch")
	cobraCommand.Flags().BoolVar(&cmd.options.step3AddSaSamPermissions, flagNames.step3AddSaSamPermissions, true, "add permissions for app service accounts in Sam")
	cobraCommand.Flags().BoolVar(&cmd.options.step4RegisterTestUsers, flagNames.step4RegisterTestUsers, true, "register test user accounts with Orch and accept TOS with Sam")
	cobraCommand.Flags().BoolVar(&cmd.options.step5CreateAgora, flagNames.step5CreateAgora, true, "create Agora's methods repository with Orch")
	cobraCommand.Flags().StringSliceVarP(&cmd.options.step6ExtraUser, flagNames.step6ExtraUser, "u", []string{}, "optionally register extra users for log-in (skipped by default; can specify multiple times; provide `email` address, \"set-adc\", or \"use-adc\")")
	cobraCommand.Flags().BoolVar(&cmd.options.noSteps, flagNames.noSteps, false, "convenience flag to skip all unspecified steps, which would otherwise run by default")
	cobraCommand.Flags().BoolVar(&cmd.options.ifExists, flagNames.ifExists, false, "do not return an error if the BEE does not exist")
	cobraCommand.Flags().BoolVar(&cmd.options.registerSelfShortcut, flagNames.registerSelfShortcut, false, "shorthand for --step-6-extra-user use-adc")

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

	// handle --no-steps
	if cmd.options.noSteps {
		if !flags.Changed(flagNames.step1CreateElasticsearch) {
			cmd.options.step1CreateElasticsearch = false
		}
		if !flags.Changed(flagNames.step2RegisterSaProfiles) {
			cmd.options.step2RegisterSaProfiles = false
		}
		if !flags.Changed(flagNames.step3AddSaSamPermissions) {
			cmd.options.step3AddSaSamPermissions = false
		}
		if !flags.Changed(flagNames.step4RegisterTestUsers) {
			cmd.options.step4RegisterTestUsers = false
		}
		if !flags.Changed(flagNames.step5CreateAgora) {
			cmd.options.step5CreateAgora = false
		}
		// No need to handle step6ExtraUser; it is empty and does nothing by default
	}

	// handle --me
	if cmd.options.registerSelfShortcut {
		cmd.options.step6ExtraUser = append(cmd.options.step6ExtraUser, "use-adc")
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
		if cmd.options.ifExists {
			log.Warn().Msgf("BEE %s not found, it could be a vanilla FiaB or not might exist at all", cmd.options.name)
			log.Info().Msgf("Cannot seed, exiting normally to due --%s", flagNames.ifExists)
			return nil
		}
		return fmt.Errorf("BEE %s not found, it could be a vanilla FiaB or not might exist at all", cmd.options.name)
	}
	if !env.Lifecycle().IsDynamic() {
		err = cmd.handleErrorWithForce(
			fmt.Errorf("environment %s has a lifecycle of %s, instead of %s", env.Name(), env.Lifecycle().String(), terra.Dynamic.String()),
		)
		if err != nil {
			return err
		}
	}
	appReleases := make(map[string]terra.AppRelease)
	for _, release := range env.Releases() {
		if release.IsAppRelease() {
			appRelease, wasAppRelease := release.(terra.AppRelease)
			if wasAppRelease {
				appReleases[appRelease.Name()] = appRelease
			} else {
				log.Warn().Msgf("%s was an App Release but failed to type-assert", release.Name())
			}
		}
	}

	if cmd.options.step1CreateElasticsearch {
		if err := cmd.handleErrorWithForce(cmd.step1CreateElasticsearch(app, appReleases)); err != nil {
			return err
		}
	}
	if cmd.options.step2RegisterSaProfiles {
		if err := cmd.handleErrorWithForce(cmd.step2RegisterSaProfiles(app, appReleases)); err != nil {
			return err
		}
	}
	if cmd.options.step3AddSaSamPermissions {
		if err := cmd.handleErrorWithForce(cmd.step3AddSaSamPermissions(app, appReleases)); err != nil {
			return err
		}
	}
	if cmd.options.step4RegisterTestUsers {
		if err := cmd.handleErrorWithForce(cmd.step4RegisterTestUsers(app, appReleases)); err != nil {
			return err
		}
	}
	if cmd.options.step5CreateAgora {
		if err := cmd.handleErrorWithForce(cmd.step5CreateAgora(app, appReleases)); err != nil {
			return err
		}
	}
	if len(cmd.options.step6ExtraUser) > 0 {
		if err := cmd.handleErrorWithForce(cmd.step6ExtraUser(app, appReleases)); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *seedCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}

func (cmd *seedCommand) handleErrorWithForce(err error) error {
	if err != nil && cmd.options.force {
		log.Warn().Msgf("%v", err.Error())
		log.Warn().Msgf("Continuing despite above error due to --%s", flagNames.force)
		return nil
	} else {
		return err
	}
}
