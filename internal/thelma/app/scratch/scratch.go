package scratch

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"os"
)

// thelma configuration key for this package
const configPrefix = "scratch"

// Naming pattern for root dir
const rootPattern = "thelma-scratch"

// Scratch is a utility for creating temporary directories that are automatically cleaned up when Thelma exits
type Scratch interface {
	// Mkdir creates a new scratch directory with the given nickname
	Mkdir(nickname string) (string, error)
	// Cleanup deletes all scratch directories created with CreateScratchDir(), UNLESS the CleanupOnExit config setting is false. It should be called when Thelma exits.
	Cleanup() error
}

type scratch struct {
	rootDir string
	cfg     *scratchConfig
}

type scratchConfig struct {
	TmpDir        string // Directory where tmp directories should be created. Default is OS tmp dir
	CleanupOnExit bool   `default:"true"` // If true, clean up this process's scratch directory on exit. Else, leave it for inspection.
}

func NewScratch(config config.Config) (Scratch, error) {
	cfg := &scratchConfig{}
	if err := config.Unmarshal(configPrefix, cfg); err != nil {
		return nil, err
	}
	if cfg.TmpDir == "" {
		cfg.TmpDir = os.TempDir()
	}
	scratchRoot, err := os.MkdirTemp(cfg.TmpDir, rootPattern)
	if err != nil {
		return nil, errors.Errorf("error creating scratch directory in %s for process: %v", cfg.TmpDir, err)
	}

	return &scratch{
		rootDir: scratchRoot,
		cfg:     cfg,
	}, nil
}

func (s *scratch) Mkdir(nickname string) (string, error) {
	dir, err := os.MkdirTemp(s.rootDir, nickname)
	if err != nil {
		return "", err
	}
	log.Debug().Msgf("Created scratch directory %s", dir)
	return dir, nil
}

func (s *scratch) Cleanup() error {
	if !s.cfg.CleanupOnExit {
		log.Warn().Msgf("Won't clean up scratch directory %s", s.rootDir)
		return nil
	}
	log.Debug().Msgf("Deleting scratch root directory %s", s.rootDir)
	return os.RemoveAll(s.rootDir)
}
