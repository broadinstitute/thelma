package tests

// this file needs to be in a separate package to avoid an import cycle

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// short hand for filter builders
var rf = filter.Releases()
var ef = filter.Environments()
var df = filter.Destinations()

func TestRequiredRoleIsSetCorrectly(t *testing.T) {
	//nolint:staticcheck // SA1019
	fixture, err := statefixtures.LoadFixture(statefixtures.Default)
	require.NoError(t, err)
	thelmaBuilder := builder.NewBuilder().WithTestDefaults(t).UseCustomStateLoader(fixture.Mocks().StateLoader)
	app, err := thelmaBuilder.Build()
	require.NoError(t, err)

	state, err := app.State()
	require.NoError(t, err)

	devCluster, err := state.Clusters().Get("terra-dev")
	require.NoError(t, err)

	devEnv, err := state.Environments().Get("dev")
	require.NoError(t, err)

	prodCluster, err := state.Clusters().Get("terra-prod")
	require.NoError(t, err)

	prodEnv, err := state.Environments().Get("prod")
	require.NoError(t, err)

	assert.Equal(t, "all-users", devCluster.RequiredRole())
	assert.Equal(t, "all-users", devEnv.RequiredRole())
	assert.Equal(t, "all-users-suspend-nonsuitable", prodCluster.RequiredRole())
	assert.Equal(t, "all-users-suspend-nonsuitable", prodEnv.RequiredRole())
}

// verify the default fixture has the data we expect
func TestDefaultFixtureHasExpectedInitialState(t *testing.T) {
	//nolint:staticcheck // SA1019
	fixture, err := statefixtures.LoadFixture(statefixtures.Default)
	require.NoError(t, err)
	thelmaBuilder := builder.NewBuilder().WithTestDefaults(t).UseCustomStateLoader(fixture.Mocks().StateLoader)
	app, err := thelmaBuilder.Build()
	require.NoError(t, err)

	state, err := app.State()
	require.NoError(t, err)

	// make sure we have the expected number of environments
	envs, err := state.Environments().All()
	require.NoError(t, err)
	assert.Equal(t, 17, len(envs))

	lives := ef.HasBase("live").Filter(envs)
	assert.Equal(t, len(lives), 5)

	bees := ef.HasBase("bee").Filter(envs)
	assert.Equal(t, 9, len(bees))

	personals := ef.HasBase("personal").Filter(envs)
	assert.Equal(t, 3, len(personals))

	// make sure we have the expected number of clusters
	clusters, err := state.Clusters().All()
	require.NoError(t, err)
	assert.Equal(t, 12, len(clusters))

	// make sure we have the expected number of releases
	releases, err := state.Releases().All()
	require.NoError(t, err)

	sams := rf.HasName("sam").Filter(releases)
	assert.Equal(t, 13, len(sams))

	liveSams := rf.And(
		rf.HasName("sam"),
		rf.DestinationMatches(df.HasBase("live")),
	).Filter(releases)
	assert.Equal(t, 5, len(liveSams))

	personalSams := rf.And(
		rf.HasName("sam"),
		rf.DestinationMatches(df.HasBase("personal")),
	).Filter(releases)
	assert.Equal(t, 0, len(personalSams))

	swatBeeSams := rf.And(
		rf.HasName("sam"),
		rf.DestinationMatches(
			df.IsEnvironmentMatching(ef.HasTemplateName("swatomation")),
		),
	).Filter(releases)
	assert.Equal(t, 4, len(swatBeeSams))

	samciBeeSams := rf.And(
		rf.HasName("sam"),
		rf.DestinationMatches(
			df.IsEnvironmentMatching(ef.HasTemplateName("sam-ci")),
		),
	).Filter(releases)
	assert.Equal(t, 2, len(samciBeeSams))

	templateSams := rf.And(
		rf.HasName("sam"),
		rf.DestinationMatches(
			df.IsEnvironmentMatching(
				ef.HasLifecycle(terra.Template),
			),
		),
	).Filter(releases)
	assert.Equal(t, 2, len(templateSams))

	rawlses := rf.HasName("rawls").Filter(releases)
	assert.Equal(t, 11, len(rawlses)) // 5 live, 1 template, 5 bees

	datarepos := rf.HasName("datarepo").Filter(releases)
	assert.Equal(t, 3, len(datarepos)) // 3 live (only in alpha, staging, prod)

	externalcredses := rf.HasName("externalcreds").Filter(releases)
	require.NoError(t, err)
	assert.Equal(t, 2, len(externalcredses)) // 2 live (only in dev and perf)
}

