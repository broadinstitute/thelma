package delete

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/views"
	"github.com/spf13/cobra"
)

const helpMessage = `Destroy a BEE (Branch Engineering Environment)

Examples:

thelma bee delete --name=swat-grungy-puma
`

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	name     string
	ifExists string
}{
	name:     "name",
	ifExists: "if-exists",
}

type deleteCommand struct {
	name    string
	options bee.DeleteOptions
}

func NewBeeDeleteCommand() cli.ThelmaCommand {
	return &deleteCommand{}
}

func (cmd *deleteCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "delete"
	cobraCommand.Short = "Destroy a BEE"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.name, flagNames.name, "n", "", "Required. Name of the BEE to delete")
	cobraCommand.Flags().BoolVar(&cmd.options.IgnoreMissing, flagNames.ifExists, false, "Do not return an error if the BEE does not exist")
}

func (cmd *deleteCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	// validate --name
	if !ctx.CobraCommand().Flags().Changed(flagNames.name) {
		return fmt.Errorf("--%s is required", flagNames.name)
	}

	return nil
}

func (cmd *deleteCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	bees, err := builders.NewBees(app)
	env, err := bees.DeleteWith(cmd.name, cmd.options)
	if env != nil {
		rc.SetOutput(views.ForTerraEnv(env))
	}
	return err
}

func (cmd *deleteCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}
