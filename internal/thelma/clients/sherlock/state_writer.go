package sherlock

import (
	"fmt"
	"github.com/broadinstitute/sherlock/clients/go/client/changesets"
	"strings"

	"github.com/broadinstitute/sherlock/clients/go/client/chart_releases"
	"github.com/broadinstitute/sherlock/clients/go/client/charts"
	"github.com/broadinstitute/sherlock/clients/go/client/clusters"
	"github.com/broadinstitute/sherlock/clients/go/client/environments"
	"github.com/broadinstitute/sherlock/clients/go/client/models"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
)

type StateWriter interface {
	terra.StateWriter

	// CreateEnvironmentFromTemplate has the same *effect* as WriteEnvironments for the narrow case of creating a
	// BEE from a template, but it relies entirely on Sherlock for the template-default logic.
	// All the desired fields can be left blank to let Sherlock's defaulting behavior take over.
	// The name of the new environment is returned.
	// Of note here is that calling this function doesn't touch Thelma's in-memory state, only Sherlock's state.
	// Thelma's in-memory state will need to be reloaded to work with the newly-created environment.
	CreateEnvironmentFromTemplate(templateName string, desiredNamePrefix string, desiredName string, desiredOwnerEmail string) (string, error)

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
}

func (c *Client) CreateEnvironmentFromTemplate(templateName string, desiredNamePrefix string, desiredName string, desiredOwnerEmail string) (string, error) {
	creatableEnvironment := &models.V2controllersCreatableEnvironment{
		TemplateEnvironment: templateName,
	}
	if desiredNamePrefix != "" {
		creatableEnvironment.NamePrefix = desiredNamePrefix
	}
	if desiredName != "" {
		creatableEnvironment.Name = desiredName
	}
	if desiredOwnerEmail != "" {
		creatableEnvironment.Owner = desiredOwnerEmail
	}
	existing, created, err := c.client.Environments.PostAPIV2Environments(
		environments.NewPostAPIV2EnvironmentsParams().WithEnvironment(creatableEnvironment))
	if err != nil {
		return "", fmt.Errorf("error from Sherlock creating environment from '%s' template: %v", templateName, err)
	} else if existing != nil && existing.Payload != nil {
		return "", fmt.Errorf("error handling Sherlock response, it didn't create a new environment and said that '%s' already matched the request", existing.Payload.Name)
	} else if created != nil && created.Payload != nil {
		return created.Payload.Name, nil
	} else {
		return "", fmt.Errorf("error reading Sherlock response, it didn't respond with an error but the client library couldn't parse a payload")
	}
}

