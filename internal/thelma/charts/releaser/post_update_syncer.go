package releaser

import (
	"github.com/broadinstitute/thelma/internal/thelma/ops/sync"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
)

// PostUpdateSyncer is a helper interface for syncing chart releases after their chart versions have been updated
type PostUpdateSyncer interface {
	// Sync given a list of chart release full names (eg. ["agora-dev", "yale-terra-dev"]),
	// sync the chart releases in parallel
	Sync(chartReleaseNames []string) error
}

func NewPostUpdateSyncer(syncFactory func() (sync.Sync, error), state terra.State, dryRun bool) PostUpdateSyncer {
	return &postUpdateSyncer{
		dryRun:      dryRun,
		syncFactory: syncFactory,
		state:       state,
	}
}

type postUpdateSyncer struct {
	dryRun      bool
	syncFactory func() (sync.Sync, error)
	state       terra.State
}

func (p *postUpdateSyncer) Sync(chartReleaseNames []string) error {
	if p.dryRun {
		log.Info().Msgf("%d chart releases to sync; skipping since this is a dry run", len(chartReleaseNames))
		return nil
	}

	log.Info().Msgf("Syncing %d chart releases: %s", len(chartReleaseNames), strings.Join(chartReleaseNames, ", "))
	releases, err := p.namesToChartReleases(chartReleaseNames)
	if err != nil {
		return errors.Errorf("deployed version updater: error looking up chart releases: %v", err)
	}

	if len(releases) == 0 {
		log.Info().Msg("No chart releases to sync")
		return nil
	}

	syncer, err := p.syncFactory()
	if err != nil {
		return errors.Errorf("deployed version updater: error creating sync wrapper: %v", err)
	}

	_, err = syncer.Sync(releases, maxParallelSync)
	return err
}

func (p *postUpdateSyncer) namesToChartReleases(chartReleaseNames []string) ([]terra.Release, error) {
	var matchingReleases []terra.Release
	allReleases, err := p.state.Releases().All()
	if err != nil {
		return nil, err
	}

	for _, name := range chartReleaseNames {
		var match terra.Release
		for _, release := range allReleases {
			if release.FullName() == name {
				match = release
				break
			}
		}
		if match == nil {
			log.Warn().Msgf("Won't sync chart release %s because it doesn't exist", name)
		} else {
			matchingReleases = append(matchingReleases, match)
		}
	}

	return matchingReleases, nil
}