func TestDefaultFixtureHasCorrectVersions(t *testing.T) {
	//nolint:staticcheck // SA1019
	fixture, err := statefixtures.LoadFixture(statefixtures.Default)
	require.NoError(t, err)
	thelmaBuilder := builder.NewBuilder().WithTestDefaults(t).UseCustomStateLoader(fixture.Mocks().StateLoader)
	app, err := thelmaBuilder.Build()
	require.NoError(t, err)

	state, err := app.State()
	require.NoError(t, err)

	releases, err := state.Releases().All()
	require.NoError(t, err)

	// test we have correct app and chart version for sam in 4 types of envs
	devSam := rf.And(
		rf.HasName("sam"),
		rf.HasDestinationName("dev"),
	).Filter(releases)
	assert.Equal(t, "2d309b1645a0", devSam[0].AppVersion())
	assert.Equal(t, "0.34.0", devSam[0].ChartVersion())

	prodSam := rf.And(
		rf.HasName("sam"),
		rf.HasDestinationName("prod"),
	).Filter(releases)
	assert.Equal(t, "8f69c32bd9fe", prodSam[0].AppVersion())
	assert.Equal(t, "0.33.0", prodSam[0].ChartVersion())

	chipmunkSam := rf.And(
		rf.HasName("sam"),
		rf.HasDestinationName("fiab-funky-chipmunk"),
	).Filter(releases)
	assert.Equal(t, "2d309b1645a0", chipmunkSam[0].AppVersion())
	assert.Equal(t, "0.34.0", chipmunkSam[0].ChartVersion())

	walrusSam := rf.And(
		rf.HasName("sam"),
		rf.HasDestinationName("fiab-nerdy-walrus"),
	).Filter(releases)
	assert.Equal(t, "1.2.3", walrusSam[0].AppVersion())
	assert.Equal(t, "0.34.0", walrusSam[0].ChartVersion())

	snowflakeSam := rf.And(
		rf.HasName("sam"),
		rf.HasDestinationName("fiab-special-snowflake"),
	).Filter(releases)
	assert.Equal(t, 0, len(snowflakeSam), "sam is disabled in fiab-special-snowflake")

	snowflakeRawls := rf.And(
		rf.HasName("rawls"),
		rf.HasDestinationName("fiab-special-snowflake"),
	).Filter(releases)
	assert.Equal(t, "cead2f9206b5", snowflakeRawls[0].AppVersion())
	assert.Equal(t, "100.200.300", snowflakeRawls[0].ChartVersion())
	assert.Equal(t, "my-terra-helmfile-branch", snowflakeRawls[0].TerraHelmfileRef())

	paniniSam := rf.And(
		rf.HasName("sam"),
		rf.HasDestinationName("fiab-snarky-panini"),
	).Filter(releases)
	assert.Equal(t, "some-pr", paniniSam[0].TerraHelmfileRef())

	paniniRawls := rf.And(
		rf.HasName("rawls"),
		rf.HasDestinationName("fiab-snarky-panini"),
	).Filter(releases)
	assert.Equal(t, "completely-different-pr", paniniRawls[0].TerraHelmfileRef())

	// test urp is loaded correctly
	swirlyRabbit, err := state.Environments().Get("fiab-swirly-rabbit")
	require.NoError(t, err)
	assert.Equal(t, "e100", swirlyRabbit.UniqueResourcePrefix())
}

func TestUpdateState(t *testing.T) {
	t.Skip("stub state requires users to set up their own mocks")
	thelmaBuilder := builder.NewBuilder().WithTestDefaults(t)
	app, err := thelmaBuilder.Build()
	require.NoError(t, err)

	state, err := app.State()
	require.NoError(t, err)

	template, err := state.Environments().Get("sam-ci")
	require.NoError(t, err)

	missingEnv, err := state.Environments().Get("sam-ci-003")
	require.NoError(t, err)
	assert.Nil(t, missingEnv)

	_, err = state.Environments().CreateFromTemplate(template, terra.CreateOptions{
		Name: "sam-ci-003",
	})
	require.NoError(t, err)

	stateLoader, err := app.StateLoader()
	require.NoError(t, err)

	state, err = stateLoader.Reload() // reload state
	require.NoError(t, err)

	newEnv, err := state.Environments().Get("sam-ci-003")
	require.NoError(t, err)

	assert.Equal(t, "sam-ci-003", newEnv.Name())
	assert.Equal(t, 2, len(newEnv.Releases())) // opendj & sam
}

