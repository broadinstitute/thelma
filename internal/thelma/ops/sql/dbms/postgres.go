package dbms

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"
	"path"
	"strconv"
	"text/template"
)

const postgresImageName = "postgres:15"
const scriptsMount = "/scripts"

var scriptNames = struct {
	init   string
	psqlrc string
}{
	init:   "init.sh",
	psqlrc: "psqlrc",
}

//go:embed resources/psql/psqlrc.gotmpl
var psqlrcTemplate string

//go:embed resources/psql/init.sh
var psqlInitScript string

type postgres struct {
	conn api.Connection
}

func (p postgres) Type() api.DBMS {
	return api.Postgres
}

func (p postgres) PodSpec(settings ClientSettings) (podrun.DBMSSpec, error) {
	psqlrc, err := p.renderPsqlrc(settings)
	if err != nil {
		return podrun.DBMSSpec{}, err
	}

	scripts := map[string][]byte{
		scriptNames.psqlrc: psqlrc,
		scriptNames.init:   []byte(psqlInitScript),
	}

	return podrun.DBMSSpec{
		ContainerImage: postgresImageName,
		Env: map[string]string{
			"PGUSER":            settings.Username,
			"PGPASSWORD":        settings.Password,
			"PGHOST":            settings.Host,
			"PGDATABASE":        settings.Database,
			"PSQLRC":            path.Join(scriptsMount, scriptNames.psqlrc),
			"INIT_CREATE_USERS": strconv.FormatBool(settings.Init.CreateUsers),
			"INIT_RO_USER":      settings.Init.ReadOnlyUser.Name,
			"INIT_RO_PASSWORD":  settings.Init.ReadOnlyUser.Password,
			"INIT_RW_USER":      settings.Init.ReadWriteUser.Name,
			"INIT_RW_PASSWORD":  settings.Init.ReadWriteUser.Password,
		},
		Scripts:      scripts,
		ScriptsMount: scriptsMount,
	}, nil
}

func (p postgres) InitCommand() []string {
	return []string{path.Join(scriptsMount, scriptNames.init), "reinit"}
}

func (p postgres) ShellCommand() []string {
	if p.conn.Options.Shell {
		return []string{"/bin/bash"}
	}

	return []string{api.Postgres.CLIClient()}
}

func (p postgres) ensureTargetDatabaseSelected() error {
	if p.conn.Options.PrivilegeLevel == api.Admin {
		// postgres user can connect to its default database, postgres
		return nil
	}
	// thelma-sql-ro and thelma-sql-srw do not have default databases; if we try
	// to connect without specifying one, psql will return an error
	if p.conn.Options.Database == "" {
		return fmt.Errorf("please specify a target database")
	}
	return nil
}

func (p postgres) renderPsqlrc(settings ClientSettings) ([]byte, error) {
	if err := p.ensureTargetDatabaseSelected(); err != nil {
		return nil, err
	}

	type psqlrcContext struct {
		Prompt  string
		SetRole string
	}

	var ctx psqlrcContext
	ctx.Prompt = p.buildPrompt(settings)
	if p.conn.Options.PrivilegeLevel == api.ReadWrite {
		ctx.SetRole = p.conn.Options.Database
	}

	t, err := template.New(scriptNames.psqlrc).Parse(psqlrcTemplate)
	if err != nil {
		panic(fmt.Errorf("error parsing internal template %s: %v", scriptNames.psqlrc, err))
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, ctx); err != nil {
		return nil, fmt.Errorf("error rendering %s: %v", scriptNames.psqlrc, err)
	}
	return buf.Bytes(), nil
}

// https://stackoverflow.com/a/19155772
var psqlColors = struct {
	green  string
	yellow string
	red    string
}{
	green:  "%[%033[1;32;40m%]", // green
	yellow: "%[%033[1;33;40m%]", // yellow
	red:    "%[%033[1;31;40m%]", // red
}

const endStyle = "%[%033[0m%]"

func (p postgres) buildPrompt(settings ClientSettings) string {
	user := "%n"
	if settings.Nickname != "" {
		user = settings.Nickname
	}

	var host string
	if p.conn.Options.Release != nil {
		host = p.conn.Options.Release.FullName()
	} else {
		if p.conn.Provider == api.Google {
			host = p.conn.GoogleInstance.InstanceName
		} else {
			host = "%m"
		}
	}

	host = abbreviate(host, 24)

	database := "%/"

	prompt := user + "@" + host + " " + database

	var style string
	if p.conn.Instance().IsProd() {
		style = psqlColors.red
	} else if p.conn.Instance().IsShared() {
		style = psqlColors.yellow
	} else {
		style = psqlColors.green
	}
	prompt = style + prompt + endStyle

	prompt += "%# "

	return prompt
}

func abbreviate(s string, maxlen int) string {
	if len(s) <= maxlen {
		return s
	}
	maxlen -= 3 // account for ellipsis "..."

	i := maxlen/2 + maxlen%2
	j := len(s) - (maxlen / 2)

	rs := []rune(s)

	return string(rs[0:i]) + "..." + string(rs[j:])
}
