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

	kubectl, err := thelmaApp.Clients().Kubernetes().Kubectl()
	if err != nil {
		return nil, err
	}

	return bee.NewBees(_argocd, thelmaApp.StateLoader(), seeder, _cleanup, kubectl)
}

func newSeeder(thelma app.ThelmaApp) (seed.Seeder, error) {
	_kubectl, err := thelma.Clients().Kubernetes().Kubectl()
	if err != nil {
		return nil, fmt.Errorf("error getting kubectl client: %v", err)
	}

	return seed.New(_kubectl, thelma.Clients(), thelma.Config(), thelma.ShellRunner()), nil
}
