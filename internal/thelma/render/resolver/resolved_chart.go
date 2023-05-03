package resolver

import (
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

// ResolvedChart represents the outcome of a successful resolution of a chart release.
type ResolvedChart interface {
	// Path where the resolved chart can be found on disk
	Path() string
	// Version of the resolved chart.
	Version() string
	// SourceDescription of where the chart was found.
	// If local, this will be path on disk relative to current working directory. (eg. "./charts/agora")
	// If repo, this will be the name of the Helm repo. (eg. "terra-helm")
	SourceDescription() string
}

// NewLocallyResolvedChart creates a ResolvedChart based from the local filesystem
func NewLocallyResolvedChart(path string, chartVersion string) ResolvedChart {
	return &locallyResolvedChart{
		path: path, chartVersion: chartVersion,
	}
}

type locallyResolvedChart struct {
	path         string
	chartVersion string
}

func (c *locallyResolvedChart) Path() string {
	return c.path
}

func (c *locallyResolvedChart) Version() string {
	return c.chartVersion
}

func (c *locallyResolvedChart) SourceDescription() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Warn().Msgf("resolver: unexpected error calling os.GetWd(): %v", err)
		return c.path
	}
	abs, err := filepath.Abs(c.path)
	if err != nil {
		log.Warn().Msgf("resolver: unexpected error calling filepath.Abs(): %v", err)
		return c.path
	}
	relPath, err := filepath.Rel(cwd, abs)
	if err != nil {
		log.Warn().Msgf("resolver: unexpected error calling filepath.Rel(): %v", err)
		return c.path
	}

	// Need to use strings.Join instead of path.Join because the latter omits "."
	return strings.Join([]string{".", relPath}, string(os.PathSeparator))
}

// NewRemotelyResolvedChart creates a ResolvedChart based from a remote Helm repo
func NewRemotelyResolvedChart(path string, chartVersion string, chartRepo string) ResolvedChart {
	return &remotelyResolvedChart{
		path: path, chartVersion: chartVersion, chartRepo: chartRepo,
	}
}

type remotelyResolvedChart struct {
	path         string
	chartVersion string
	chartRepo    string
}

func (c *remotelyResolvedChart) Path() string {
	return c.path
}

func (c *remotelyResolvedChart) Version() string {
	return c.chartVersion
}

func (c *remotelyResolvedChart) SourceDescription() string {
	return c.chartRepo
}
