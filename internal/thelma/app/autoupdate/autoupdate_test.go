package autoupdate

import (
	bootstrapmocks "github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/bootstrap/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/installer"
	installermocks "github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/installer/mocks"
	spawnmocks "github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/spawn/mocks"

	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/utils/lazy"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

const testTag = "stable"

type AutoUpdateSuite struct {
	suite.Suite
	installer    *installermocks.Installer
	bootstrapper *bootstrapmocks.Bootstrapper
	spawn        *spawnmocks.Spawn
	autoupdate   AutoUpdate
}

func TestAutoUpdate(t *testing.T) {
	suite.Run(t, new(AutoUpdateSuite))
}

func (suite *AutoUpdateSuite) SetupTest() {
	t := suite.T()

	var cfg updateConfig
	thelmaConfig, err := config.NewTestConfig(t, map[string]interface{}{
		"autoupdate.tag": testTag,
	})
	require.NoError(t, err)
	require.NoError(t, thelmaConfig.Unmarshal(configKey, &cfg))

	suite.bootstrapper = bootstrapmocks.NewBootstrapper(t)
	suite.installer = installermocks.NewInstaller(t)
	suite.spawn = spawnmocks.NewSpawn(t)

	suite.autoupdate = &autoupdate{
		config: cfg,
		installer: lazy.NewLazyE(func() (installer.Installer, error) {
			return suite.installer, nil
		}),
		bootstrapper: suite.bootstrapper,
		spawn:        suite.spawn,
	}
}

func (suite *AutoUpdateSuite) Test_Update() {
	suite.installer.EXPECT().UpdateThelma(testTag).Return(nil)
	require.NoError(suite.T(), suite.autoupdate.Update())
}

func (suite *AutoUpdateSuite) Test_UpdateTo() {
	err := suite.autoupdate.UpdateTo("v4.5.6")
	require.Error(suite.T(), err)
	assert.ErrorContains(suite.T(), err, "auto-update is enabled")

	suite.disableAutoUpdate()
	suite.installer.EXPECT().UpdateThelma("v4.5.6").Return(nil)
	require.NoError(suite.T(), suite.autoupdate.UpdateTo("v4.5.6"))
}

func (suite *AutoUpdateSuite) Test_BackgroundUpdate_NotLaunchedIfNotNeeded() {
	suite.spawn.EXPECT().CurrentProcessIsSpawn().Return(false)
	suite.installer.EXPECT().ResolveVersions(testTag).Return(installer.ResolvedVersions{
		VersionAlias:   testTag,
		TargetVersion:  "v1.2.3",
		CurrentVersion: "v1.2.3",
	}, nil)

	require.NoError(suite.T(), suite.autoupdate.StartBackgroundUpdateIfEnabled())
}

func (suite *AutoUpdateSuite) Test_BackgroundUpdate_LaunchedIfNeeded() {
	suite.spawn.EXPECT().CurrentProcessIsSpawn().Return(false)
	suite.installer.EXPECT().ResolveVersions(testTag).Return(installer.ResolvedVersions{
		VersionAlias:   testTag,
		TargetVersion:  "v1.1.1",
		CurrentVersion: "v1.2.3",
	}, nil)
	suite.spawn.EXPECT().Spawn(updateCommandName).Return(nil)

	require.NoError(suite.T(), suite.autoupdate.StartBackgroundUpdateIfEnabled())
}

func (suite *AutoUpdateSuite) Test_Bootstrap() {
	suite.bootstrapper.EXPECT().Bootstrap().Return(nil)
	require.NoError(suite.T(), suite.autoupdate.Bootstrap())
}

func (suite *AutoUpdateSuite) disableAutoUpdate() {
	suite.autoupdate.(*autoupdate).config.Enabled = false
}
