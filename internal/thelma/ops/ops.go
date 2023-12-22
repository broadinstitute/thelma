// Package ops contains operational tools for Terra services
package ops

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients"
	"github.com/broadinstitute/thelma/internal/thelma/ops/artifacts"
	"github.com/broadinstitute/thelma/internal/thelma/ops/logs"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql"
	"github.com/broadinstitute/thelma/internal/thelma/ops/status"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sync"
)

type Ops interface {
	Logs() logs.Logs
	Sql() sql.Sql
	Status() (status.Reporter, error)
	Sync() (sync.Sync, error)
}

func NewOps(clients clients.Clients) Ops {
	return &ops{
		clients: clients,
	}
}

type ops struct {
	clients clients.Clients
}

func (o *ops) Logs() logs.Logs {
	return logs.New(o.clients.Kubernetes(), artifacts.New(o.clients.Google()))
}

func (o *ops) Sql() sql.Sql {
	return sql.New(o.clients)
}

func (o *ops) Status() (status.Reporter, error) {
	argocd, err := o.clients.ArgoCD()
	if err != nil {
		return nil, err
	}
	return status.NewReporter(argocd, o.clients.Kubernetes()), nil
}

func (o *ops) Sync() (sync.Sync, error) {
	argocd, err := o.clients.ArgoCD()
	if err != nil {
		return nil, err
	}

	statusReporter, err := o.Status()
	if err != nil {
		return nil, err
	}

	sherlock, err := o.clients.Sherlock()
	if err != nil {
		return nil, err
	}
	return sync.New(argocd, statusReporter, sherlock), nil
}
