package validate

import (
	"fmt"
	"regexp"
)

// this is on the long side, but we need to accommodate fiab names like "ngoldman-futile-narwhal"
const maxEnvNameLen = 32
const maxEnvPrefixLen = 16

var envNameRegexp = regexp.MustCompile(`\A[a-z][a-z0-9]*(-[a-z0-9]+)*\z`)

// EnvironmentName returns an error if the given environment name is invalid
func EnvironmentName(name string) error {
	if len(name) > maxEnvNameLen {
		return fmt.Errorf("environment names must be <= %d characters in length", maxEnvNameLen)
	}

	if !envNameRegexp.MatchString(name) {
		return fmt.Errorf("environment name must match regular expression %s", envNameRegexp.String())
	}

	return nil
}

// EnvironmentNamePrefix returns an error if the given environment prefix is invalid
func EnvironmentNamePrefix(prefix string) error {
	if len(prefix) > maxEnvPrefixLen {
		return fmt.Errorf("environment name prefixes must be <= %d characters in length", maxEnvPrefixLen)
	}

	if !envNameRegexp.MatchString(prefix) {
		return fmt.Errorf("environment name prefix must match regular expression %s", envNameRegexp.String())
	}

	return nil
}
