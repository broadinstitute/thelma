package init

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	sqlcli "github.com/broadinstitute/thelma/internal/thelma/cli/commands/sql/sqlhelpers"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewSqlInitCommand() cli.ThelmaCommand {
	return sqlcli.AsThelmaCommand(&initCommand{})
}

type initCommand struct {
}

func (c *initCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	// no flags yet
}

func (c *initCommand) Run(conn api.Connection, app app.ThelmaApp, rc cli.RunContext) error {
	// ignore any permission-level settings specified with cli flags; init has to connect
	// as a db admin
	conn.Options.PermissionLevel = api.Admin
	if err := app.Ops().Sql().Init(conn); err != nil {
		return err
	}
	log.Info().Msgf("DB was successfully initialized")
	return nil
}
