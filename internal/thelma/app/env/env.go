// Package env includes utilities for interacting with environment variables
package env

import "fmt"

// EnvPrefix is the prefix that should be used for all environment variables used by Thelma
const EnvPrefix = "THELMA_"

func WithEnvPrefix(envVarName string) string {
	return fmt.Sprintf("%s%s", EnvPrefix, envVarName)
}
