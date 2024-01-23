package bee

// Basic imports
import (
	cleanupmocks "github.com/broadinstitute/thelma/internal/thelma/bee/cleanup/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/bee/seed"
	seedmocks "github.com/broadinstitute/thelma/internal/thelma/bee/seed/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/ops/artifacts"
	"github.com/broadinstitute/thelma/internal/thelma/ops/logs"
	logsmocks "github.com/broadinstitute/thelma/internal/thelma/ops/logs/mocks"
	opsmocks "github.com/broadinstitute/thelma/internal/thelma/ops/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/ops/status"
	syncmocks "github.com/broadinstitute/thelma/internal/thelma/ops/sync/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	argocd_names "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/argocd"
	argomocks "github.com/broadinstitute/thelma/internal/thelma/toolbox/argocd/mocks"
	kubectlmocks "github.com/broadinstitute/thelma/internal/thelma/toolbox/kubectl/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

const beeName = "my-bee"

type BeesTestSuite struct {
	suite.Suite
	statefixture statefixtures.Fixture

	env      terra.Environment
	releases struct {
		leo terra.Release
		sam terra.Release
		wsm terra.Release
	}

	mocks struct {
		argocd  *argomocks.ArgoCD
		seeder  *seedmocks.Seeder
		cleanup *cleanupmocks.Cleanup
		kubectl *kubectlmocks.Kubectl
		sync    *syncmocks.Sync
		logs    *logsmocks.Logs
	}

	bees Bees
}

func (suite *BeesTestSuite) SetupSubTest() {
	statefixture, err := statefixtures.LoadFixtureFromFile("testdata/statefixture.yaml")
	require.NoError(suite.T(), err)
	suite.statefixture = statefixture

	suite.env = statefixture.Environment(beeName)
	suite.releases.leo = statefixture.Release("leonardo", beeName)
	suite.releases.sam = statefixture.Release("sam", beeName)
	suite.releases.wsm = statefixture.Release("workspacemanager", beeName)

	suite.mocks.argocd = argomocks.NewArgoCD(suite.T())
	suite.mocks.seeder = seedmocks.NewSeeder(suite.T())
	suite.mocks.cleanup = cleanupmocks.NewCleanup(suite.T())
	suite.mocks.kubectl = kubectlmocks.NewKubectl(suite.T())

	ops := opsmocks.NewOps(suite.T())
	suite.mocks.sync = syncmocks.NewSync(suite.T())
	suite.mocks.logs = logsmocks.NewLogs(suite.T())
	ops.EXPECT().Sync().Return(suite.mocks.sync, nil).Maybe()
	ops.EXPECT().Logs().Return(suite.mocks.logs).Maybe()

	bees, err := NewBees(
		suite.mocks.argocd,
		statefixture.Mocks().StateLoader,
		suite.mocks.seeder,
		suite.mocks.cleanup,
		suite.mocks.kubectl,
		ops,
		nil,
	)
	require.NoError(suite.T(), err)
	suite.bees = bees
}

func (suite *BeesTestSuite) TestProvisionWith() {
	testCases := []struct {
		name      string
		opts      ProvisionOptions
		setup     func(*BeesTestSuite, ProvisionOptions)
		expectErr string
	}{
		{
			name: "basic successful test case",
			opts: provisionOptions(),
			setup: func(s *BeesTestSuite, opts ProvisionOptions) {
				s.expectPinReleaseVersionsEmptyOverrides()
				s.expectProvisionBeeNamespaceAndGenerator()
				s.expectSyncArgoAppsForReleases(opts.WaitHealthy, opts.WaitHealthTimeoutSeconds)
				s.expectSeed(opts.SeedOptions)
			},
		},
		{
			name: "helmfile ref override",
			opts: provisionOptions(func(options *ProvisionOptions) {
				options.PinOptions.Flags.TerraHelmfileRef = "my-helmfile-ref"
			}),
			setup: func(s *BeesTestSuite, opts ProvisionOptions) {
				s.expectPinEnvHelmfileRef("my-helmfile-ref")
				s.expectPinReleaseVersions(map[string]terra.VersionOverride{
					"leonardo": {
						TerraHelmfileRef: "my-helmfile-ref",
					},
					"sam": {
						TerraHelmfileRef: "my-helmfile-ref",
					},
					"workspacemanager": {
						TerraHelmfileRef: "my-helmfile-ref",
					},
				})
				s.expectProvisionBeeNamespaceAndGenerator()
				s.expectSyncArgoAppsForReleases(opts.WaitHealthy, opts.WaitHealthTimeoutSeconds)
				s.expectSeed(opts.SeedOptions)
			},
		},
		{
			name: "sam version override",
			opts: provisionOptions(func(options *ProvisionOptions) {
				options.PinOptions.FileOverrides = map[string]terra.VersionOverride{
					"sam": {
						ChartVersion:     "my-chart-version",
						AppVersion:       "my-app-version",
						TerraHelmfileRef: "my-helmfile-ref",
					},
				}
			}),
			setup: func(s *BeesTestSuite, opts ProvisionOptions) {
				s.expectPinEnvHelmfileRef("my-helmfile-ref")
				s.expectPinReleaseVersions(map[string]terra.VersionOverride{
					"sam": {
						ChartVersion:     "my-chart-version",
						AppVersion:       "my-app-version",
						TerraHelmfileRef: "my-helmfile-ref",
					},
				})
				s.expectProvisionBeeNamespaceAndGenerator()
				s.expectSyncArgoAppsForReleases(opts.WaitHealthy, opts.WaitHealthTimeoutSeconds)
				s.expectSeed(opts.SeedOptions)
			},
		},
		{
			name: "failed pin does not export logs",
			opts: provisionOptions(func(options *ProvisionOptions) {
				options.PinOptions.Flags.TerraHelmfileRef = "my-helmfile-ref"
			}),
			setup: func(s *BeesTestSuite, opts ProvisionOptions) {
				s.expectPinEnvHelmfileRefReturnError("my-helmfile-ref", errors.New("pin totally failed"))
			},
			expectErr: "pin totally failed",
		},
		{
			name: "failed sync does export logs",
			opts: provisionOptions(),
			setup: func(s *BeesTestSuite, opts ProvisionOptions) {
				s.expectPinReleaseVersionsEmptyOverrides()
				s.expectProvisionBeeNamespaceAndGenerator()
				s.expectSyncArgoAppsForReleasesReturnSamFailure(opts.WaitHealthy, opts.WaitHealthTimeoutSeconds, errors.New("sync totally failed"))

				s.expectExportLogs()

			},
			expectErr: "sync totally failed",
		},
		{
			name: "failed log export should still report original error",
			opts: provisionOptions(),
			setup: func(s *BeesTestSuite, opts ProvisionOptions) {
				s.expectPinReleaseVersionsEmptyOverrides()
				s.expectProvisionBeeNamespaceAndGenerator()
				s.expectSyncArgoAppsForReleasesReturnSamFailure(opts.WaitHealthy, opts.WaitHealthTimeoutSeconds, errors.New("sync totally failed"))
				s.expectExportLogsReturnError(errors.New("log export boo-boo"))
			},
			expectErr: "sync totally failed",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.setup != nil {
				tc.setup(suite, tc.opts)
			}
			bee, err := suite.bees.ProvisionWith(beeName, tc.opts)

			if tc.expectErr != "" {
				require.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.expectErr)
				return
			}

			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), beeName, bee.Environment.Name())
		})
	}
}

