package root

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path"
)

//
// The `root` package exists to identify the Thelma root directory, which is where:
// * the Thelma config file lives
// * debug logs are captured
// * thelma installation files live
//
// This package is very low-level and is used in both config and logging initializaiton.
// It should NOT depend on any other Thelma packages.
//

// Users can override Thelma root dir by setting this environment variable.
const envVar = "THELMA_ROOT"

// Name of the directory inside user's home directory
const dirName = ".thelma"

// Dir returns the path to the thelma installation root.
// It will be:
// * If THELMA_ROOT env var is set, it will be $THELMA_ROOT
// * If a valid home directory exists for current user, it will be $HOME/.thelma
// * Else, /tmp/.thelma.<pid> (worst-case fallback option in weird environments)
// Note that this function identifies the root directory path, but does NOT create the root directory; it may or may not exist.
func Dir() string {
	dir, exists := os.LookupEnv(envVar)
	if exists {
		return dir
	}
	homeDir, err := os.UserHomeDir()
	if err == nil {
		return path.Join(homeDir, dirName)
	}

	dir = path.Join(os.TempDir(), fmt.Sprintf("%s.%d", dirName, os.Getpid()))
	log.Warn().Msgf("Could not identify home dir for current user: %v", err)
	log.Warn().Msgf("Will use temporary root directory: %s", dir)

	return dir
}
