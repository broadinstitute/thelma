package tests

// this file needs to be in a separate package to avoid an import cycle

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statefixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// short hand for filter builders
var rf = filter.Releases()
var ef = filter.Environments()
var df = filter.Destinations()

func TestSuitabilityIsLoaded(t *testing.T) {
	t.Fatalf("Makek sure RequireSuitable is set properly on both clusters and environments")
}

// verify the default fixture has the data we expect
func TestDefaultFixtureHasExpectedInitialState(t *testing.T) {
	thelmaBuilder := builder.NewBuilder().WithTestDefaults(t)
	app, err := thelmaBuilder.Build()
	require.NoError(t, err)

	state, err := app.State()
	require.NoError(t, err)

	// make sure we have the expected number of environments
	envs, err := state.Environments().All()
	require.NoError(t, err)
	assert.Equal(t, 17, len(envs))

	lives, err := state.Environments().Filter(ef.HasBase("live"))
	require.NoError(t, err)
	assert.Equal(t, len(lives), 5)

	bees, err := state.Environments().Filter(ef.HasBase("bee"))
	require.NoError(t, err)
	assert.Equal(t, 9, len(bees))

	personals, err := state.Environments().Filter(ef.HasBase("personal"))
	require.NoError(t, err)
	assert.Equal(t, 3, len(personals))

	// make sure we have the expected number of clusters
	clusters, err := state.Clusters().All()
	require.NoError(t, err)
	assert.Equal(t, 12, len(clusters))

	// make sure we have the expected number of releases
	sams, err := state.Releases().Filter(rf.HasName("sam"))
	require.NoError(t, err)
	assert.Equal(t, 13, len(sams))

	liveSams, err := state.Releases().Filter(rf.And(
		rf.HasName("sam"),
		rf.DestinationMatches(df.HasBase("live")),
	))
	require.NoError(t, err)
	assert.Equal(t, 5, len(liveSams))

	personalSams, err := state.Releases().Filter(rf.And(
		rf.HasName("sam"),
		rf.DestinationMatches(df.HasBase("personal")),
	))
	require.NoError(t, err)
	assert.Equal(t, 0, len(personalSams))

	swatBeeSams, err := state.Releases().Filter(rf.And(
		rf.HasName("sam"),
		rf.DestinationMatches(
			df.IsEnvironmentMatching(ef.HasTemplateName("swatomation")),
		),
	))
	require.NoError(t, err)
	assert.Equal(t, 4, len(swatBeeSams))

	samciBeeSams, err := state.Releases().Filter(rf.And(
		rf.HasName("sam"),
		rf.DestinationMatches(
			df.IsEnvironmentMatching(ef.HasTemplateName("sam-ci")),
		),
	))
	require.NoError(t, err)
	assert.Equal(t, 2, len(samciBeeSams))

	templateSams, err := state.Releases().Filter(rf.And(
		rf.HasName("sam"),
		rf.DestinationMatches(
			df.IsEnvironmentMatching(
				ef.HasLifecycle(terra.Template),
			),
		),
	))
	require.NoError(t, err)
	assert.Equal(t, 2, len(templateSams))

	rawlses, err := state.Releases().Filter(rf.HasName("rawls"))
	require.NoError(t, err)
	assert.Equal(t, 11, len(rawlses)) // 5 live, 1 template, 5 bees

	datarepos, err := state.Releases().Filter(rf.HasName("datarepo"))
	require.NoError(t, err)
	assert.Equal(t, 3, len(datarepos)) // 3 live (only in alpha, staging, prod)

	externalcredses, err := state.Releases().Filter(rf.HasName("externalcreds"))
	require.NoError(t, err)
	assert.Equal(t, 2, len(externalcredses)) // 2 live (only in dev and perf)
}

