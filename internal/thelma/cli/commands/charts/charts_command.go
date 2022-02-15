package charts

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

const helpMessage = `Tools for interacting with Terra Helm charts`

type chartsCommand struct{}

func NewChartsCommand() cli.ThelmaCommand {
	return &chartsCommand{}
}

func (c chartsCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "charts [action]"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (c chartsCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do
	return nil
}

func (c chartsCommand) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	panic("Run() is only executed for leaf commands")
}

func (c chartsCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do
	return nil
}
