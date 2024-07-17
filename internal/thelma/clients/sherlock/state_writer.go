package sherlock

import (
	"fmt"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/changesets"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/chart_releases"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/charts"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/clusters"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/environments"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/models"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

func init() {
	// This function automatically converts times to UTC when serializing.
	// Without it, time.Time instances that Thelma sends to Sherlock will be stripped of their timezones,
	// meaning that if the local system sends a time set to 3pm EST, the sherlock will receive a timestamp
	// set to 3pm UTC.
	// This is especially relevant to Thelma because it runs on laptops which are set to local time, not UTC.
	// Further reading:
	// https://github.com/go-openapi/strfmt/issues/72
	strfmt.NormalizeTimeForMarshal = func(t time.Time) time.Time {
		return t.UTC()
	}
}

type StateWriter interface {
	terra.StateWriter

	// CreateEnvironmentFromTemplate has the same *effect* as WriteEnvironments for the narrow case of creating a
	// BEE from a template, but it relies entirely on Sherlock for the template-default logic.
	// All the desired fields can be left blank to let Sherlock's defaulting behavior take over.
	// The name of the new environment is returned.
	// Of note here is that calling this function doesn't touch Thelma's in-memory state, only Sherlock's state.
	// Thelma's in-memory state will need to be reloaded to work with the newly-created environment.
	CreateEnvironmentFromTemplate(templateName string, options terra.CreateOptions) (string, error)

	// PinEnvironmentVersions adapts Thelma's pin-override pattern to Sherlock's changeset pattern to apply specific
	// versions to chart releases in an environment.
	PinEnvironmentVersions(environmentName string, versions map[string]terra.VersionOverride) error

	// SetTerraHelmfileRefForEntireEnvironment sets it both for the environment itself and for every chart release it
	// contains. This generally matches the semantics of Thelma's old state bucket, where the environment's value
	// would be used for chart releases if they didn't have their own override. Sherlock enforces that every chart
	// release always contain its own version information, so the closest we can get is just set that value to what
	// we're given here.
	SetTerraHelmfileRefForEntireEnvironment(environment terra.Environment, terraHelmfileRef string) error

	// ResetEnvironmentAndPinToDev is a temporary-ish hack implementation of what Thelma could do with state bucket
	// overrides. Because overrides were stored separately, Thelma could drop them, leading to really nice pin-and-
	// unpin semantics. Sherlock doesn't have this notion, at least not right now. Versions just are, and while this
	// allows cool things like following a specific branch, Sherlock doesn't currently have a mechanism to go back to
	// some default set of versions. The closest we can get is effectively re-pinning the environment to dev. We can
	// think about adding this to Sherlock in the future, because it does have some default-version mechanisms, they're
	// just not exposed in the API like how unpinning would expect.
	ResetEnvironmentAndPinToDev(environment terra.Environment) error

	// SetEnvironmentOffline controls the "offline" flag on a given environment.
	// Of note here is that calling this function doesn't touch Thelma's in-memory state, only Sherlock's state.
	// Thelma's in-memory state will need to be reloaded to work with the mutated environment.
	SetEnvironmentOffline(environmentName string, offline bool) error
}

func (c *clientImpl) CreateEnvironmentFromTemplate(templateName string, options terra.CreateOptions) (string, error) {
	creatableEnvironment := &models.SherlockEnvironmentV3Create{
		TemplateEnvironment: templateName,
	}

	if options.Name != "" {
		creatableEnvironment.Name = options.Name
	}
	if options.Owner != "" {
		creatableEnvironment.Owner = options.Owner
	}
	if options.AutoDelete.Enabled {
		creatableEnvironment.DeleteAfter = strfmt.DateTime(options.AutoDelete.After)
	}
	if options.StopSchedule.Enabled {
		creatableEnvironment.OfflineScheduleBeginEnabled = true
		creatableEnvironment.OfflineScheduleBeginTime = strfmt.DateTime(options.StopSchedule.RepeatingTime)
	}
	if options.StartSchedule.Enabled {
		creatableEnvironment.OfflineScheduleEndEnabled = true
		creatableEnvironment.OfflineScheduleEndTime = strfmt.DateTime(options.StartSchedule.RepeatingTime)
		creatableEnvironment.OfflineScheduleEndWeekends = options.StartSchedule.Weekends
	}
	created, err := c.client.Environments.PostAPIEnvironmentsV3(
		environments.NewPostAPIEnvironmentsV3Params().WithEnvironment(creatableEnvironment))
	if err != nil {
		return "", errors.Errorf("error from Sherlock creating environment from '%s' template: %v", templateName, err)
	} else if created != nil && created.Payload != nil {
		return created.Payload.Name, nil
	} else {
		return "", errors.Errorf("error reading Sherlock response, it didn't respond with an error but the client library couldn't parse a payload")
	}
}

func (c *clientImpl) PinEnvironmentVersions(environmentName string, versions map[string]terra.VersionOverride) error {
	var chartReleaseEntries []*models.SherlockChangesetV3PlanRequestChartReleaseEntry
	for chartName, overrides := range versions {
		entry := &models.SherlockChangesetV3PlanRequestChartReleaseEntry{
			ChartRelease: fmt.Sprintf("%s/%s", environmentName, chartName),
		}
		if overrides.AppVersion != "" {
			entry.ToAppVersionResolver = "exact"
			entry.ToAppVersionExact = overrides.AppVersion
		}
		if overrides.ChartVersion != "" {
			entry.ToChartVersionResolver = "exact"
			entry.ToChartVersionExact = overrides.ChartVersion
		}
		if overrides.TerraHelmfileRef != "" {
			entry.ToHelmfileRef = overrides.TerraHelmfileRef
		}
		chartReleaseEntries = append(chartReleaseEntries, entry)
	}
	changesetPlanRequest := &models.SherlockChangesetV3PlanRequest{
		ChartReleases: chartReleaseEntries,
	}
	_, _, err := c.client.Changesets.PostAPIChangesetsProceduresV3PlanAndApply(
		changesets.NewPostAPIChangesetsProceduresV3PlanAndApplyParams().WithChangesetPlanRequest(changesetPlanRequest))
	if err != nil {
		return errors.Errorf("error from Sherlock setting environment '%s' releases to overrides: %v", environmentName, err)
	}
	return nil
}

func (c *clientImpl) SetTerraHelmfileRefForEntireEnvironment(environment terra.Environment, terraHelmfileRef string) error {
	editableEnvironment := &models.SherlockEnvironmentV3Edit{
		HelmfileRef: &terraHelmfileRef,
	}
	_, err := c.client.Environments.PatchAPIEnvironmentsV3Selector(
		environments.NewPatchAPIEnvironmentsV3SelectorParams().WithEnvironment(editableEnvironment).WithSelector(environment.Name()))
	if err != nil {
		return errors.Errorf("error from Sherlock setting environment '%s' terra-helmfile ref to '%s': %v", environment.Name(), terraHelmfileRef, err)
	}
	var chartReleaseEntries []*models.SherlockChangesetV3PlanRequestChartReleaseEntry
	for _, release := range environment.Releases() {
		chartReleaseEntries = append(chartReleaseEntries, &models.SherlockChangesetV3PlanRequestChartReleaseEntry{
			ChartRelease:  fmt.Sprintf("%s/%s", environment.Name(), release.ChartName()),
			ToHelmfileRef: terraHelmfileRef,
		})
	}
	changesetPlanRequest := &models.SherlockChangesetV3PlanRequest{
		ChartReleases: chartReleaseEntries,
	}
	_, _, err = c.client.Changesets.PostAPIChangesetsProceduresV3PlanAndApply(
		changesets.NewPostAPIChangesetsProceduresV3PlanAndApplyParams().WithChangesetPlanRequest(changesetPlanRequest))
	if err != nil {
		return errors.Errorf("error from Sherlock setting environment '%s' releases terra-helmfile ref to '%s': %v", environment.Name(), terraHelmfileRef, err)
	}
	return nil
}

func (c *clientImpl) ResetEnvironmentAndPinToDev(environment terra.Environment) error {
	editableEnvironment := &models.SherlockEnvironmentV3Edit{
		HelmfileRef: utils.Nullable("HEAD"),
	}
	_, err := c.client.Environments.PatchAPIEnvironmentsV3Selector(
		environments.NewPatchAPIEnvironmentsV3SelectorParams().WithEnvironment(editableEnvironment).WithSelector(environment.Name()))
	if err != nil {
		return errors.Errorf("error from Sherlock unpinning environment '%s': %v", environment.Name(), err)
	}
	changesetPlanRequest := &models.SherlockChangesetV3PlanRequest{
		Environments: []*models.SherlockChangesetV3PlanRequestEnvironmentEntry{
			{
				Environment:                          environment.Name(),
				UseExactVersionsFromOtherEnvironment: "dev",
			},
		},
	}
	_, _, err = c.client.Changesets.PostAPIChangesetsProceduresV3PlanAndApply(
		changesets.NewPostAPIChangesetsProceduresV3PlanAndApplyParams().WithChangesetPlanRequest(changesetPlanRequest))
	if err != nil {
		return errors.Errorf("error from Sherlock pinning environment '%s' to dev: %v", environment.Name(), err)
	}
	return nil
}

func (c *clientImpl) SetEnvironmentOffline(environmentName string, offline bool) error {
	editableEnvironment := &models.SherlockEnvironmentV3Edit{
		Offline: &offline,
	}
	_, err := c.client.Environments.PatchAPIEnvironmentsV3Selector(
		environments.NewPatchAPIEnvironmentsV3SelectorParams().WithSelector(environmentName).WithEnvironment(editableEnvironment))
	return err
}

// WriteEnvironments will take a list of terra.Environment interfaces them and issue POST requests
// to write both the environment and any releases within that environment. 409 Conflict responses are ignored
func (c *clientImpl) WriteEnvironments(envs []terra.Environment) ([]string, error) {
	createdEnvNames := make([]string, 0)
	for _, environment := range envs {
		log.Info().Msgf("exporting state for environment: %s", environment.Name())
		// When exporting state, we don't want Sherlock to try to be smart and interpolate
		// BEE chart releases. We'll create them manually based on our own gitops state
		// in the next step.
		newEnv := toModelCreatableEnvironment(environment, false)

		newEnvRequestParams := environments.NewPostAPIEnvironmentsV3Params().
			WithEnvironment(newEnv)
		createdEnv, err := c.client.Environments.PostAPIEnvironmentsV3(newEnvRequestParams)
		var envAlreadyExists bool
		if err != nil {
			// Don't error if creating the chart results in 409 conflict
			if _, ok := err.(*environments.PostAPIEnvironmentsV3Conflict); !ok {
				return nil, errors.Errorf("error creating cluster: %v", err)
			}
			envAlreadyExists = true
		}

		// extract the generated name from a new dynamic environment
		var envName string
		if environment.Lifecycle().IsDynamic() && !envAlreadyExists {
			envName = createdEnv.Payload.Name
		} else {
			envName = environment.Name()
		}

		log.Debug().Msgf("environment name: %s", envName)
		if err := c.writeReleases(envName, environment.Releases()); err != nil {
			return nil, err
		}
		createdEnvNames = append(createdEnvNames, envName)
	}
	return createdEnvNames, nil
}

// WriteClusters will take a list of terra.Cluster interfaces them and issue POST requests
// to create both the cluster and any releases within that cluster. 409 Conflict responses are ignored
func (c *clientImpl) WriteClusters(cls []terra.Cluster) error {
	for _, cluster := range cls {
		log.Info().Msgf("exporting state for cluster: %s", cluster.Name())
		if cluster.Name() == "dsp-tools-az" {
			log.Warn().Msgf("skipping dsp-tools-az as Thelma does not have support yet")
			continue
		}
		newCluster := toModelCreatableCluster(cluster)
		newClusterRequestParams := clusters.NewPostAPIClustersV3Params().
			WithCluster(newCluster)
		_, err := c.client.Clusters.PostAPIClustersV3(newClusterRequestParams)
		if err != nil {
			// Don't error if creating the chart results in 409 conflict
			if _, ok := err.(*clusters.PostAPIClustersV3Conflict); !ok {
				return errors.Errorf("error creating cluster: %v", err)
			}
		}

		if err := c.writeReleases(cluster.Name(), cluster.Releases()); err != nil {
			return err
		}
	}
	return nil
}

func (c *clientImpl) DeleteEnvironments(envs []terra.Environment) ([]string, error) {
	deletedEnvs := make([]string, 0)
	for _, env := range envs {
		// delete chart releases associated with environment
		releases := env.Releases()
		for _, release := range releases {
			if err := c.deleteRelease(release); err != nil {
				return nil, errors.Errorf("error deleting chart release %s in environment %s: %v", release.Name(), env.Name(), err)
			}
		}
		params := environments.NewDeleteAPIEnvironmentsV3SelectorParams().
			WithSelector(env.Name())

		deletedEnv, err := c.client.Environments.DeleteAPIEnvironmentsV3Selector(params)
		if err != nil {
			return nil, errors.Errorf("error deleting environment %s: %v", env.Name(), err)
		}
		log.Debug().Msgf("%#v", deletedEnv)
		deletedEnvs = append(deletedEnvs, deletedEnv.Payload.Name)
	}
	return deletedEnvs, nil
}

func (c *clientImpl) EnableRelease(env terra.Environment, releaseName string) error {
	// need to pull info about the template env in order to set chart and app versions
	templateEnv, err := c.getEnvironment(env.Template())
	if err != nil {
		return errors.Errorf("unable to fetch template %s: %v", env.Template(), err)
	}
	templateEnvName := templateEnv.Name
	// now look up the chart release to enable in the template
	templateRelease, err := c.getChartRelease(templateEnvName, releaseName)
	if err != nil {
		return errors.Errorf("unable to enable release, error retrieving from template: %v", err)
	}

	// enable the chart-release in environment
	enabledChart := &models.SherlockChartReleaseV3Create{
		AppVersionExact:   templateRelease.AppVersionExact,
		Chart:             templateRelease.Chart,
		ChartVersionExact: templateRelease.ChartVersionExact,
		Environment:       env.Name(),
		HelmfileRef:       templateRelease.HelmfileRef,
		Port:              templateRelease.Port,
		Protocol:          templateRelease.Protocol,
		Subdomain:         templateRelease.Subdomain,
	}
	log.Info().Msgf("enabling chart-release: %q in environment: %q", releaseName, env.Name())
	params := chart_releases.NewPostAPIChartReleasesV3Params().WithChartRelease(enabledChart)
	_, err = c.client.ChartReleases.PostAPIChartReleasesV3(params)
	return err
}

func (c *clientImpl) DisableRelease(envName, releaseName string) error {
	params := chart_releases.NewDeleteAPIChartReleasesV3SelectorParams().WithSelector(strings.Join([]string{envName, releaseName}, "/"))
	_, err := c.client.ChartReleases.DeleteAPIChartReleasesV3Selector(params)
	return err
}

func toModelCreatableEnvironment(env terra.Environment, autoPopulateChartReleases bool) *models.SherlockEnvironmentV3Create {
	// if Helmfile ref isn't set it should default to head
	var helmfileRef string
	if env.TerraHelmfileRef() == "" {
		helmfileRef = "HEAD"
	} else {
		helmfileRef = env.TerraHelmfileRef()
	}
	var deleteAfter strfmt.DateTime
	if env.AutoDelete().Enabled() {
		deleteAfter = strfmt.DateTime(env.AutoDelete().After())
	}
	return &models.SherlockEnvironmentV3Create{
		Base:                      env.Base(),
		BaseDomain:                utils.Nullable(env.BaseDomain()),
		DefaultCluster:            env.DefaultCluster().Name(),
		DefaultNamespace:          env.Namespace(),
		Lifecycle:                 utils.Nullable(env.Lifecycle().String()),
		Name:                      env.Name(),
		NamePrefixesDomain:        utils.Nullable(env.NamePrefixesDomain()),
		RequiresSuitability:       env.RequireSuitable(),
		TemplateEnvironment:       env.Template(),
		HelmfileRef:               utils.Nullable(helmfileRef),
		AutoPopulateChartReleases: &autoPopulateChartReleases,
		UniqueResourcePrefix:      env.UniqueResourcePrefix(),
		PreventDeletion:           utils.Nullable(env.PreventDeletion()),
		DeleteAfter:               deleteAfter,
	}
}

func toModelCreatableCluster(cluster terra.Cluster) *models.SherlockClusterV3Create {
	// Hard coding to google for now since we don't have azure clusters
	provider := "google"
	// if Helmfile ref isn't set it should default to head
	var helmfileRef string
	if cluster.TerraHelmfileRef() == "" {
		helmfileRef = "HEAD"
	} else {
		helmfileRef = cluster.TerraHelmfileRef()
	}
	return &models.SherlockClusterV3Create{
		Address:             cluster.Address(),
		Base:                cluster.Base(),
		Name:                cluster.Name(),
		Provider:            &provider,
		GoogleProject:       cluster.Project(),
		RequiresSuitability: cluster.RequireSuitable(),
		HelmfileRef:         &helmfileRef,
		Location:            utils.Nullable(cluster.Location()),
	}
}

func (c *clientImpl) writeReleases(destinationName string, releases []terra.Release) error {
	// for each release attempt to create a chart
	for _, release := range releases {
		log.Info().Msgf("exporting release: %v", release.Name())
		// attempt to convert to app release
		if release.IsAppRelease() {
			appRelease := release.(terra.AppRelease)
			if err := c.writeAppRelease(destinationName, appRelease); err != nil {
				return err
			}
		} else if release.IsClusterRelease() {
			clusterRelease := release.(terra.ClusterRelease)
			if err := c.writeClusterRelease(clusterRelease); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *clientImpl) writeAppRelease(environmentName string, release terra.AppRelease) error {
	log.Debug().Msgf("release name: %v", release.Name())
	modelChart := models.SherlockChartV3Create{
		Name:            release.ChartName(),
		ChartRepo:       utils.Nullable(release.Repo()),
		DefaultPort:     utils.Nullable(int64(release.Port())),
		DefaultProtocol: utils.Nullable(release.Protocol()),
	}
	// first try to create the chart
	newChartRequestParams := charts.NewPostAPIChartsV3Params().
		WithChart(&modelChart)

	_, err := c.client.Charts.PostAPIChartsV3(newChartRequestParams)
	if err != nil {
		// Don't error if creating the chart results in 409 conflict
		if _, ok := err.(*charts.PostAPIChartsV3Conflict); !ok {
			return errors.Errorf("error creating chart: %v", err)
		}
	}
	// check for a release name override
	var releaseName string
	if release.Name() == release.ChartName() {
		releaseName = strings.Join([]string{release.ChartName(), environmentName}, "-")
	} else {
		releaseName = strings.Join([]string{release.Name(), environmentName}, "-")
	}

	// helmfile ref should default to HEAD if unspecified
	var helmfileRef string
	if release.TerraHelmfileRef() == "" {
		helmfileRef = "HEAD"
	} else {
		helmfileRef = release.TerraHelmfileRef()
	}

	modelChartRelease := models.SherlockChartReleaseV3Create{
		AppVersionExact:   release.AppVersion(),
		Chart:             release.ChartName(),
		ChartVersionExact: release.ChartVersion(),
		Cluster:           release.ClusterName(),
		Environment:       environmentName,
		HelmfileRef:       utils.Nullable(helmfileRef),
		Name:              releaseName,
		Namespace:         release.Namespace(),
		Port:              int64(release.Port()),
		Protocol:          release.Protocol(),
		Subdomain:         release.Subdomain(),
	}

	newChartReleaseRequestParams := chart_releases.NewPostAPIChartReleasesV3Params().
		WithChartRelease(&modelChartRelease)

	_, err = c.client.ChartReleases.PostAPIChartReleasesV3(newChartReleaseRequestParams)
	if err != nil {
		if _, ok := err.(*chart_releases.PostAPIChartReleasesV3Conflict); !ok {
			return errors.Errorf("error creating chart release: %v", err)
		}
	}
	return nil
}

func (c *clientImpl) writeClusterRelease(release terra.ClusterRelease) error {
	modelChart := models.SherlockChartV3Create{
		Name:            release.ChartName(),
		ChartRepo:       utils.Nullable(release.Repo()),
		DefaultPort:     nil,
		DefaultProtocol: nil,
	}

	// first try to create the chart
	newChartRequestParams := charts.NewPostAPIChartsV3Params().
		WithChart(&modelChart)

	_, err := c.client.Charts.PostAPIChartsV3(newChartRequestParams)
	if err != nil {
		// Don't error if creating the chart results in 409 conflict
		if _, ok := err.(*charts.PostAPIChartsV3Conflict); !ok {
			return errors.Errorf("error creating chart: %v", err)
		}
	}

	// check for a release name override
	var releaseName string
	if release.Name() == release.ChartName() {
		releaseName = strings.Join([]string{release.ChartName(), release.ClusterName()}, "-")
	} else {
		releaseName = strings.Join([]string{release.Name(), release.ClusterName()}, "-")
	}
	// helmfile ref should default to HEAD if unspecified
	var helmfileRef string
	if release.TerraHelmfileRef() == "" {
		helmfileRef = "HEAD"
	} else {
		helmfileRef = release.TerraHelmfileRef()
	}
	modelChartRelease := models.SherlockChartReleaseV3Create{
		Chart:             release.ChartName(),
		ChartVersionExact: release.ChartVersion(),
		Cluster:           release.ClusterName(),
		HelmfileRef:       utils.Nullable(helmfileRef),
		Name:              releaseName,
		Namespace:         release.Namespace(),
	}

	newChartReleaseRequestParams := chart_releases.NewPostAPIChartReleasesV3Params().
		WithChartRelease(&modelChartRelease)

	_, err = c.client.ChartReleases.PostAPIChartReleasesV3(newChartReleaseRequestParams)
	if err != nil {
		if _, ok := err.(*chart_releases.PostAPIChartReleasesV3Conflict); !ok {
			return errors.Errorf("error creating chart release: %v", err)
		}
	}
	return nil
}

func (c *clientImpl) deleteRelease(release terra.Release) error {
	params := chart_releases.NewDeleteAPIChartReleasesV3SelectorParams().
		WithSelector(strings.Join([]string{release.ChartName(), release.Destination().Name()}, "-"))
	_, err := c.client.ChartReleases.DeleteAPIChartReleasesV3Selector(params)
	return err
}

func (c *clientImpl) getEnvironment(name string) (*Environment, error) {
	params := environments.NewGetAPIEnvironmentsV3SelectorParams().WithSelector(name)
	environment, err := c.client.Environments.GetAPIEnvironmentsV3Selector(params)
	if err != nil {
		return nil, err
	}

	return &Environment{environment.Payload}, nil
}

func (c *clientImpl) getChartRelease(environmentName, releaseName string) (*Release, error) {
	params := chart_releases.NewGetAPIChartReleasesV3SelectorParams().WithSelector(strings.Join([]string{environmentName, releaseName}, "/"))
	release, err := c.client.ChartReleases.GetAPIChartReleasesV3Selector(params)
	if err != nil {
		return nil, err
	}
	return &Release{release.Payload}, nil
}