func TestDefaultFixtureHasCorrectVersions(t *testing.T) {
	thelmaBuilder := builder.NewBuilder().WithTestDefaults(t)
	app, err := thelmaBuilder.Build()
	require.NoError(t, err)

	state, err := app.State()
	require.NoError(t, err)

	// test we have correct app and chart version for sam in 4 types of envs
	devSam, err := state.Releases().Filter(rf.And(
		rf.HasName("sam"),
		rf.HasDestinationName("dev"),
	))
	require.NoError(t, err)
	assert.Equal(t, "2d309b1645a0", devSam[0].(terra.AppRelease).AppVersion())
	assert.Equal(t, "0.34.0", devSam[0].ChartVersion())

	prodSam, err := state.Releases().Filter(rf.And(
		rf.HasName("sam"),
		rf.HasDestinationName("prod"),
	))
	require.NoError(t, err)
	assert.Equal(t, "8f69c32bd9fe", prodSam[0].(terra.AppRelease).AppVersion())
	assert.Equal(t, "0.33.0", prodSam[0].ChartVersion())

	chipmunkSam, err := state.Releases().Filter(rf.And(
		rf.HasName("sam"),
		rf.HasDestinationName("fiab-funky-chipmunk"),
	))
	require.NoError(t, err)
	assert.Equal(t, "2d309b1645a0", chipmunkSam[0].(terra.AppRelease).AppVersion())
	assert.Equal(t, "0.34.0", chipmunkSam[0].ChartVersion())

	walrusSam, err := state.Releases().Filter(rf.And(
		rf.HasName("sam"),
		rf.HasDestinationName("fiab-nerdy-walrus"),
	))
	require.NoError(t, err)
	assert.Equal(t, "1.2.3", walrusSam[0].(terra.AppRelease).AppVersion())
	assert.Equal(t, "0.34.0", walrusSam[0].ChartVersion())

	snowflakeSam, err := state.Releases().Filter(rf.And(
		rf.HasName("sam"),
		rf.HasDestinationName("fiab-special-snowflake"),
	))
	require.NoError(t, err)
	assert.Equal(t, 0, len(snowflakeSam), "sam is disabled in fiab-special-snowflake")

	snowflakeRawls, err := state.Releases().Filter(rf.And(
		rf.HasName("rawls"),
		rf.HasDestinationName("fiab-special-snowflake"),
	))
	require.NoError(t, err)
	assert.Equal(t, "cead2f9206b5", snowflakeRawls[0].(terra.AppRelease).AppVersion())
	assert.Equal(t, "100.200.300", snowflakeRawls[0].ChartVersion())
	assert.Equal(t, "my-terra-helmfile-branch", snowflakeRawls[0].TerraHelmfileRef())
	assert.Equal(t, "", snowflakeRawls[0].FirecloudDevelopRef())

	paniniSam, err := state.Releases().Filter(rf.And(
		rf.HasName("sam"),
		rf.HasDestinationName("fiab-snarky-panini"),
	))
	require.NoError(t, err)
	assert.Equal(t, "some-pr", paniniSam[0].TerraHelmfileRef())

	paniniRawls, err := state.Releases().Filter(rf.And(
		rf.HasName("rawls"),
		rf.HasDestinationName("fiab-snarky-panini"),
	))
	require.NoError(t, err)
	assert.Equal(t, "completely-different-pr", paniniRawls[0].TerraHelmfileRef())
}

func TestUpdateState(t *testing.T) {
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

	err = state.Environments().CreateFromTemplate("sam-ci-003", template)
	require.NoError(t, err)

	state, err = app.State() // reload state
	require.NoError(t, err)

	newEnv, err := state.Environments().Get("sam-ci-003")
	require.NoError(t, err)

	assert.Equal(t, "sam-ci-003", newEnv.Name())
	assert.Equal(t, 2, len(newEnv.Releases())) // opendj & sam
}

func Test_EnvironmentAttributes(t *testing.T) {
	f := statefixtures.LoadFixture(statefixtures.Default, t)

	devEnv := f.Environment("dev")
	swatEnv := f.Environment("swatomation")
	hybridEnv := f.Environment("fiab-funky-chipmunk")

	assert.Equal(t, 8, len(devEnv.Releases()))
	assert.Equal(t, terra.Static, devEnv.Lifecycle())
	assert.False(t, devEnv.IsHybrid())

	assert.Equal(t, 5, len(swatEnv.Releases()))
	assert.Equal(t, terra.Template, swatEnv.Lifecycle())
	assert.False(t, swatEnv.IsHybrid())

	assert.Equal(t, 5, len(hybridEnv.Releases()))
	assert.Equal(t, terra.Dynamic, hybridEnv.Lifecycle())
	assert.True(t, hybridEnv.IsHybrid())
}
