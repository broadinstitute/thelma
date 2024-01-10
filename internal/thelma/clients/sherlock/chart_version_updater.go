package sherlock

import (
	"fmt"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/changesets"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/chart_releases"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/chart_versions"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/environments"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type ChartVersionUpdater interface {
	// UpdateForNewChartVersion does three things in sequence, all directly with Sherlock's API.
	//
	// 1. Report new chart version to Sherlock (meaning there's a new latest chart version)
	// 2. Update given chart releases (via chartReleaseSelectors) to point at the latest chart version and refresh to get the new latest version
	//    - 9 times out of 10 this means `dev/${chartSelector}` based on how .autorelease.yaml file gets defaulted
	// 3. Refresh **template** chart releases that either:
	//    - already follow latest chart version (so step 1 means they'd have an update if we didn't catch them in step 2)
	//    - follow a chart release we just updated in step 2 (so step 2 means they'd have an update)
	UpdateForNewChartVersion(chartSelector string, newVersion string, lastVersion string, description string, chartReleaseSelectors []string) error
}

// Step 1 of UpdateForNewChartVersion
func (c *clientImpl) reportNewChartVersion(chartSelector string, newVersion string, lastVersion string, description string) error {
	chartVersion := &models.SherlockChartVersionV3Create{
		Chart:        chartSelector,
		ChartVersion: newVersion,
		Description:  description,
	}
	if lastVersion != "" {
		chartVersion.ParentChartVersion = fmt.Sprintf("%s/%s", chartSelector, lastVersion)
	}
	_, err := c.client.ChartVersions.PutAPIChartVersionsV3(
		chart_versions.NewPutAPIChartVersionsV3Params().WithChartVersion(chartVersion))
	if err != nil {
		return errors.Errorf("error from Sherlock: %v", err)
	}
	return nil
}

// Step 2 of UpdateForNewChartVersion
func (c *clientImpl) setChartReleasesToLatestChartVersion(chartReleaseSelectors ...string) error {
	var chartReleaseEntriesToUpdate []*models.V2controllersChangesetPlanRequestChartReleaseEntry
	for _, chartReleaseSelector := range chartReleaseSelectors {
		chartReleaseEntriesToUpdate = append(chartReleaseEntriesToUpdate, &models.V2controllersChangesetPlanRequestChartReleaseEntry{
			ChartRelease:           chartReleaseSelector,
			ToChartVersionResolver: "latest",
		})
	}
	changesetPlanRequest := &models.V2controllersChangesetPlanRequest{
		ChartReleases: chartReleaseEntriesToUpdate,
	}
	_, _, err := c.client.Changesets.PostAPIV2ProceduresChangesetsPlanAndApply(
		changesets.NewPostAPIV2ProceduresChangesetsPlanAndApplyParams().WithChangesetPlanRequest(changesetPlanRequest))
	if err != nil {
		return errors.Errorf("error from Sherlock: %v", err)
	}
	return nil
}

// Step 3 of UpdateForNewChartVersion
func (c *clientImpl) refreshDownstreamTemplateChartReleases(chartSelector string, updatedChartReleases ...string) (refreshedChartReleases []string, err error) {
	// Get list of template environments
	templateString := "template"
	templates, err := c.client.Environments.GetAPIV2Environments(
		environments.NewGetAPIV2EnvironmentsParams().
			WithLifecycle(&templateString),
	)
	if err != nil {
		return []string{}, errors.Errorf("error from Sherlock getting template environments: %v", err)
	}

	// Assemble list of all the downstream template chart releases we should update
	var chartReleasesToRefresh []string
	latestString := "latest"
	followString := "follow"
	for _, template := range templates.Payload {
		// First, get applicable template chart releases that are currently following the latest chart version (that
		// we just updated)
		chartReleasesUsingLatest, err := c.client.ChartReleases.GetAPIV2ChartReleases(
			chart_releases.NewGetAPIV2ChartReleasesParams().
				WithChart(&chartSelector).
				WithEnvironment(&template.Name).
				WithChartVersionResolver(&latestString),
		)
		if err != nil {
			return []string{}, errors.Errorf("error from Sherlock getting latest chart releases in template %s: %v", template.Name, err)
		} else {
			for _, chartRelease := range chartReleasesUsingLatest.Payload {
				chartReleasesToRefresh = append(chartReleasesToRefresh, chartRelease.Name)
			}
		}

		// Second, get applicable template chart releases that are currently following a chart release we just specifically
		// updated
		for _, chartReleaseThatGotUpdated := range updatedChartReleases {
			chartReleasesUsingFollow, err := c.client.ChartReleases.GetAPIV2ChartReleases(
				chart_releases.NewGetAPIV2ChartReleasesParams().
					WithChart(&chartSelector).
					WithEnvironment(&template.Name).
					WithChartVersionResolver(&followString).
					WithChartVersionFollowChartRelease(&chartReleaseThatGotUpdated),
			)
			if err != nil {
				return []string{}, errors.Errorf("error from Sherlock getting chart releases following %s in template %s: %v", chartReleaseThatGotUpdated, template.Name, err)
			} else {
				for _, chartRelease := range chartReleasesUsingFollow.Payload {
					chartReleasesToRefresh = append(chartReleasesToRefresh, chartRelease.Name)
				}
			}
		}
	}

	// Create a changeset request to just refresh every chart release we collected
	if len(chartReleasesToRefresh) > 0 {
		var chartReleaseEntriesToRefresh []*models.V2controllersChangesetPlanRequestChartReleaseEntry
		for _, chartReleaseSelector := range chartReleasesToRefresh {
			chartReleaseEntriesToRefresh = append(chartReleaseEntriesToRefresh, &models.V2controllersChangesetPlanRequestChartReleaseEntry{
				ChartRelease: chartReleaseSelector,
			})
		}
		changesetPlanRequest := &models.V2controllersChangesetPlanRequest{
			ChartReleases: chartReleaseEntriesToRefresh,
		}
		_, _, err = c.client.Changesets.PostAPIV2ProceduresChangesetsPlanAndApply(
			changesets.NewPostAPIV2ProceduresChangesetsPlanAndApplyParams().WithChangesetPlanRequest(changesetPlanRequest))
		if err != nil {
			return []string{}, errors.Errorf("error from Sherlock: %v", err)
		}
	}

	return chartReleasesToRefresh, nil
}

func (c *clientImpl) UpdateForNewChartVersion(chartSelector string, newVersion string, lastVersion string, description string, chartReleaseSelectors []string) error {
	if err := c.reportNewChartVersion(chartSelector, newVersion, lastVersion, description); err != nil {
		return errors.Errorf("error reporting chart version %s/%s: %v", chartSelector, newVersion, err)
	}

	if err := c.setChartReleasesToLatestChartVersion(chartReleaseSelectors...); err != nil {
		return errors.Errorf("error setting chart releases to latest chart version (%s/%s): %v", chartSelector, newVersion, err)
	} else {
		log.Info().Msgf("updated chart releases in Sherlock to new version %s/%s: %v", chartSelector, newVersion, chartReleaseSelectors)
	}

	if refreshedChartReleases, err := c.refreshDownstreamTemplateChartReleases(chartSelector, chartReleaseSelectors...); err != nil {
		return errors.Errorf("error refreshing downstream template chart releases after reporting new chart version (%s/%s) and updating the following direct chart releases (%v): %v", chartSelector, newVersion, chartReleaseSelectors, err)
	} else if len(refreshedChartReleases) > 0 {
		log.Info().Msgf("updated further downstream template chart releases in Sherlock to reflect new version %s/%s: %v", chartSelector, newVersion, refreshedChartReleases)
	} else {
		log.Info().Msgf("no further downstream template chart releases in Sherlock to update to reflect new version %s/%s", chartSelector, newVersion)
	}

	return nil
}
