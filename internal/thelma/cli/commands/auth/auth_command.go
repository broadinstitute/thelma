package auth

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const helpMessage = `Authentication tools`

type authCommand struct {
	echo  bool
	force bool
}

func NewAuthCommand() cli.ThelmaCommand {
	return &authCommand{}
}

// ForProvider is called by child commands & performs authentication for the given TokenProvider
func ForProvider(provider credentials.TokenProvider, rc cli.RunContext) error {
	cmd, ok := rc.Parent().(*authCommand)
	if !ok {
		panic(errors.Errorf("unexpected parent command type: %v", rc.Parent()))
	}

	return cmd.handleAuth(provider, rc)
}

func (cmd *authCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "auth"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
	cobraCommand.PersistentFlags().BoolVar(&cmd.echo, "echo", false, "Print credentials to STDOUT (be careful!)")
	cobraCommand.PersistentFlags().BoolVar(&cmd.force, "force", false, "Force re-issue of credentials, even if a valid credential is cached")
}

func (cmd *authCommand) PreRun(app app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *authCommand) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	panic("Run() is only executed for leaf commands")
}

func (cmd *authCommand) PostRun(_ app.ThelmaApp, rc cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *authCommand) handleAuth(provider credentials.TokenProvider, rc cli.RunContext) error {
	tokenValue, err := cmd.reissueIfNeeded(provider)
	if err != nil {
		return err
	}

	if cmd.echo {
		rc.SetOutput(string(tokenValue))
	}

	return nil
}

func (cmd *authCommand) reissueIfNeeded(provider credentials.TokenProvider) ([]byte, error) {
	if cmd.force {
		return provider.Reissue()
	} else {
		return provider.Get()
	}
}
