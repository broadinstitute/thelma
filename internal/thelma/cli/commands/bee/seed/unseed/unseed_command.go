package unseed

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
	name                    string
	force                   bool
	step1UnregisterAllUsers bool
	noSteps                 bool
}

var flagNames = struct {
	name                    string
	force                   string
	step1UnregisterAllUsers string
	noSteps                 string
}{
	name:                    "name",
	force:                   "force",
	step1UnregisterAllUsers: "step-1-unregister-all-users",
	noSteps:                 "no-steps",
}

type unseedCommand struct {
	options options
}

func NewBeeUnseedCommand() cli.ThelmaCommand {
	return &unseedCommand{}
}

func (cmd *unseedCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "unseed [options]"
	cobraCommand.Aliases = []string{"deseed", "de-seed", "un-seed"}
	cobraCommand.Short = "Delete BEE resources that have non-BEE side-effects"
	cobraCommand.Long = `Delete BEE resources that have non-BEE side effects, like cloud resources.

Designed for parity with the older FiaB unpopulation scripts from firecloud-develop.

Individual steps can be disabled with --flag=false, or you can use --no-steps and just enable the ones you want.
Steps generally aren't safe to run multiple times: the BEE should be seeded before running steps again.

This command is a partial inverse of "thelma bee seed". While most of that command's effects are limited
to the BEE itself--meaning that "thelma bee reset" would fully wipe them--some actions have impacts
on cloud resources outside the confines of the BEE. This command aims to more gracefully delete those
resources in the BEE, to give apps an opportunity to clean up their resources.

Note that this is not necessarily confined to just what "thelma bee seed" does. For example, that command
will register a small set of users, but this command goes out of its way to un-register *all* users of the BEE,
to account for people registering themselves outside of seeding.

Examples (you'd need to set the --name of your environment):
  - To run the plain, FiaB-like de-population process:
      $ thelma bee unseed
  - To skip step 1:
      $ thelma bee unseed --step-1-unregister-all-users=false
`

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "", "required; name of the BEE to seed")
	cobraCommand.Flags().BoolVarP(&cmd.options.force, flagNames.force, "f", false, "attempt to ignore errors during seeding")
	cobraCommand.Flags().BoolVar(&cmd.options.step1UnregisterAllUsers, flagNames.step1UnregisterAllUsers, true, "unregister all user accounts with Sam")
	cobraCommand.Flags().BoolVar(&cmd.options.noSteps, flagNames.noSteps, false, "convenience flag to skip all unspecified steps, which would otherwise run by default")

}

func (cmd *unseedCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
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
		if !flags.Changed(flagNames.step1UnregisterAllUsers) {
			cmd.options.step1UnregisterAllUsers = false
		}
	}
	return nil
}

func (cmd *unseedCommand) Run(app app.ThelmaApp, _ cli.RunContext) error {
	state, err := app.State()
	if err != nil {
		return err
	}
	env, err := state.Environments().Get(cmd.options.name)
	if err != nil {
		return err
	}
	if env == nil {
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

	if cmd.options.step1UnregisterAllUsers {
		if err := cmd.handleErrorWithForce(cmd.step1UnregisterAllUsers(app, appReleases)); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *unseedCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}

func (cmd *unseedCommand) handleErrorWithForce(err error) error {
	if err != nil && cmd.options.force {
		log.Warn().Msg(err.Error())
		log.Warn().Msg("Continuing despite above error due to --force")
		return nil
	} else {
		return err
	}
}
