package sync

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/common"
	"github.com/broadinstitute/thelma/internal/thelma/cli/selector"
	"github.com/spf13/cobra"
)

const helpMessage = `Sync a collection of ArgoCD application(s)`

type syncOptions struct {
	maxParallel int
}

type syncCommand struct {
	selector *selector.Selector
	options  syncOptions
}

func NewArgoCDSyncCommand() cli.ThelmaCommand {
	return &syncCommand{
		selector: selector.NewSelector(func(options *selector.Options) {
			options.IncludeBulkFlags = false
			options.RequireDestinationOrExact = true
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
	// nothing to do yet
	return nil
}

func (cmd *syncCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	// compute selected releases
	state, err := app.State()
	if err != nil {
		return err
	}
	selection, err := cmd.selector.GetSelection(state, rc.CobraCommand().Flags(), rc.Args())
	if err != nil {
		return err
	}

	// workaround for DDO-2476
	// okay so, ideally we would have per-environment level-1 generators for every single environment
	// here is the thing.
	// we want to be able to sync JUST THE SPECIFIC ARGO app.
	//
	_sync, err := app.Ops().Sync()
	if err != nil {
		return err
	}
	statuses, err := _sync.Sync(selection.Releases, cmd.options.maxParallel)

	rc.SetOutput(common.ReleaseMapToStructuredView(statuses))
	return err
}

func (cmd *syncCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}
