package resolver

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/helm"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"os"
	"path"
)

// remoteResolver downloads charts from a Helm repo, unpacking them in the configured directory on disk.
type remoteResolver interface {
	resolve(chart ChartRelease) (ResolvedChart, error)
}

type remoteResolverImpl struct {
	cacheDir   string
	scratchDir string
	cache      syncCache[ChartRelease]
	runner     shell.Runner
}

func newRemoteResolver(cacheDir string, scratchDir string, runner shell.Runner) remoteResolver {
	r := &remoteResolverImpl{
		cacheDir:   cacheDir,
		scratchDir: scratchDir,
		runner:     runner,
	}
	r.cache = newSyncCache(r.resolverFn)
	return r
}

func (r *remoteResolverImpl) resolve(chartRelease ChartRelease) (ResolvedChart, error) {
	return r.cache.get(chartRelease)
}

// Fetch the chart from the Helm repo and unpack in the cache directory
func (r *remoteResolverImpl) resolverFn(chartRelease ChartRelease) (ResolvedChart, error) {
	// Create a tmp dir for downloading and unpacking the chart
	tmpDir := path.Join(r.scratchDir, fmt.Sprintf("%s-%s-%s", chartRelease.Repo, chartRelease.Name, chartRelease.Version))
	if err := os.MkdirAll(tmpDir, 0775); err != nil {
		return nil, errors.Errorf("failed to make tmp dir in %s: %v", r.scratchDir, err)
	}
	defer r.cleanupTmpDir(tmpDir)

	// Run `helm pull` to download the chart into tmp dir
	cmd := shell.Command{
		Prog: helm.ProgName,
		Args: []string{
			"fetch",
			path.Join(chartRelease.Repo, chartRelease.Name),
			"--version",
			chartRelease.Version,
			"--untar",
			"-d",
			tmpDir,
		},
	}

	if err := r.runner.Run(cmd); err != nil {
		return nil, errors.Errorf("error downloading chart %s/%s version %s to %s: %v", chartRelease.Repo, chartRelease.Name, chartRelease.Version, tmpDir, err)
	}

	// Move downloaded chart to correct location in the cache directory
	// ${tmpDir}/${chart} -> ${cacheDir}/${repo}/${chart}-${version}
	cachePath := r.cachePath(chartRelease)

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, err
	}
	if len(entries) != 1 {
		return nil, errors.Errorf("expected exactly one file in %s, got: %v", tmpDir, entries)
	}
	tmpChartPath := path.Join(tmpDir, entries[0].Name())

	log.Debug().Msgf("Rename %s to %s", tmpChartPath, cachePath)
	if err = os.MkdirAll(path.Dir(cachePath), 0775); err != nil {
		return nil, err
	}
	if err = os.Rename(tmpChartPath, cachePath); err != nil {
		return nil, err
	}
	return NewRemotelyResolvedChart(cachePath, chartRelease.Version, chartRelease.Repo), nil
}

// Path in the filesystem where cached chart should be kept.
// eg. "${cacheDir}/terra-helm/agora-1.2.3"
func (r *remoteResolverImpl) cachePath(chart ChartRelease) string {
	return path.Join(r.cacheDir, chart.Repo, fmt.Sprintf("%s-%s", chart.Name, chart.Version))
}

// Cleans up tmp directory, logging error instead of returning so it can be used with `defer`
func (r *remoteResolverImpl) cleanupTmpDir(tmpDir string) {
	err := os.RemoveAll(tmpDir)
	if err != nil {
		log.Warn().Msgf("Error deleting tmp dir %s: %v", tmpDir, err)
	}
}
