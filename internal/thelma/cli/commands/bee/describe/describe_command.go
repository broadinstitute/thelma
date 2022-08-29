package list

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/views"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statebucket"
	"github.com/spf13/cobra"
)

const helpMessage = `Print detailed information about a BEE

Examples:

thelma bee describe -n <name>
`

type options struct {
	name string
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	name string
}{
	name: "name",
}

type describeCommand struct {
	options options
}

func NewBeeDescribeCommand() cli.ThelmaCommand {
	return &describeCommand{}
}

func (cmd *describeCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "describe"
	cobraCommand.Short = "Describe BEEs"
	cobraCommand.Long = helpMessage
	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "NAME", "Required. Name of the BEE to describe")
}

func (cmd *describeCommand) PreRun(_ app.ThelmaApp, rc cli.RunContext) error {
	// validate --name
	if !rc.CobraCommand().Flags().Changed(flagNames.name) {
		return fmt.Errorf("no environment name specified; --%s is required", flagNames.name)
	}
	return nil
}

func (cmd *describeCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	// only show dynamic environments
	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}

	bee, err := bees.GetBee(cmd.options.name)
	if err != nil {
		return err
	}

	dynEnv, err := getDynamicEnvironment(app, bee)
	if err != nil {
		return err
	}

	view := views.DescribeBee(bee, dynEnv)

	rc.SetOutput(view)

	return nil
}

func getDynamicEnvironment(thelmaApp app.ThelmaApp, bee terra.Environment) (statebucket.DynamicEnvironment, error) {
	var empty statebucket.DynamicEnvironment

	sb, err := statebucket.New(thelmaApp.Config(), thelmaApp.Clients().Google())
	if err != nil {
		return empty, err
	}
	dynEnvs, err := sb.Environments()
	if err != nil {
		return empty, err
	}

	for _, d := range dynEnvs {
		if d.Name == bee.Name() {
			return d, nil
		}
	}

	return empty, fmt.Errorf("no dynamic environment with name %q", bee.Name())
}

func (cmd *describeCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}
