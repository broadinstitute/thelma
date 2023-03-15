package create

import (
	"context"
	"fmt"

	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Reports Thelma's version`

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
	if len(ctx.Args()) != 0 {
		return fmt.Errorf("expected 0 arguments, got: %v", ctx.Args())
	}
	return nil
}

func (v *createCommand) Run(thelma app.ThelmaApp, ctx cli.RunContext) error {
	github, err := thelma.Clients().Github()
	if err != nil {
		return err
	}
	u, _, err := github.Users.Get(context.Background(), "")
	if err != nil {
		return err
	}

	log.Debug().Msgf("github user: %+v", u)
	return nil
}

func (v *createCommand) PostRun(app app.ThelmaApp, ctx cli.RunContext) error {
	// nothing to do here
	return nil
}
