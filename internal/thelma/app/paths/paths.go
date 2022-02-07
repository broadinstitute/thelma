package paths

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"path"
)

const etcDir = "etc"
const chartsDir = "charts"

// Paths provides information about terra-helmfile paths
type Paths interface {
	// ChartsDir returns default directory in terra-helmfile where chart sources live
	ChartsDir() string
	// EtcDir returns directory in terra-helmfile containing miscellaneous config files
	EtcDir() string
}

type paths struct {
	cfg config.Config
}

// New constructor for Paths object.
func New(cfg config.Config) (Paths, error) {
	p := &paths{
		cfg: cfg,
	}
	return p, nil
}

func (p *paths) ChartsDir() string {
	return path.Join(p.cfg.Home(), chartsDir)
}

func (p *paths) EtcDir() string {
	return path.Join(p.cfg.Home(), etcDir)
}
