package builders

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/bee/cleanup"
	"github.com/broadinstitute/thelma/internal/thelma/bee/seed"
)

func NewBees(thelmaApp app.ThelmaApp) (bee.Bees, error) {
	_argocd, err := thelmaApp.Clients().ArgoCD()
	if err != nil {
		return nil, err
	}

	seeder, err := newSeeder(thelmaApp)
	if err != nil {
		return nil, err
	}

	_cleanup := cleanup.NewCleanup(thelmaApp.Clients().Google())

	kubectl, err := thelmaApp.Clients().Google().Kubectl()
	if err != nil {
		return nil, err
	}

	slack, _ := thelmaApp.Clients().Slack()

	return bee.NewBees(_argocd, thelmaApp.StateLoader(), seeder, _cleanup, kubectl, slack)
}

func newSeeder(thelma app.ThelmaApp) (seed.Seeder, error) {
	_kubectl, err := thelma.Clients().Google().Kubectl()
	if err != nil {
		return nil, fmt.Errorf("error getting kubectl client: %v", err)
	}

	return seed.New(_kubectl, thelma.Clients(), thelma.Config(), thelma.ShellRunner()), nil
}
