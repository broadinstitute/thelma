package delete

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/views"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

const helpMessage = `Destroy a BEE (Branch Engineering Environment)

Examples:

thelma bee delete --name=swat-grungy-puma
`

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	name       string
	unseed     string
	exportLogs string
}{
	name:       "name",
	unseed:     "unseed",
	exportLogs: "export-logs",
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
	cobraCommand.Flags().BoolVar(&cmd.options.Unseed, flagNames.unseed, true, "Attempt to unseed BEE before deleting")
	cobraCommand.Flags().BoolVar(&cmd.options.ExportLogs, flagNames.exportLogs, true, "If true, export BEE's logs to GCS before deleting")
}

func (cmd *deleteCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	// validate --name
	if !ctx.CobraCommand().Flags().Changed(flagNames.name) {
		return errors.Errorf("no environment name specified; --%s is required", flagNames.name)
	}
	if strings.TrimSpace(cmd.name) == "" {
		log.Warn().Msg("Is Thelma running in CI? Check that you're setting the name of your environment when running your job")
		return errors.Errorf("no environment name specified; --%s was passed but no name was given", flagNames.name)
	}
	return nil
}

func (cmd *deleteCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}
	_bee, err := bees.DeleteWith(cmd.name, cmd.options)
	if _bee != nil {
		rc.SetOutput(views.DescribeBee(_bee))
	}
	return err
}

func (cmd *deleteCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}
