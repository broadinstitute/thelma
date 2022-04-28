package iap

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/clients/iap"
	"github.com/broadinstitute/thelma/internal/thelma/clients/vault"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Generate a new IAP token and store it in ~/.thelma/credentials`

type iapCommand struct {
}

func NewAuthIAPCommand() cli.ThelmaCommand {
	return &iapCommand{}
}

func (cmd *iapCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "iap"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (cmd *iapCommand) PreRun(app app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *iapCommand) Run(thelmaApp app.ThelmaApp, rc cli.RunContext) error {
	vaultClient, err := vault.NewClient(thelmaApp.Config(), thelmaApp.Credentials())
	if err != nil {
		return err
	}
	idToken, err := iap.GetIDToken(thelmaApp.Config(), thelmaApp.Credentials(), vaultClient)
	if err != nil {
		return err
	}
	log.Info().Msgf("IAP identity token is valid")
	rc.SetOutput(idToken)
	return nil
}

func (cmd *iapCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}
