package sql

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

const helpMessage = `Tools for interacting with Terra SQL databases`

type sqlCommand struct{}

func NewSqlCommand() cli.ThelmaCommand {
	return &sqlCommand{}
}

func (cmd *sqlCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "sql"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (cmd *sqlCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *sqlCommand) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	panic("Run() is only executed for leaf commands")
}

func (cmd *sqlCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}