func (suite *BeesTestSuite) expectPinEnvHelmfileRef(helmfileRef string) {
	suite.statefixture.Mocks().Environments.EXPECT().PinEnvironmentToTerraHelmfileRef(beeName, helmfileRef).Return(nil)
}

func (suite *BeesTestSuite) expectPinEnvHelmfileRefReturnError(helmfileRef string, err error) {
	suite.statefixture.Mocks().Environments.EXPECT().PinEnvironmentToTerraHelmfileRef(beeName, helmfileRef).Return(err)
}

func (suite *BeesTestSuite) expectPinReleaseVersionsEmptyOverrides() {
	suite.expectPinReleaseVersions(make(map[string]terra.VersionOverride))
}

func (suite *BeesTestSuite) expectPinReleaseVersions(overrides map[string]terra.VersionOverride) {
	for _, r := range suite.getReleases() {
		// add an empty override for each release that doesn't already have one
		if _, exists := overrides[r.Name()]; !exists {
			overrides[r.Name()] = terra.VersionOverride{}
		}
	}

	suite.statefixture.Mocks().Environments.EXPECT().PinVersions(beeName, overrides).Return(overrides, nil)
}

func (suite *BeesTestSuite) expectProvisionBeeNamespaceAndGenerator() {
	// create namespace
	suite.mocks.kubectl.EXPECT().CreateNamespace(suite.env).Return(nil)

	// sync terra-bee-generator
	suite.mocks.argocd.EXPECT().HardRefresh(generatorArgoApp).Return(nil)

	myBeeGenerator := argocd_names.GeneratorName(suite.env)

	// wait for terra-<name>-generator to exist
	suite.mocks.argocd.EXPECT().WaitExist(myBeeGenerator).Return(nil)

	// sync terra-<name>-generator
	suite.mocks.argocd.EXPECT().SyncApp(myBeeGenerator).Return(argocd.SyncResult{Synced: true}, nil)
}

