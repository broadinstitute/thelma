package connector

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/dbms"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/provider"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/provider/google"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/provider/kubernetes"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
)

type Connector interface {
	Init() error
	Connect() error
}

func New(conn api.Connection, clients clients.Clients) (Connector, error) {
	podrunner, err := podrun.New(conn, clients.Kubernetes())
	if err != nil {
		return nil, err
	}

	var _provider provider.Provider

	switch conn.Provider {
	case api.Google:
		sqladmin, err := clients.Google().SqlAdmin()
		if err != nil {
			return nil, err
		}
		vault, err := clients.Vault()
		if err != nil {
			return nil, err
		}
		_provider = google.New(conn, sqladmin, vault)

	case api.Kubernetes:
		_provider, err = kubernetes.New(conn, clients.Kubernetes())
		if err != nil {
			return nil, err
		}

	case api.Azure:
		panic("TODO")

	default:
		panic(fmt.Errorf("unsupported provider: %#v", conn.Provider))
	}

	_dbms, err := buildDBMSForProvider(conn, _provider)
	if err != nil {
		return nil, err
	}

	return &connector{
		conn:      conn,
		provider:  _provider,
		podrunner: podrunner,
		dbms:      _dbms,
	}, nil
}

type connector struct {
	conn      api.Connection
	provider  provider.Provider
	podrunner podrun.Runner
	dbms      dbms.DBMS
}

func buildDBMSForProvider(c api.Connection, p provider.Provider) (dbms.DBMS, error) {
	t, err := p.DetectDBMS()
	if err != nil {
		return nil, err
	}
	return dbms.New(t, c), nil
}

func (c *connector) Init() error {
	if err := c.provider.Initialize(); err != nil {
		return err
	}

	pod, err := c.createPod()
	if err != nil {
		return err
	}

	err = pod.Exec(c.dbms.InitCommand())

	return utils.CloseWarn(pod, err)
}

func (c *connector) Connect() error {
	if err := c.ensureInitialized(); err != nil {
		return err
	}

	pod, err := c.createPod()
	if err != nil {
		return err
	}

	err = pod.ExecInteractive(c.dbms.ShellCommand())

	return utils.CloseWarn(pod, err)
}

func (c *connector) ensureInitialized() error {
	initalized, err := c.provider.Initialized()
	if err != nil {
		return err
	}
	if !initalized {
		return fmt.Errorf("database instance %s has not been initialized; please run `thelma sql init` and try again", c.conn.Instance().Name())
	}
	return nil
}

func (c *connector) createPod() (podrun.Pod, error) {
	settings, err := c.provider.ClientSettings()
	if err != nil {
		return nil, err
	}

	providerSpec, err := c.provider.PodSpec()
	if err != nil {
		return nil, err
	}

	dbmsSpec, err := c.dbms.PodSpec(settings)
	if err != nil {
		return nil, err
	}

	return c.podrunner.Create(podrun.Spec{
		DBMSSpec:     dbmsSpec,
		ProviderSpec: providerSpec,
	})
}
