package sherlock

import (
	"fmt"
	"github.com/broadinstitute/sherlock/clients/go/client/changesets"
	"github.com/broadinstitute/sherlock/clients/go/client/chart_versions"
	"github.com/broadinstitute/sherlock/clients/go/client/models"
)

type ChartVersionUpdater interface {
	// UpdateForNewChartVersion does two things in sequence, both directly with Sherlock's API.
	//
	// First, it reports the new chart version to Sherlock. Sherlock will now most likely use that version as the
	// new "latest" chart version, unless Sherlock's handling for replay-requests or concurrency control kicks in.
	//
	// If and only if that first step succeeds, each chart release is set to use the latest chart version. Sherlock
	// automatically recalculates the latest version for each chart release, even if it was already set to use latest.
	// Any version changes are applied automatically, assuming Thelma's authentication has sufficient access for the
	// chart releases requested. If any part of this second step fails, the entire second step will have no effect.
	UpdateForNewChartVersion(chartSelector string, newVersion string, lastVersion string, description string, chartReleaseSelectors ...string) error
}

func (c *Client) UpdateForNewChartVersion(chartSelector string, newVersion string, lastVersion string, description string, chartReleaseSelectors ...string) error {
	chartVersion := &models.V2controllersCreatableChartVersion{
		Chart:        chartSelector,
		ChartVersion: newVersion,
		Description:  description,
	}
	if lastVersion != "" {
		chartVersion.ParentChartVersion = fmt.Sprintf("%s/%s", chartVersion, lastVersion)
	}
	_, _, err := c.client.ChartVersions.PostAPIV2ChartVersions(
		chart_versions.NewPostAPIV2ChartVersionsParams().WithChartVersion(chartVersion))
	if err != nil {
		return fmt.Errorf("error from Sherlock creating chart version %s/%s: %v", chartSelector, newVersion, err)
	}
	var chartReleaseEntries []*models.V2controllersChangesetPlanRequestChartReleaseEntry
	for _, chartReleaseSelector := range chartReleaseSelectors {
		chartReleaseEntries = append(chartReleaseEntries, &models.V2controllersChangesetPlanRequestChartReleaseEntry{
			ChartRelease:           chartReleaseSelector,
			ToChartVersionResolver: "latest",
		})
	}
	changesetPlanRequest := &models.V2controllersChangesetPlanRequest{
		ChartReleases: chartReleaseEntries,
	}
	_, _, err = c.client.Changesets.PostAPIV2ProceduresChangesetsPlanAndApply(
		changesets.NewPostAPIV2ProceduresChangesetsPlanAndApplyParams().WithChangesetPlanRequest(changesetPlanRequest))
	if err != nil {
		return fmt.Errorf("error from Sherlock updating chart releases to new version %s/%s: %v", chartSelector, newVersion, err)
	}
	return nil
}
