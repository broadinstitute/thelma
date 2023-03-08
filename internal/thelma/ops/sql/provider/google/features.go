package google

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"google.golang.org/api/sqladmin/v1"
	"strings"
)

const mysql56VersionString = "MYSQL_5_6"

type features struct {
	iamSupported bool
	iamEnabled   bool
	dbms         api.DBMS
	region       string
}

func extractFeatures(instance *sqladmin.DatabaseInstance) (features, error) {
	var f features

	_dbms, err := parseCloudSQLVersion(instance.DatabaseVersion)
	if err != nil {
		return f, err
	}

	f.dbms = _dbms
	f.iamSupported = instance.DatabaseVersion != mysql56VersionString

	flags := make(map[string]string)
	for _, flag := range instance.Settings.DatabaseFlags {
		flags[flag.Name] = flag.Value
	}

	if flags[cloudSqlIamAuthenticationFlag] == cloudSqlFlagEnabled {
		f.iamEnabled = true
	} else {
		f.iamEnabled = false
	}

	f.region = instance.Region

	return f, nil
}

func parseCloudSQLVersion(version string) (api.DBMS, error) {
	if strings.HasPrefix(version, "POSTGRES") {
		return api.Postgres, nil
	}
	if strings.HasPrefix(version, "MYSQL") {
		return api.MySQL, nil
	}
	return -1, fmt.Errorf("unsupported CloudSQL version: %q", version)
}
