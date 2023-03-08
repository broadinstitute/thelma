package sql

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/connector"
)

const podNamespace = "thelma-workloads"
const thelmaRwUser = "thelma-rw"
const thelmaRoUser = "thelma-ro"

type Sql interface {
	// Init initialize database for Thelma connections.
	Init(c api.Connection) error
	// Connect connect to a database via K8s pod running CLI client
	Connect(c api.Connection) error
}

func New(clients clients.Clients) Sql {
	return &sql{
		clients: clients,
	}
}

type sql struct {
	clients clients.Clients
}

func (s *sql) Init(conn api.Connection) error {
	cxr, err := connector.New(conn, s.clients)
	if err != nil {
		return err
	}
	return cxr.Init()
}

func (s *sql) Connect(conn api.Connection) error {
	cxr, err := connector.New(conn, s.clients)
	if err != nil {
		return err
	}
	return cxr.Connect()
}
