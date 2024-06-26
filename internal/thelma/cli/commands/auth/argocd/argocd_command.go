package argocd

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/clients/iap"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/argocd"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Authenticate to ArgoCD`

type argocdCommand struct{}

func NewAuthArgoCDCommand() cli.ThelmaCommand {
	return &argocdCommand{}
}

func (cmd *argocdCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "argocd"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (cmd *argocdCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *argocdCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	iapToken, err := app.Clients().IAPToken(iap.DspDevopsSuperProd)
	if err != nil {
		return errors.Errorf("failed to retrieve IAP token: %v", err)
	}
	if err = argocd.BrowserLogin(app.Config(), app.ShellRunner(), iapToken); err != nil {
		return err
	}
	log.Info().Msgf("Successfully authenticated to ArgoCD")
	return nil
}

func (cmd *argocdCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}
