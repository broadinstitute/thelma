package api

import (
	"github.com/pkg/errors"
)

// DBMS kind of database (MySQL or Postgres)
type DBMS int64

const (
	Postgres DBMS = iota
	MySQL
)

func (d DBMS) String() string {
	switch d {
	case Postgres:
		return "Postgres"
	case MySQL:
		return "MySQL"
	default:
		panic(errors.Errorf("unknown dbms type: %#v", d))
	}
}

func (d DBMS) AdminUser() string {
	switch d {
	case Postgres:
		return "postgres"
	case MySQL:
		return "root"
	default:
		panic(errors.Errorf("unknown dbms type: %#v", d))
	}
}

func (d DBMS) CLIClient() string {
	switch d {
	case Postgres:
		return "psql"
	case MySQL:
		return "mysql"
	default:
		panic(errors.Errorf("unknown dbms type: %#v", d))
	}
}
