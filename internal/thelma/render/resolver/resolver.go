package resolver

import (
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
)

// Mode is an enum type referring to the two types of releases supported by terra-helmfile.
type Mode int

const (
	Development Mode = iota // Development mode: prefer source copies of charts
	Deploy                  // Deploy mode: prefer released versions of charts (in versions/ directories)
)

type Options struct {
	Mode       Mode   // development / deploy
	SourceDir  string // path to chart source directory
	CacheDir   string // path where downloaded charts should be cached
	ScratchDir string // scratch directory where temporary files should be created
}

// Resolver determines where a Helm chart should be sourced from.
type Resolver interface {
	// Resolve determines where a Helm chart should be sourced from.
	// If in "deploy" mode, downloads and unpacks the published chart from the Helm repository.
	// If in "development" mode, runs "helm dependency update" on the local working/source copy of the chart.
	// Under some conditions, "development" mode falls back to downloading published chart and vice versa.
	Resolve(chart ChartRelease) (ResolvedChart, error)
}

type chartResolver struct {
	options        Options
	cache          syncCache[ChartRelease]
	localResolver  localResolver
	remoteResolver remoteResolver
}

func NewResolver(runner shell.Runner, options Options) Resolver {
	local := newLocalResolver(options.SourceDir, runner)
	remote := newRemoteResolver(options.CacheDir, options.ScratchDir, runner)
	r := &chartResolver{
		options:        options,
		localResolver:  local,
		remoteResolver: remote,
	}
	r.cache = newSyncCache(r.resolverFn)
	return r
}

func (r *chartResolver) Resolve(chartRelease ChartRelease) (ResolvedChart, error) {
	return r.cache.get(chartRelease)
}

func (r *chartResolver) resolverFn(chartRelease ChartRelease) (ResolvedChart, error) {
	existsInSource, err := r.localResolver.chartExists(chartRelease)
	if err != nil {
		return nil, err
	}

	if r.options.Mode == Development {
		// In development mode, render from source (unless the chart does not exist in source)
		if !existsInSource {
			// This behavior is necessary to support renders for charts that live outside the terra-helmfile repo. (eg. charts in the datarepo-helm and terra-helm-thirdparty repos).
			log.Warn().Msgf("Chart %s does not exist in source dir %s, will try to download from Helm repo", chartRelease.Name, r.options.SourceDir)
			return r.remoteResolver.resolve(chartRelease)
		}
		return r.localResolver.resolve(chartRelease)
	}

	// We're in deploy mode, so download released version from Helm repo
	resolved, err := r.remoteResolver.resolve(chartRelease)
	if err != nil {
		// Welp, we failed to download the chart from the repo.
		// So try to use source copy, but only if the version in the source's Chart.yaml matches what we've been asked for.
		// This behavior is necessary for supporting renders for new charts that haven't been published yet.
		if !existsInSource {
			return nil, err
		}
		sourceVersion, versionErr := r.localResolver.sourceVersion(chartRelease)
		if versionErr != nil {
			log.Warn().Msgf("error checking source version for %s: %v", chartRelease.Name, versionErr)
			return nil, err
		}
		if sourceVersion == chartRelease.Version {
			log.Warn().Msgf("Failed to download chart %s/%s version %s from Helm repo, will fall back to source copy", chartRelease.Repo, chartRelease.Name, chartRelease.Version)
			return r.localResolver.resolve(chartRelease)
		}
		return nil, err
	}

	return resolved, nil
}
