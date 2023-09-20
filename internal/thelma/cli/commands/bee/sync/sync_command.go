package provision

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/views"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

const helpMessage = `Sync the resources for an existing BEE (Branch Engineering Environment)

This is different from thelma argocd sync because this command is aware of BEE structure and will exhaustively
sync an entire BEE to make sure that added or deleted chart releases are correctly reflected.

Examples:

thelma bee sync --name=bee-swat-ecstatic-spider
`

var flagNames = struct {
	name                      string
	generatorOnly             string
	waitHealthy               string
	waitHealthyTimeoutSeconds string
	notify                    string
}{
	name:                      "name",
	generatorOnly:             "generator-only",
	waitHealthy:               "wait-healthy",
	waitHealthyTimeoutSeconds: "wait-healthy-timeout-seconds",
	notify:                    "notify",
}

type syncCommand struct {
	name    string
	options bee.ProvisionExistingOptions
}

func NewBeeSyncCommand() cli.ThelmaCommand {
	return &syncCommand{}
}

func (cmd *syncCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "sync"
	cobraCommand.Short = "Sync the resources for an existing BEE"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.name, flagNames.name, "n", "NAME", "Name of the existing BEE to sync")
	cobraCommand.Flags().BoolVar(&cmd.options.SyncGeneratorOnly, flagNames.generatorOnly, false, "Sync the BEE generator but not the BEE's Argo apps")
	cobraCommand.Flags().BoolVar(&cmd.options.WaitHealthy, flagNames.waitHealthy, true, "Wait for BEE's Argo apps to become healthy after syncing")
	cobraCommand.Flags().IntVar(&cmd.options.WaitHealthTimeoutSeconds, flagNames.waitHealthyTimeoutSeconds, 1200, "How long to wait for BEE's Argo apps to become healthy after syncing")
	cobraCommand.Flags().BoolVar(&cmd.options.Notify, flagNames.notify, true, "Attempt to notify the owner via Slack upon success")

}

func (cmd *syncCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	if !ctx.CobraCommand().Flags().Changed(flagNames.name) {
		return errors.Errorf("no environment name specified; --%s is required", flagNames.name)
	}
	if strings.TrimSpace(cmd.name) == "" {
		log.Warn().Msg("Is Thelma running in CI? Check that you're setting the name of your environment when running your job")
		return errors.Errorf("no environment name specified; --%s was passed but no name was given", flagNames.name)
	}

	return nil
}

func (cmd *syncCommand) Run(thelmaApp app.ThelmaApp, ctx cli.RunContext) error {
	bees, err := builders.NewBees(thelmaApp)
	if err != nil {
		return err
	}
	_bee, err := bees.SyncWith(cmd.name, cmd.options)
	if _bee != nil {
		ctx.SetOutput(views.DescribeBee(_bee))
	}
	return err
}

func (cmd *syncCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}
