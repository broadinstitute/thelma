package provision

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/views"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

const helpMessage = `Re-provision the resources for an existing BEE (Branch Engineering Environment)

Examples:

thelma bee re-provision --name=bee-swat-ecstatic-spider
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

type reProvisionCommand struct {
	name    string
	options bee.ProvisionExistingOptions
}

func NewBeeReProvisionCommand() cli.ThelmaCommand {
	return &reProvisionCommand{}
}

func (cmd *reProvisionCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "re-provision"
	cobraCommand.Short = "Re-provision the resources for an existing BEE"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.name, flagNames.name, "n", "NAME", "Name of the existing BEE to provision")
	cobraCommand.Flags().BoolVar(&cmd.options.SyncGeneratorOnly, flagNames.generatorOnly, false, "Sync the BEE generator but not the BEE's Argo apps")
	cobraCommand.Flags().BoolVar(&cmd.options.WaitHealthy, flagNames.waitHealthy, true, "Wait for BEE's Argo apps to become healthy after syncing")
	cobraCommand.Flags().IntVar(&cmd.options.WaitHealthTimeoutSeconds, flagNames.waitHealthyTimeoutSeconds, 1200, "How long to wait for BEE's Argo apps to become healthy after syncing")
	cobraCommand.Flags().BoolVar(&cmd.options.Notify, flagNames.notify, true, "Attempt to notify the owner via Slack upon success")

}

func (cmd *reProvisionCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	if !ctx.CobraCommand().Flags().Changed(flagNames.name) {
		return fmt.Errorf("no environment name specified; --%s is required", flagNames.name)
	}
	if strings.TrimSpace(cmd.name) == "" {
		log.Warn().Msg("Is Thelma running in CI? Check that you're setting the name of your environment when running your job")
		return fmt.Errorf("no environment name specified; --%s was passed but no name was given", flagNames.name)
	}

	return nil
}

func (cmd *reProvisionCommand) Run(thelmaApp app.ThelmaApp, ctx cli.RunContext) error {
	bees, err := builders.NewBees(thelmaApp)
	if err != nil {
		return err
	}
	_bee, err := bees.ReProvisionWith(cmd.name, cmd.options)
	if _bee != nil {
		ctx.SetOutput(views.DescribeBee(_bee))
	}
	return err
}

func (cmd *reProvisionCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}
