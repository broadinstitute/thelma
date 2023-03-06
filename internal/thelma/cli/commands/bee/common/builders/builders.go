package builders

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/bee/cleanup"
	"github.com/broadinstitute/thelma/internal/thelma/bee/seed"
	"github.com/rs/zerolog/log"
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

	slack, err := thelmaApp.Clients().Slack()
	if err != nil {
		// Never error out on Slack issues, downstream calls are resilient to it having failed
		log.Debug().Msgf("error configuring slack client: %v", err)
	}

	stateLoader, err := thelmaApp.StateLoader()
	if err != nil {
		return nil, err
	}
	return bee.NewBees(_argocd, stateLoader, seeder, _cleanup, kubectl, thelmaApp.Ops(), slack)
}

func newSeeder(thelma app.ThelmaApp) (seed.Seeder, error) {
	_kubectl, err := thelma.Clients().Kubernetes().Kubectl()
	if err != nil {
		return nil, fmt.Errorf("error getting kubectl client: %v", err)
	}

	return seed.New(_kubectl, thelma.Clients(), thelma.Config(), thelma.ShellRunner()), nil
}
