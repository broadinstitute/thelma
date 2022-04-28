package vault

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth"
	"github.com/broadinstitute/thelma/internal/thelma/clients/vault"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Generate a new Vault token and store it in ~/.vault-token`

type vaultCommand struct{}

func NewAuthVaultCommand() cli.ThelmaCommand {
	return &vaultCommand{}
}

func (cmd *vaultCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "vault"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (cmd *vaultCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *vaultCommand) Run(thelmaApp app.ThelmaApp, rc cli.RunContext) error {
	provider, err := vault.TokenProvider(thelmaApp.Config(), thelmaApp.Credentials())
	if err != nil {
		return err
	}

	if err := auth.ForProvider(provider, rc); err != nil {
		return err
	}

	log.Info().Msgf("Vault token is valid")

	return nil
}

func (cmd *vaultCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}
