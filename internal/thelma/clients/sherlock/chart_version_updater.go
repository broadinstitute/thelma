package sherlock

import (
	"fmt"
	"github.com/broadinstitute/sherlock/clients/go/client/changesets"
	"github.com/broadinstitute/sherlock/clients/go/client/chart_releases"
	"github.com/broadinstitute/sherlock/clients/go/client/chart_versions"
	"github.com/broadinstitute/sherlock/clients/go/client/environments"
	"github.com/broadinstitute/sherlock/clients/go/client/models"
	"github.com/rs/zerolog/log"
)

type ChartVersionUpdater interface {
	// UpdateForNewChartVersion does three things in sequence, all directly with Sherlock's API.
	//
	// First, it reports the new chart version to Sherlock. Sherlock will now most likely use that version as the
	// new "latest" chart version, unless Sherlock's handling for replay-requests or concurrency control kicks in.
	//
	// If and only if that first step succeeds, each chart release is set to use the latest chart version. Sherlock
	// automatically recalculates the latest version for each chart release, even if it was already set to use latest.
	// Any version changes are applied automatically, assuming Thelma's authentication has sufficient access for the
	// chart releases requested. If any part of this second step fails, the entire second step will have no effect.
	//
	// Lastly, Thelma looks up any template instances of this chart that are following the latest version of this chart,
	// which would've just gotten updated. Thelma will issue a refresh for any such instances, so that directly
	// template manifests will correctly have the new version.
	UpdateForNewChartVersion(chartSelector string, newVersion string, lastVersion string, description string, chartReleaseSelectors ...string) error
}

func (c *Client) UpdateForNewChartVersion(chartSelector string, newVersion string, lastVersion string, description string, chartReleaseSelectors ...string) error {
	chartVersion := &models.V2controllersCreatableChartVersion{
		Chart:        chartSelector,
		ChartVersion: newVersion,
		Description:  description,
	}
	if lastVersion != "" {
		chartVersion.ParentChartVersion = fmt.Sprintf("%s/%s", chartSelector, lastVersion)
	}
	_, _, err := c.client.ChartVersions.PostAPIV2ChartVersions(
		chart_versions.NewPostAPIV2ChartVersionsParams().WithChartVersion(chartVersion))
	if err != nil {
		return fmt.Errorf("error from Sherlock creating chart version %s/%s: %v", chartSelector, newVersion, err)
	}
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
	_, _, err = c.client.Changesets.PostAPIV2ProceduresChangesetsPlanAndApply(
		changesets.NewPostAPIV2ProceduresChangesetsPlanAndApplyParams().WithChangesetPlanRequest(changesetPlanRequest))
	if err != nil {
		return fmt.Errorf("error from Sherlock updating chart releases to new version %s/%s: %v", chartSelector, newVersion, err)
	} else {
		log.Info().Msgf("updated chart releases in Sherlock to new version %s/%s: %v", chartSelector, newVersion, chartReleaseSelectors)
	}

	templateString := "template"
	templates, err := c.client.Environments.GetAPIV2Environments(
		environments.NewGetAPIV2EnvironmentsParams().
			WithLifecycle(&templateString),
	)
	if err != nil {
		return fmt.Errorf("error from Sherlock getting template environments: %v", err)
	}
	var chartReleasesToRefresh []string
	latestString := "latest"
	followString := "follow"
	for _, template := range templates.Payload {
		chartReleasesUsingLatest, err := c.client.ChartReleases.GetAPIV2ChartReleases(
			chart_releases.NewGetAPIV2ChartReleasesParams().
				WithChart(&chartSelector).
				WithEnvironment(&template.Name).
				WithChartVersionResolver(&latestString),
		)
		if err != nil {
			return fmt.Errorf("error from Sherlock getting latest chart releases in template %s: %v", template.Name, err)
		} else {
			for _, chartRelease := range chartReleasesUsingLatest.Payload {
				chartReleasesToRefresh = append(chartReleasesToRefresh, chartRelease.Name)
			}
		}
		for _, chartReleaseThatGotUpdated := range chartReleaseSelectors {
			chartReleasesUsingFollow, err := c.client.ChartReleases.GetAPIV2ChartReleases(
				chart_releases.NewGetAPIV2ChartReleasesParams().
					WithChart(&chartSelector).
					WithEnvironment(&template.Name).
					WithChartVersionResolver(&followString).
					WithChartVersionFollowChartRelease(&chartReleaseThatGotUpdated),
			)
			if err != nil {
				return fmt.Errorf("error from Sherlock getting chart releases following %s in template %s: %v", chartReleaseThatGotUpdated, template.Name, err)
			} else {
				for _, chartRelease := range chartReleasesUsingFollow.Payload {
					chartReleasesToRefresh = append(chartReleasesToRefresh, chartRelease.Name)
				}
			}
		}
	}

	if len(chartReleasesToRefresh) > 0 {
		var chartReleaseEntriesToRefresh []*models.V2controllersChangesetPlanRequestChartReleaseEntry
		for _, chartReleaseSelector := range chartReleasesToRefresh {
			chartReleaseEntriesToRefresh = append(chartReleaseEntriesToRefresh, &models.V2controllersChangesetPlanRequestChartReleaseEntry{
				ChartRelease: chartReleaseSelector,
			})
		}
		changesetPlanRequest = &models.V2controllersChangesetPlanRequest{
			ChartReleases: chartReleaseEntriesToRefresh,
		}
		_, _, err = c.client.Changesets.PostAPIV2ProceduresChangesetsPlanAndApply(
			changesets.NewPostAPIV2ProceduresChangesetsPlanAndApplyParams().WithChangesetPlanRequest(changesetPlanRequest))
		if err != nil {
			return fmt.Errorf("error from Sherlock refreshing template chart releases to reflect new version %s/%s: %v", chartSelector, newVersion, err)
		} else {
			log.Info().Msgf("refreshed template chart releases in Sherlock to reflect new version %s/%s: %v", chartSelector, newVersion, chartReleasesToRefresh)
		}
	} else {
		log.Info().Msg("no template chart releases to refresh")
	}

	return nil
}
