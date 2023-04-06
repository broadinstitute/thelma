package dbms

import (
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PostgresSuite struct {
	suite.Suite
	p DBMS
}

func (suite *PostgresSuite) SetupTest() {
	suite.p = New(api.Postgres, api.Connection{
		Provider: api.Google,
		GoogleInstance: api.GoogleInstance{
			InstanceName: "my-instance",
			Project:      "my-project",
		},
		Options: api.ConnectionOptions{
			Database:       "my-database",
			PrivilegeLevel: api.ReadOnly,
			Shell:          false,
		},
	})
}

func (suite *PostgresSuite) Test_Type() {
	assert.Equal(suite.T(), api.Postgres, suite.p.Type())
}

func (suite *PostgresSuite) Test_PodSpec() {
	spec, err := suite.p.PodSpec(ClientSettings{
		Username: "ro",
		Password: "ropw",
		Host:     "localhost",
		Database: "my-database",
		Nickname: "",
		Init: InitSettings{
			CreateUsers: false,
			ReadOnlyUser: InitUser{
				Name:     "ro",
				Password: "ropw",
			},
			ReadWriteUser: InitUser{
				Name:     "rw",
				Password: "rwpw",
			},
		},
	})

	require.NoError(suite.T(), err)

	scripts := spec.Scripts
	assert.Equal(suite.T(), psqlInitScript, string(scripts[scriptNames.init]))
	assert.Equal(suite.T(), `/* silence messages from formatting changes */
\set QUIET true

/* set prompt */
\set PROMPT1 '%[%033[1;32;40m%]%n@my-instance %/%[%033[0m%]%# '

\set QUIET false
`, string(scripts[scriptNames.psqlrc]))

	spec.Scripts = nil
	assert.Equal(suite.T(), podrun.DBMSSpec{
		ContainerImage: "postgres:15",
		Env: map[string]string{
			"INIT_CREATE_USERS": "false",
			"INIT_RO_PASSWORD":  "ropw",
			"INIT_RO_USER":      "ro",
			"INIT_RW_PASSWORD":  "rwpw",
			"INIT_RW_USER":      "rw",
			"PGHOST":            "localhost",
			"PGDATABASE":        "my-database",
			"PGPASSWORD":        "ropw",
			"PGUSER":            "ro",
			"PSQLRC":            "/scripts/psqlrc",
		},
		ScriptsMount: scriptsMount,
	}, spec)
}

func (suite *PostgresSuite) Test_InitCommand() {
	assert.Equal(suite.T(), []string{"/scripts/init.sh", "reinit"}, suite.p.InitCommand())
}

func (suite *PostgresSuite) Test_ShellCommand() {
	assert.Equal(suite.T(), []string{"psql"}, suite.p.ShellCommand())
}

func TestPostgres(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}
