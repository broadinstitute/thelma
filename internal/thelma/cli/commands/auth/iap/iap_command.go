package iap

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
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
	return fmt.Errorf("TODO")
}

func (cmd *iapCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}
