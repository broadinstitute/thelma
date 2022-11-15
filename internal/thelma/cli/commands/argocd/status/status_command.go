package status

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/selector"
	"github.com/broadinstitute/thelma/internal/thelma/ops/status"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/spf13/cobra"
)

const helpMessage = `Report status for a collection of ArgoCD application(s)`

type statusCommand struct {
	argocd   argocd.ArgoCD
	selector *selector.Selector
}

func NewArgoCDStatusCommand() cli.ThelmaCommand {
	return &statusCommand{
		selector: selector.NewSelector(func(options *selector.Options) {
			options.IncludeBulkFlags = false
			options.RequireDestinationOrExact = true
		}),
	}
}

func (cmd *statusCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "status"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage

	// Release selector flags -- these flags determine which Argo apps will be synced
	cmd.selector.AddFlags(cobraCommand)
}

func (cmd *statusCommand) PreRun(app app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

func (cmd *statusCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	// compute selected releases
	state, err := app.State()
	if err != nil {
		return err
	}
	selection, err := cmd.selector.GetSelection(state, rc.CobraCommand().Flags(), rc.Args())
	if err != nil {
		return err
	}

	releases := selection.Releases

	// build argo client
	_argocd, err := app.Clients().ArgoCD()
	if err != nil {
		return err
	}

	k8sclients := app.Clients().Kubernetes()
	if err != nil {
		return err
	}

	reporter := status.NewReporter(_argocd, k8sclients)
	statuses, err := reporter.Statuses(releases)
	if err != nil {
		return err
	}

	output := make(map[string]map[string]status.Status)
	for release, _status := range statuses {
		destName := release.Destination().Name()
		destMap, exists := output[destName]
		if !exists {
			destMap = make(map[string]status.Status)
		}
		destMap[release.Name()] = _status
		output[destName] = destMap
	}
	rc.SetOutput(output)
	return nil
}

func (cmd *statusCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}
