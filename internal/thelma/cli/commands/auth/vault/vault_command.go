package vault

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/clients/vault"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Generate a new Vault token and store it in ~/.vault-token`

type vaultCommand struct {
	force bool
}

func NewAuthVaultCommand() cli.ThelmaCommand {
	return &vaultCommand{}
}

func (cmd *vaultCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "vault"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
	cobraCommand.Flags().BoolVarP(&cmd.force, "force", "f", false, "Issue new Vault token even if existing token is valid")
}

func (cmd *vaultCommand) PreRun(app app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *vaultCommand) Run(thelmaApp app.ThelmaApp, rc cli.RunContext) error {
	if cmd.force {
		if err := vault.BackupToken(); err != nil {
			return err
		}
	}

	client, err := vault.NewClient(thelmaApp.Config(), thelmaApp.Credentials())
	if err != nil {
		return err
	}
	log.Info().Msgf("Vault token is valid")
	rc.SetOutput(client.Token())
	return nil
}

func (cmd *vaultCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}
