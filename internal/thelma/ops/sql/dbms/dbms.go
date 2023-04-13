package dbms

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"
)

// things we need to know:
// how to set up permissions for thelma users
// initialization files for various things

// DBMS the kind of remote database that Thelma is trying to connect to (Postgres or MySQL)
type DBMS interface {
	// Type returns type of database
	Type() api.DBMS
	// PodSpec returns information about Kubernetes resources that should be created in order to execute commands
	// against the database
	PodSpec(ClientSettings, ...api.ConnectionOverride) (podrun.DBMSSpec, error)
	// InitCommand returns a command that should be run during `thelma sql init` to set up the database
	InitCommand() []string
	// ShellCommand returns a command that should be run during `thelma sql connect` to launch an interactive shell
	ShellCommand() []string
}

type ClientSettings struct {
	// Username to use when connecting to the database
	Username string
	// Password to use when connecting (if empty, passwordless auth will be used)
	Password string
	// Host to connect to (if blank, defaults to "127.0.0.1"
	Host string
	// Database database to connect to (leave blank to not connect to any database)
	Database string
	// Nickname optional short username to use in terminal prompt -- useful for
	// long Cloud IAM usernames like "thelma-sql-rw-terra-dev@broad-dsde-dev.iam"
	Nickname string

	// Init initialization script settings
	Init InitSettings
}

// InitUser settings for initializing a new local database user
type InitUser struct {
	// Name (local username) for the user
	Name string
	// Password (optional). If not empty, the user's password will be set to this value
	Password string
}

// InitSettings controlling Thelma user initialization
type InitSettings struct {
	// CreateUsers if true, create users instead of assuming they exist
	CreateUsers bool
	// ReadOnlyUser information about Thelma's read only user
	ReadOnlyUser InitUser
	// ReadWriteUser information about Thelma's read write user
	ReadWriteUser InitUser
}

func New(t api.DBMS, conn api.Connection) DBMS {
	switch t {
	case api.Postgres:
		return postgres{
			conn: conn,
		}
	case api.MySQL:
		return mysql{}
	default:
		panic(fmt.Errorf("unsupported DBMS: %#v", t))
	}
}
