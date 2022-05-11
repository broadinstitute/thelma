package builders

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
)

func NewBees(thelmaApp app.ThelmaApp) (bee.Bees, error) {
	_argocd, err := thelmaApp.Clients().ArgoCD()
	if err != nil {
		return nil, err
	}

	return bee.NewBees(_argocd, thelmaApp.StateLoader())
}
