package sync

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/selector"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/spf13/cobra"
)

const helpMessage = `Sync a collection of ArgoCD application(s)`

type syncOptions struct {
	maxParallel int
}

type syncCommand struct {
	argocd   argocd.ArgoCD
	selector *selector.Selector
	releases []terra.Release
	options  syncOptions
}

func NewArgoCDSyncCommand() cli.ThelmaCommand {
	return &syncCommand{
		selector: selector.NewSelector(func(options *selector.Options) {
			options.IncludeBulkFlags = false
			options.RequireDestination = true
		}),
	}
}

func (cmd *syncCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "sync"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage

	// Release selector flags -- these flags determine which Argo apps will be synced
	cmd.selector.AddFlags(cobraCommand)

	cobraCommand.Flags().IntVarP(&cmd.options.maxParallel, "max-parallel", "p", 15, "Max number of ArgoCD apps to sync simultaneously")
}

func (cmd *syncCommand) PreRun(app app.ThelmaApp, ctx cli.RunContext) error {
	// build argo client
	_argocd, err := app.Clients().ArgoCD()
	if err != nil {
		return err
	}
	cmd.argocd = _argocd

	// compute selected releases
	state, err := app.State()
	if err != nil {
		return err
	}
	selection, err := cmd.selector.GetSelection(state, ctx.CobraCommand().Flags(), ctx.Args())
	if err != nil {
		return err
	}

	cmd.releases = selection.Releases
	return nil
}

func (cmd *syncCommand) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	return cmd.argocd.SyncReleases(cmd.releases, cmd.options.maxParallel)
}

func (cmd *syncCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}
