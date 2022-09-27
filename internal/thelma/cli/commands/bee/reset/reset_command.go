package reset

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

const helpMessage = `Wipe the data from a Bee's persistent volumes (for, eg. MySQL and Postgres)`

type options struct {
	name string
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	name string
}{
	name: "name",
}

type resetCommand struct {
	options options
}

func NewBeeResetCommand() cli.ThelmaCommand {
	return &resetCommand{}
}

func (cmd *resetCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "reset [options]"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "", "Required. Name of the BEE to reset statefulsets for")
}

func (cmd *resetCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
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

func (cmd *resetCommand) Run(app app.ThelmaApp, _ cli.RunContext) error {
	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}

	env, err := bees.GetBee(cmd.options.name)
	if err != nil {
		return err
	}

	return bees.ResetStatefulSets(env)
}

func (cmd *resetCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}