func (suite *BeesTestSuite) expectSyncArgoAppsForReleases(expectWaitHealthy bool, expectWaitHealthyTimeoutSeconds int) {
	statuses := make(map[terra.Release]*status.Status)

	releases := suite.getReleases()

	for _, r := range releases {
		statuses[r] = &status.Status{
			Health: argocd.Healthy,
			Sync:   argocd.Synced,
		}
	}

	suite.mocks.sync.EXPECT().Sync(mock.Anything, len(releases), mock.Anything).Run(func(rs []terra.Release, _ int, options ...argocd.SyncOption) {
		assert.ElementsMatch(suite.T(), releases, rs)

		var opts argocd.SyncOptions
		for _, option := range options {
			option(&opts)
		}
		assert.Equal(suite.T(), true, opts.SkipLegacyConfigsRestart)
		assert.Equal(suite.T(), true, opts.SyncIfNoDiff)
		assert.Equal(suite.T(), expectWaitHealthy, opts.WaitHealthy)
		assert.Equal(suite.T(), expectWaitHealthyTimeoutSeconds, opts.WaitHealthyTimeoutSeconds)
	}).Return(statuses, nil)
}

func (suite *BeesTestSuite) expectSyncArgoAppsForReleasesReturnSamFailure(expectWaitHealthy bool, expectWaitHealthyTimeoutSeconds int, samError error) {
	statuses := make(map[terra.Release]*status.Status)

	releases := suite.getReleases()

	for _, r := range releases {
		statuses[r] = &status.Status{
			Health: argocd.Healthy,
			Sync:   argocd.Synced,
		}
	}

	statuses[suite.releases.sam] = &status.Status{
		Health: argocd.Degraded,
		Sync:   argocd.Synced,
	}

	suite.mocks.sync.EXPECT().Sync(mock.Anything, len(releases), mock.Anything).Run(func(rs []terra.Release, _ int, options ...argocd.SyncOption) {
		assert.ElementsMatch(suite.T(), releases, rs)

		var opts argocd.SyncOptions
		for _, option := range options {
			option(&opts)
		}
		assert.Equal(suite.T(), true, opts.SkipLegacyConfigsRestart)
		assert.Equal(suite.T(), true, opts.SyncIfNoDiff)
		assert.Equal(suite.T(), expectWaitHealthy, opts.WaitHealthy)
		assert.Equal(suite.T(), expectWaitHealthyTimeoutSeconds, opts.WaitHealthyTimeoutSeconds)
	}).Return(statuses, errors.Errorf("sync failed for %s: %v", suite.releases.sam.FullName(), samError))
}

func (suite *BeesTestSuite) expectSeed(opts seed.SeedOptions) {
	suite.mocks.seeder.EXPECT().Seed(suite.env, opts).Return(nil)
}

func (suite *BeesTestSuite) expectExportLogs() {
	suite.expectExportLogsReturnError(nil)
}

func (suite *BeesTestSuite) expectExportLogsReturnError(exportErr error) {
	releases := suite.getReleases()

	locations := make(map[terra.Release]artifacts.Location)
	for _, r := range releases {
		locations[r] = artifacts.Location{
			CloudConsoleURL: "https://ignored",
		}
	}

	suite.mocks.logs.EXPECT().Export(mock.Anything, mock.Anything).Run(func(rs []terra.Release, opts ...logs.ExportOption) {
		// make sure expected releases are passed
		assert.ElementsMatch(suite.T(), releases, rs)

		// make sure expected log options are passed
		var options logs.ExportOptions
		for _, optfn := range opts {
			optfn(&options)
		}

		assert.Equal(suite.T(), logs.ExportOptions{
			Artifacts: artifacts.Options{
				Upload: true, // harccoded in bees.go
			},
		}, options)
	}).Return(locations, exportErr)
}

func (suite *BeesTestSuite) getReleases() []terra.Release {
	return []terra.Release{
		suite.releases.leo,
		suite.releases.sam,
		suite.releases.wsm,
	}
}

// mimic the default provision options supplied by Thelma CLI
func provisionOptions(overrideFn ...func(options *ProvisionOptions)) ProvisionOptions {
	opts := ProvisionOptions{
		Seed: true,
		SeedOptions: seed.SeedOptions{
			Step1CreateElasticsearch: true,
			Step2RegisterSaProfiles:  true,
			Step3AddSaSamPermissions: true,
			Step4RegisterTestUsers:   true,
			Step5CreateAgora:         true,
			Step6ExtraUser:           nil,
			RegisterSelfShortcut:     false,
		},
		ExportLogsOnFailure: true,
		ProvisionExistingOptions: ProvisionExistingOptions{
			WaitHealthy:              true,
			WaitHealthTimeoutSeconds: 1800,
			Notify:                   false,
		},
	}

	for _, fn := range overrideFn {
		fn(&opts)
	}

	return opts
}

func TestBeesTestSuite(t *testing.T) {
	suite.Run(t, new(BeesTestSuite))
}
