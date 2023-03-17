package create

import (
	"context"

	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Creates New DataBiosphere repos from an existing template and performs a number of common setup operations`

type createCommand struct{}

func NewCreateCommand() cli.ThelmaCommand {
	return &createCommand{}
}

func (v *createCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "create"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (v *createCommand) PreRun(app app.ThelmaApp, ctx cli.RunContext) error {
	// nothing to do here
	return nil
}

func (v *createCommand) Run(thelma app.ThelmaApp, ctx cli.RunContext) error {
	github, err := thelma.Clients().Github()
	if err != nil {
		return err
	}
	userInfo, err := github.GetCallingUser(context.Background())
	if err != nil {
		return err
	}

	log.Debug().Msgf("github user: %+v", userInfo)
	return nil
}

func (v *createCommand) PostRun(app app.ThelmaApp, ctx cli.RunContext) error {
	// nothing to do here
	return nil
}
