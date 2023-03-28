package dbms

import (
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"
)

// TODO
// const mysqlAdminUser = "root"

type mysql struct{}

func (m mysql) Type() api.DBMS {
	//TODO implement me
	panic("implement me")
}

func (m mysql) AdminUser() string {
	//TODO implement me
	panic("implement me")
}

func (m mysql) PodSpec(clientSettings ClientSettings) (podrun.DBMSSpec, error) {
	//TODO implement me
	panic("implement me")
}

func (m mysql) InitCommand() []string {
	//TODO implement me
	panic("implement me")
}

func (m mysql) ShellCommand() []string {
	//TODO implement me
	panic("implement me")
}