func Test_EnvironmentAttributes(t *testing.T) {
	//nolint:staticcheck // SA1019
	f, err := statefixtures.LoadFixture(statefixtures.Default)
	require.NoError(t, err)

	devEnv := f.Environment("dev")
	templateEnv := f.Environment("swatomation")
	beeEnv := f.Environment("fiab-funky-chipmunk")

	assert.Equal(t, 8, len(devEnv.Releases()))
	assert.Equal(t, terra.Static, devEnv.Lifecycle())
	assert.Equal(t, "", devEnv.UniqueResourcePrefix())

	assert.Equal(t, 6, len(templateEnv.Releases()))
	assert.Equal(t, terra.Template, templateEnv.Lifecycle())
	assert.Equal(t, "", templateEnv.UniqueResourcePrefix())

	assert.Equal(t, 6, len(beeEnv.Releases()))
	assert.Equal(t, terra.Dynamic, beeEnv.Lifecycle())
	assert.Equal(t, "e101", beeEnv.UniqueResourcePrefix())

	assert.Equal(t, "codemonkey42@broadinstitute.org", beeEnv.Owner())
}

func Test_ReleaseURLs(t *testing.T) {
	t.Skip("this functionality was not ported over")
	//nolint:staticcheck // SA1019
	f, err := statefixtures.LoadFixture(statefixtures.Default)
	require.NoError(t, err)

	devReleases := make(map[string]terra.AppRelease)
	for _, r := range f.Environment("dev").Releases() {
		devReleases[r.Name()] = r.(terra.AppRelease)
	}
	perfReleases := make(map[string]terra.AppRelease)
	for _, r := range f.Environment("perf").Releases() {
		perfReleases[r.Name()] = r.(terra.AppRelease)
	}
	swatReleases := make(map[string]terra.AppRelease)
	for _, r := range f.Environment("swatomation").Releases() {
		swatReleases[r.Name()] = r.(terra.AppRelease)
	}
	beeReleases := make(map[string]terra.AppRelease)
	for _, r := range f.Environment("fiab-funky-chipmunk").Releases() {
		beeReleases[r.Name()] = r.(terra.AppRelease)
	}

	t.Run("domain handled per environment via flag", func(t *testing.T) {
		assert.Equal(t, "leonardo", devReleases["leonardo"].Host()) // env defines no domain at all
		assert.Equal(t, "leonardo.dsde-perf.broadinstitute.org", perfReleases["leonardo"].Host())
		assert.Equal(t, "leonardo.swatomation.bee.envs-terra.bio", swatReleases["leonardo"].Host())
		assert.Equal(t, "leonardo.fiab-funky-chipmunk.bee.envs-terra.bio", beeReleases["leonardo"].Host())
	})
	t.Run("protocol can be overridden", func(t *testing.T) {
		assert.Equal(t, "ldap://opendj.fiab-funky-chipmunk.bee.envs-terra.bio", beeReleases["opendj"].URL())
		assert.Equal(t, "https://leonardo.fiab-funky-chipmunk.bee.envs-terra.bio", beeReleases["leonardo"].URL())
	})
	t.Run("port can be overridden", func(t *testing.T) {
		assert.Equal(t, 389, beeReleases["opendj"].Port())
		assert.Equal(t, 443, beeReleases["leonardo"].Port())
	})
	t.Run("subdomain can be overridden", func(t *testing.T) {
		assert.Equal(t, "https://workspace.fiab-funky-chipmunk.bee.envs-terra.bio", beeReleases["workspacemanager"].URL())
		assert.Equal(t, "https://leonardo.fiab-funky-chipmunk.bee.envs-terra.bio", beeReleases["leonardo"].URL())
	})
	t.Run("defaults still work outside bees", func(t *testing.T) {
		assert.Equal(t, 443, devReleases["leonardo"].Port())
		assert.Equal(t, "leonardo", devReleases["leonardo"].Subdomain())
		assert.Equal(t, "https", devReleases["leonardo"].Protocol())
	})
}
