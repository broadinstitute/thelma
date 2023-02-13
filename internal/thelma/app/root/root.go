package root

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/env"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/helm"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"path/filepath"
)

//
// The `root` package exists to identify the Thelma root directory, which is where:
// * the Thelma config file lives
// * debug logs are captured
// * thelma installation files live
// and more.
//
// This package is very low-level and is used in both config and logging initialization.
// It should NOT depend on any other Thelma packages.
//

// Users can override Thelma root dir by setting this environment variable ("THELMA_ROOT").
const envVarSuffix = "ROOT"

// Name of the directory inside user's home directory
const dirName = ".thelma"

type Root interface {
	// Dir returns the Thelma installation root directory, usually ~/.thelma
	Dir() string
	// LogDir returns the path where Thelma debug logs are stored ($ROOT/logs)
	LogDir() string
	// CachesDir returns the Thelma cache directory ($ROOT/caches)
	CachesDir() string
	// CredentialsDir returns the path to Thelma credentials directory ($ROOT/credentials)
	CredentialsDir() string
	// ConfigDir returns the path to Thelma config directory
	ConfigDir() string
	// ReleasesDir returns the Thelma installation directory ($ROOT/releases)
	ReleasesDir() ReleasesDir
	// ToolsDir returns the path to Thelma's bundled third-party tool binaries, such as Helm, Helmfile, etc.
	ToolsDir() (ToolsDir, error)
	// CreateDirectories create directories if they do not exist. Will be called as part of Thelma initialization
	CreateDirectories() error
}

// Default returns a Root instance rooted at the default directory returned by DefaultDir
func Default() Root {
	return root{
		dir: DefaultDir(),
	}
}

// New (FOR TESTING ONLY) returns a Root instance rooted at the given directory
func New(dir string) Root {
	return root{
		dir: dir,
	}
}

// DefaultDir derives the default Thelma root directory. It will be:
// * The value of the THELMA_ROOT env var, if set
// * If a valid home directory exists for current user, it will be $HOME/.thelma
// * Else, /tmp/.thelma.<pid> (worst-case fallback option in weird environments)
// Note that this function identifies the root directory path, but does NOT create the root directory; it may or may not exist.
func DefaultDir() string {
	dir, exists := os.LookupEnv(env.WithEnvPrefix(envVarSuffix))
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

// implements Root interface
type root struct {
	dir string
}

func (r root) Dir() string {
	return r.dir
}

func (r root) LogDir() string {
	return path.Join(r.Dir(), "logs")
}

func (r root) CachesDir() string {
	return path.Join(r.Dir(), "caches")
}

func (r root) CredentialsDir() string {
	return path.Join(r.Dir(), "credentials")
}

func (r root) ConfigDir() string {
	return path.Join(r.Dir(), "config")
}

func (r root) ReleasesDir() ReleasesDir {
	return releasesDir{
		dir: path.Join(r.Dir(), "releases"),
	}
}

func (r root) ToolsDir() (ToolsDir, error) {
	dir, err := findToolsDir(r.ReleasesDir())
	if err != nil {
		return nil, err
	}
	return toolsDir{
		dir: dir,
	}, nil
}

func (r root) findToolsDirRelativeToExe() (string, error) {
	exepath, err := os.Executable()
	if err != nil {
		return "", err
	}

	exepath, err = filepath.EvalSymlinks(exepath)
	if err != nil {
		return "", err
	}

	toolsdir := filepath.Clean(path.Join(exepath, "..", "tools", "bin"))

	if err = r.validateToolsDir(toolsdir); err != nil {
		return "", err
	}
	return toolsdir, nil
}

func (r root) validateToolsDir(toolsdir string) error {
	// make sure a helm executable exists in the tools dir
	helmpath := path.Join(toolsdir, helm.ProgName)
	exists, err := utils.FileExists(helmpath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("tools dir not found; %s does not exist", helmpath)
	}
	return nil
}

func (r root) CreateDirectories() error {
	dirs := []string{
		r.CachesDir(),
		r.CredentialsDir(),
		r.ConfigDir(),
		r.LogDir(),
		r.ReleasesDir().Root(),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("error creating directory %s: %v", dir, err)
		}
	}

	return nil
}
