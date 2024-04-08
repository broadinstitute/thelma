package iap

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Generate a new IAP token and store it in ~/.thelma/credentials, for dsp-devops-super-prod`

type iapCommand struct {
}

func NewAuthIAPCommand() cli.ThelmaCommand {
	return &iapCommand{}
}

func (cmd *iapCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "new-iap"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (cmd *iapCommand) PreRun(app app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *iapCommand) Run(thelmaApp app.ThelmaApp, rc cli.RunContext) error {
	tokenProvider, err := thelmaApp.Clients().NewIAP()
	if err != nil {
		return err
	}

	if err := auth.ForProvider(tokenProvider, rc); err != nil {
		return err
	}

	log.Info().Msgf("Successfully authenticated to dsp-devops-super-prod IAP")

	return nil
}

func (cmd *iapCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}