func (c *Client) PinEnvironmentVersions(environmentName string, versions map[string]terra.VersionOverride) error {
	var chartReleaseEntries []*models.V2controllersChangesetPlanRequestChartReleaseEntry
	for chartName, overrides := range versions {
		entry := &models.V2controllersChangesetPlanRequestChartReleaseEntry{
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
		if overrides.FirecloudDevelopRef != "" {
			entry.ToFirecloudDevelopRef = overrides.FirecloudDevelopRef
		}
		chartReleaseEntries = append(chartReleaseEntries, entry)
	}
	changesetPlanRequest := &models.V2controllersChangesetPlanRequest{
		ChartReleases: chartReleaseEntries,
	}
	_, _, err := c.client.Changesets.PostAPIV2ProceduresChangesetsPlanAndApply(
		changesets.NewPostAPIV2ProceduresChangesetsPlanAndApplyParams().WithChangesetPlanRequest(changesetPlanRequest))
	if err != nil {
		return fmt.Errorf("error from Sherlock setting environment '%s' releases to overrides: %v", environmentName, err)
	}
	return nil
}

func (c *Client) SetTerraHelmfileRefForEntireEnvironment(environment terra.Environment, terraHelmfileRef string) error {
	editableEnvironment := &models.V2controllersEditableEnvironment{
		HelmfileRef: &terraHelmfileRef,
	}
	_, err := c.client.Environments.PatchAPIV2EnvironmentsSelector(
		environments.NewPatchAPIV2EnvironmentsSelectorParams().WithEnvironment(editableEnvironment).WithSelector(environment.Name()))
	if err != nil {
		return fmt.Errorf("error from Sherlock setting environment '%s' terra-helmfile ref to '%s': %v", environment.Name(), terraHelmfileRef, err)
	}
	var chartReleaseEntries []*models.V2controllersChangesetPlanRequestChartReleaseEntry
	for _, release := range environment.Releases() {
		chartReleaseEntries = append(chartReleaseEntries, &models.V2controllersChangesetPlanRequestChartReleaseEntry{
			ChartRelease:  fmt.Sprintf("%s/%s", environment.Name(), release.ChartName()),
			ToHelmfileRef: terraHelmfileRef,
		})
	}
	changesetPlanRequest := &models.V2controllersChangesetPlanRequest{
		ChartReleases: chartReleaseEntries,
	}
	_, _, err = c.client.Changesets.PostAPIV2ProceduresChangesetsPlanAndApply(
		changesets.NewPostAPIV2ProceduresChangesetsPlanAndApplyParams().WithChangesetPlanRequest(changesetPlanRequest))
	if err != nil {
		return fmt.Errorf("error from Sherlock setting environment '%s' releases terra-helmfile ref to '%s': %v", environment.Name(), terraHelmfileRef, err)
	}
	return nil
}

func (c *Client) ResetEnvironmentAndPinToDev(environment terra.Environment) error {
	editableEnvironment := &models.V2controllersEditableEnvironment{
		HelmfileRef:                utils.Nullable("HEAD"),
		DefaultFirecloudDevelopRef: utils.Nullable("dev"),
	}
	_, err := c.client.Environments.PatchAPIV2EnvironmentsSelector(
		environments.NewPatchAPIV2EnvironmentsSelectorParams().WithEnvironment(editableEnvironment).WithSelector(environment.Name()))
	if err != nil {
		return fmt.Errorf("error from Sherlock unpinning environment '%s': %v", environment.Name(), err)
	}
	changesetPlanRequest := &models.V2controllersChangesetPlanRequest{
		Environments: []*models.V2controllersChangesetPlanRequestEnvironmentEntry{
			{
				Environment:                          environment.Name(),
				UseExactVersionsFromOtherEnvironment: "dev",
			},
		},
	}
	_, _, err = c.client.Changesets.PostAPIV2ProceduresChangesetsPlanAndApply(
		changesets.NewPostAPIV2ProceduresChangesetsPlanAndApplyParams().WithChangesetPlanRequest(changesetPlanRequest))
	if err != nil {
		return fmt.Errorf("error from Sherlock pinning environment '%s' to dev: %v", environment.Name(), err)
	}
	return nil
}

// WriteEnvironments will take a list of terra.Environment interfaces them and issue POST requests
// to write both the environment and any releases within that environment. 409 Conflict responses are ignored
func (c *Client) WriteEnvironments(envs []terra.Environment) ([]string, error) {
	createdEnvNames := make([]string, 0)
	for _, environment := range envs {
		log.Info().Msgf("exporting state for environment: %s", environment.Name())
		// When exporting state, we don't want Sherlock to try to be smart and interpolate
		// BEE chart releases. We'll create them manually based on our own gitops state
		// in the next step.
		newEnv := toModelCreatableEnvironment(environment, false)

		newEnvRequestParams := environments.NewPostAPIV2EnvironmentsParams().
			WithEnvironment(newEnv)
		_, createdEnv, err := c.client.Environments.PostAPIV2Environments(newEnvRequestParams)
		var envAlreadyExists bool
		if err != nil {
			// Don't error if creating the chart results in 409 conflict
			if _, ok := err.(*environments.PostAPIV2EnvironmentsConflict); !ok {
				return nil, fmt.Errorf("error creating cluster: %v", err)
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
func (c *Client) WriteClusters(cls []terra.Cluster) error {
	for _, cluster := range cls {
		log.Info().Msgf("exporting state for cluster: %s", cluster.Name())
		newCluster := toModelCreatableCluster(cluster)
		newClusterRequestParams := clusters.NewPostAPIV2ClustersParams().
			WithCluster(newCluster)
		_, _, err := c.client.Clusters.PostAPIV2Clusters(newClusterRequestParams)
		if err != nil {
			// Don't error if creating the chart results in 409 conflict
			if _, ok := err.(*clusters.PostAPIV2ClustersConflict); !ok {
				return fmt.Errorf("error creating cluster: %v", err)
			}
		}

		if err := c.writeReleases(cluster.Name(), cluster.Releases()); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) DeleteEnvironments(envs []terra.Environment) ([]string, error) {
	deletedEnvs := make([]string, 0)
	for _, env := range envs {
		// delete chart releases associated with environment
		releases := env.Releases()
		for _, release := range releases {
			if err := c.deleteRelease(release); err != nil {
				log.Warn().Msgf("error deleting chart release %s in environment %s: %v", release.Name(), env.Name(), err)
			}
		}
		params := environments.NewDeleteAPIV2EnvironmentsSelectorParams().
			WithSelector(env.Name())

		deletedEnv, err := c.client.Environments.DeleteAPIV2EnvironmentsSelector(params)
		if err != nil {
			return nil, fmt.Errorf("error deleting environment %s: %v", env.Name(), err)
		}
		log.Debug().Msgf("%#v", deletedEnv)
		deletedEnvs = append(deletedEnvs, deletedEnv.Payload.Name)
	}
	return deletedEnvs, nil
}

func (c *Client) EnableRelease(env terra.Environment, releaseName string) error {
	// need to pull info about the template env in order to set chart and app versions
	templateEnv, err := c.getEnvironment(env.Template())
	if err != nil {
		return fmt.Errorf("unable to fetch template %s: %v", env.Template(), err)
	}
	templateEnvName := templateEnv.Name
	// now look up the chart release to enable in the template
	templateRelease, err := c.getChartRelease(templateEnvName, releaseName)
	if err != nil {
		return fmt.Errorf("unable to enable release, error retrieving from template: %v", err)
	}

	// enable the chart-release in environment
	enabledChart := &models.V2controllersCreatableChartRelease{
		AppVersionExact:     templateRelease.AppVersionExact,
		Chart:               templateRelease.Chart,
		ChartVersionExact:   templateRelease.ChartVersionExact,
		Environment:         env.Name(),
		HelmfileRef:         templateRelease.HelmfileRef,
		Port:                templateRelease.Port,
		Protocol:            templateRelease.Protocol,
		Subdomain:           templateRelease.Subdomain,
		FirecloudDevelopRef: templateRelease.FirecloudDevelopRef,
	}
	log.Info().Msgf("enabling chart-release: %q in environment: %q", releaseName, env.Name())
	params := chart_releases.NewPostAPIV2ChartReleasesParams().WithChartRelease(enabledChart)
	_, _, err = c.client.ChartReleases.PostAPIV2ChartReleases(params)
	return err
}

func (c *Client) DisableRelease(envName, releaseName string) error {
	params := chart_releases.NewDeleteAPIV2ChartReleasesSelectorParams().WithSelector(strings.Join([]string{envName, releaseName}, "/"))
	_, err := c.client.ChartReleases.DeleteAPIV2ChartReleasesSelector(params)
	return err
}

func toModelCreatableEnvironment(env terra.Environment, chartReleasesFromTemplate bool) *models.V2controllersCreatableEnvironment {
	// if Helmfile ref isn't set it should default to head
	var helmfileRef string
	if env.TerraHelmfileRef() == "" {
		helmfileRef = "HEAD"
	} else {
		helmfileRef = env.TerraHelmfileRef()
	}
	return &models.V2controllersCreatableEnvironment{
		Base:                      env.Base(),
		BaseDomain:                utils.Nullable(env.BaseDomain()),
		DefaultCluster:            env.DefaultCluster().Name(),
		DefaultNamespace:          env.Namespace(),
		Lifecycle:                 utils.Nullable(env.Lifecycle().String()),
		Name:                      env.Name(),
		NamePrefixesDomain:        utils.Nullable(env.NamePrefixesDomain()),
		RequiresSuitability:       utils.Nullable(env.RequireSuitable()),
		TemplateEnvironment:       env.Template(),
		HelmfileRef:               utils.Nullable(helmfileRef),
		ChartReleasesFromTemplate: &chartReleasesFromTemplate,
		UniqueResourcePrefix:      env.UniqueResourcePrefix(),
	}
}

func toModelCreatableCluster(cluster terra.Cluster) *models.V2controllersCreatableCluster {
	// Hard coding to google for now since we don't have azure clusters
	provider := "google"
	// if Helmfile ref isn't set it should default to head
	var helmfileRef string
	if cluster.TerraHelmfileRef() == "" {
		helmfileRef = "HEAD"
	} else {
		helmfileRef = cluster.TerraHelmfileRef()
	}
	return &models.V2controllersCreatableCluster{
		Address:             cluster.Address(),
		Base:                cluster.Base(),
		Name:                cluster.Name(),
		Provider:            &provider,
		GoogleProject:       cluster.Project(),
		RequiresSuitability: utils.Nullable(cluster.RequireSuitable()),
		HelmfileRef:         &helmfileRef,
		Location:            utils.Nullable(cluster.Location()),
	}
}

func (c *Client) writeReleases(destinationName string, releases []terra.Release) error {
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

func (c *Client) writeAppRelease(environmentName string, release terra.AppRelease) error {
	log.Debug().Msgf("release name: %v", release.Name())
	modelChart := models.V2controllersCreatableChart{
		Name:            release.ChartName(),
		ChartRepo:       utils.Nullable(release.Repo()),
		DefaultPort:     utils.Nullable(int64(release.Port())),
		DefaultProtocol: utils.Nullable(release.Protocol()),
		// TODO don't default this figure out how thelma actually determines if legacy configs should be rendered
		LegacyConfigsEnabled: utils.Nullable(true),
	}
	// first try to create the chart
	newChartRequestParams := charts.NewPostAPIV2ChartsParams().
		WithChart(&modelChart)

	_, _, err := c.client.Charts.PostAPIV2Charts(newChartRequestParams)
	if err != nil {
		// Don't error if creating the chart results in 409 conflict
		if _, ok := err.(*charts.PostAPIV2ChartsConflict); !ok {
			return fmt.Errorf("error creating chart: %v", err)
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

	modelChartRelease := models.V2controllersCreatableChartRelease{
		AppVersionExact:     release.AppVersion(),
		Chart:               release.ChartName(),
		ChartVersionExact:   release.ChartVersion(),
		Cluster:             release.ClusterName(),
		Environment:         environmentName,
		HelmfileRef:         utils.Nullable(helmfileRef),
		Name:                releaseName,
		Namespace:           release.Namespace(),
		Port:                int64(release.Port()),
		Protocol:            release.Protocol(),
		Subdomain:           release.Subdomain(),
		FirecloudDevelopRef: release.FirecloudDevelopRef(),
	}

	newChartReleaseRequestParams := chart_releases.NewPostAPIV2ChartReleasesParams().
		WithChartRelease(&modelChartRelease)

	_, _, err = c.client.ChartReleases.PostAPIV2ChartReleases(newChartReleaseRequestParams)
	if err != nil {
		if _, ok := err.(*chart_releases.PostAPIV2ChartReleasesConflict); !ok {
			return fmt.Errorf("error creating chart release: %v", err)
		}
	}
	return nil
}

func (c *Client) writeClusterRelease(release terra.ClusterRelease) error {
	modelChart := models.V2controllersCreatableChart{
		Name:            release.ChartName(),
		ChartRepo:       utils.Nullable(release.Repo()),
		DefaultPort:     nil,
		DefaultProtocol: nil,
		// Cluster releases will never have legacy configs
		LegacyConfigsEnabled: utils.Nullable(false),
	}

	// first try to create the chart
	newChartRequestParams := charts.NewPostAPIV2ChartsParams().
		WithChart(&modelChart)

	_, _, err := c.client.Charts.PostAPIV2Charts(newChartRequestParams)
	if err != nil {
		// Don't error if creating the chart results in 409 conflict
		if _, ok := err.(*charts.PostAPIV2ChartsConflict); !ok {
			return fmt.Errorf("error creating chart: %v", err)
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
	modelChartRelease := models.V2controllersCreatableChartRelease{
		Chart:               release.ChartName(),
		ChartVersionExact:   release.ChartVersion(),
		Cluster:             release.ClusterName(),
		HelmfileRef:         utils.Nullable(helmfileRef),
		Name:                releaseName,
		Namespace:           release.Namespace(),
		FirecloudDevelopRef: release.FirecloudDevelopRef(),
	}

	newChartReleaseRequestParams := chart_releases.NewPostAPIV2ChartReleasesParams().
		WithChartRelease(&modelChartRelease)

	_, _, err = c.client.ChartReleases.PostAPIV2ChartReleases(newChartReleaseRequestParams)
	if err != nil {
		if _, ok := err.(*chart_releases.PostAPIV2ChartReleasesConflict); !ok {
			return fmt.Errorf("error creating chart release: %v", err)
		}
	}
	return nil
}

func (c *Client) deleteRelease(release terra.Release) error {
	params := chart_releases.NewDeleteAPIV2ChartReleasesSelectorParams().
		WithSelector(strings.Join([]string{release.ChartName(), release.Destination().Name()}, "-"))
	_, err := c.client.ChartReleases.DeleteAPIV2ChartReleasesSelector(params)
	return err
}

func (c *Client) getEnvironment(name string) (*Environment, error) {
	params := environments.NewGetAPIV2EnvironmentsSelectorParams().WithSelector(name)
	environment, err := c.client.Environments.GetAPIV2EnvironmentsSelector(params)
	if err != nil {
		return nil, err
	}

	return &Environment{environment.Payload}, nil
}

func (c *Client) getChartRelease(environmentName, releaseName string) (*Release, error) {
	params := chart_releases.NewGetAPIV2ChartReleasesSelectorParams().WithSelector(strings.Join([]string{environmentName, releaseName}, "/"))
	release, err := c.client.ChartReleases.GetAPIV2ChartReleasesSelector(params)
	if err != nil {
		return nil, err
	}
	return &Release{release.Payload}, nil
}
