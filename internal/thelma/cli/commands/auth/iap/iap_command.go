package iap

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth"
	"github.com/broadinstitute/thelma/internal/thelma/clients/iap"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Generate a new IAP token and store it in ~/.thelma/credentials`

var flagNames = struct {
	project string
}{
	project: "project",
}

type iapCommand struct {
	project string
}

func NewAuthIAPCommand() cli.ThelmaCommand {
	return &iapCommand{}
}

func (cmd *iapCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "iap"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.project, flagNames.project, "p", "dsp-devops-super-prod", "Project ID to authenticate to")
}

func (cmd *iapCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *iapCommand) Run(thelmaApp app.ThelmaApp, rc cli.RunContext) error {
	project, err := iap.ParseProject(cmd.project)
	if err != nil {
		return err
	}

	tokenProvider, err := thelmaApp.Clients().IAP(project)
	if err != nil {
		return err
	}

	if err := auth.ForProvider(tokenProvider, rc); err != nil {
		return err
	}

	log.Info().Msgf("Successfully authenticated to %s", cmd.project)

	return nil
}

func (cmd *iapCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}
