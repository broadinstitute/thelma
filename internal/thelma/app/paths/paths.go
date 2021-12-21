package paths

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/rs/zerolog/log"
	"os"
	"path"
)

const etcDir = "etc"
const chartsDir = "charts"

// Paths is a utility for interacting with terra-helmfile paths
type Paths struct {
	cfg            *config.Config
	scratchRootDir string
}

// New constructor for Paths object.
func New(cfg *config.Config) (*Paths, error) {
	scratchDir, err := os.MkdirTemp(cfg.Tmpdir(), "thelma-scratch")
	if err != nil {
		return nil, err
	}
	paths := &Paths{
		cfg:            cfg,
		scratchRootDir: scratchDir,
	}
	return paths, nil
}

// ChartsDir default directory in terra-helmfile where chart sources live
func (p *Paths) ChartsDir() string {
	return path.Join(p.cfg.Home(), chartsDir)
}

// EtcDir directory in terra-helmfile containing miscellaneous config files
func (p *Paths) EtcDir() string {
	return path.Join(p.cfg.Home(), etcDir)
}

// CreateScratchDir creates a new temporary directory
func (p *Paths) CreateScratchDir(nickname string) (string, error) {
	dir, err := os.MkdirTemp(p.scratchRootDir, nickname)
	if err != nil {
		return "", err
	}
	log.Debug().Msgf("Created scratch directory %s", dir)
	return dir, nil
}

// Cleanup will clean up all temporary/scratch directories
func (p *Paths) Cleanup() error {
	log.Debug().Msgf("Deleting scratch root directory %s", p.scratchRootDir)
	return os.RemoveAll(p.scratchRootDir)
}
