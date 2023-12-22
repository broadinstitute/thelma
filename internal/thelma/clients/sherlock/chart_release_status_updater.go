package sherlock

import (
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/ci_runs"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/models"
	"github.com/rs/zerolog/log"
)

type ChartReleaseStatusUpdater interface {
	// UpdateChartReleaseStatuses sends the given chart release statuses to Sherlock.
	// This function does nothing if the GHA OIDC token provider isn't happy
	// (since Sherlock wouldn't have a chance of correlating our request to a CiRun).
	UpdateChartReleaseStatuses(chartReleaseStatuses map[string]string) error
}

func (c *clientImpl) UpdateChartReleaseStatuses(chartReleaseStatuses map[string]string) error {
	if !c.ghaOidcTokenProviderIsHappy {
		return nil
	} else if created, err := c.client.CiRuns.PutAPICiRunsV3(
		ci_runs.
			NewPutAPICiRunsV3Params().
			WithCiRun(&models.SherlockCiRunV3Upsert{ChartReleaseStatuses: chartReleaseStatuses}),
	); err != nil {
		log.Warn().Err(err).Msg("failed to report chart release statuses to Sherlock; this will manifest as incomplete/misleading info in Slack and Beehive")
		return err
	} else if created == nil {
		log.Warn().Msg("Sherlock didn't return an error receiving chart release statuses but the response was nil; this is perhaps an issue with the client library")
		return nil
	} else if payload := created.Payload; payload == nil {
		log.Warn().Msg("Sherlock didn't return an error receiving chart release statuses but the response payload was nil; this is perhaps an issue with the client library")
		return nil
	} else {
		var chartReleasesWithStatuses, changesetsWithStatuses int
		for _, relatedResource := range payload.RelatedResources {
			if relatedResource != nil && relatedResource.ResourceStatus != "" {
				switch relatedResource.ResourceType {
				case "chart-release":
					chartReleasesWithStatuses++
				case "changeset":
					changesetsWithStatuses++
				}
			}
		}
		log.Debug().Msgf("Sherlock CiRun %d updated; now providing custom statuses for %d chart releases and %d changesets", payload.ID, chartReleasesWithStatuses, changesetsWithStatuses)
		return nil
	}
}
