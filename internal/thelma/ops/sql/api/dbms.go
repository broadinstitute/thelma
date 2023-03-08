package api

import "fmt"

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
		panic(fmt.Errorf("unknown dbms type: %#v", d))
	}
}

func (d DBMS) AdminUser() string {
	switch d {
	case Postgres:
		return "postgres"
	case MySQL:
		return "root"
	default:
		panic(fmt.Errorf("unknown dbms type: %#v", d))
	}
}

func (d DBMS) CLIClient() string {
	switch d {
	case Postgres:
		return "psql"
	case MySQL:
		return "mysql"
	default:
		panic(fmt.Errorf("unknown dbms type: %#v", d))
	}
}
