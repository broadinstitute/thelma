package releaser

import (
	publishmocks "github.com/broadinstitute/thelma/internal/thelma/charts/publish/mocks"
	indexmocks "github.com/broadinstitute/thelma/internal/thelma/charts/repo/index/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	sourcemocks "github.com/broadinstitute/thelma/internal/thelma/charts/source/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	sherlockmocks "github.com/broadinstitute/thelma/internal/thelma/clients/sherlock/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ChartReleaser(t *testing.T) {
	// construct mocks/dependencies
	chartsDir := sourcemocks.NewChartsDir(t)

	index := indexmocks.NewIndex(t)

	dir := t.TempDir()

	publisher := publishmocks.NewPublisher(t)
	publisher.EXPECT().Index().Return(index)
	publisher.EXPECT().ChartDir().Return(dir)

	description := "my commit message"
	sherlockUpdater := sherlockmocks.NewChartVersionUpdater(t)
	deployedVersionUpdater := &DeployedVersionUpdater{
		SherlockUpdaters: []sherlock.ChartVersionUpdater{sherlockUpdater},
	}

	releaser := NewChartReleaser(chartsDir, publisher, deployedVersionUpdater)

	// set additional mocks mocks
	chartsDir.EXPECT().Exists("mysql").Return(true)
	chartsDir.EXPECT().Exists("foundation").Return(true)
	chartsDir.EXPECT().Exists("yale").Return(true)

	mysql := sourcemocks.NewChart(t)
	mysql.EXPECT().Name().Return("mysql")
	chartsDir.EXPECT().GetChart("mysql").Return(mysql, nil)
	index.EXPECT().MostRecentVersion("mysql").Return("1.2.3")
	mysql.EXPECT().BumpChartVersion("1.2.3").Return("1.3.0", nil)
	chartsDir.EXPECT().UpdateDependentVersionConstraints(mysql, "*").Return(nil)
	mysql.EXPECT().GenerateDocs().Return(nil)
	mysql.EXPECT().PackageChart(dir).Return(nil)
	sherlockUpdater.EXPECT().ReportNewChartVersion("mysql", "1.3.0", "1.2.3", description).Return(nil)

	yale := sourcemocks.NewChart(t)
	yale.EXPECT().Name().Return("yale")
	chartsDir.EXPECT().GetChart("yale").Return(yale, nil)
	index.EXPECT().MostRecentVersion("yale").Return("0.23.4")
	yale.EXPECT().BumpChartVersion("0.23.4").Return("0.24.0", nil)
	chartsDir.EXPECT().UpdateDependentVersionConstraints(yale, "*").Return(nil)
	yale.EXPECT().GenerateDocs().Return(nil)
	yale.EXPECT().PackageChart(dir).Return(nil)
	sherlockUpdater.EXPECT().ReportNewChartVersion("yale", "0.24.0", "0.23.4", description).Return(nil)

	foundation := sourcemocks.NewChart(t)
	foundation.EXPECT().Name().Return("foundation")
	chartsDir.EXPECT().GetChart("foundation").Return(foundation, nil)
	index.EXPECT().MostRecentVersion("foundation").Return("1.30.5")
	foundation.EXPECT().BumpChartVersion("1.30.5").Return("1.31.0", nil)
	chartsDir.EXPECT().UpdateDependentVersionConstraints(foundation, "*").Return(nil)
	foundation.EXPECT().GenerateDocs().Return(nil)
	foundation.EXPECT().PackageChart(dir).Return(nil)
	sherlockUpdater.EXPECT().ReportNewChartVersion("foundation", "1.31.0", "1.30.5", description).Return(nil)

	agora := sourcemocks.NewChart(t)
	agora.EXPECT().Name().Return("agora")
	chartsDir.EXPECT().GetChart("agora").Return(agora, nil)
	index.EXPECT().MostRecentVersion("agora").Return("11.12.13")
	agora.EXPECT().BumpChartVersion("11.12.13").Return("11.13.0", nil)
	chartsDir.EXPECT().UpdateDependentVersionConstraints(agora, "*").Return(nil)
	agora.EXPECT().GenerateDocs().Return(nil)
	agora.EXPECT().PackageChart(dir).Return(nil)
	sherlockUpdater.EXPECT().ReportNewChartVersion("agora", "11.13.0", "11.12.13", description).Return(nil)

	sam := sourcemocks.NewChart(t)
	sam.EXPECT().Name().Return("sam")
	chartsDir.EXPECT().GetChart("sam").Return(sam, nil)
	index.EXPECT().MostRecentVersion("sam").Return("1.0.0")
	sam.EXPECT().BumpChartVersion("1.0.0").Return("1.1.0", nil)
	chartsDir.EXPECT().UpdateDependentVersionConstraints(sam, "*").Return(nil)
	sam.EXPECT().GenerateDocs().Return(nil)
	sam.EXPECT().PackageChart(dir).Return(nil)
	sherlockUpdater.EXPECT().ReportNewChartVersion("sam", "1.1.0", "1.0.0", description).Return(nil)

	bpm := sourcemocks.NewChart(t)
	bpm.EXPECT().Name().Return("bpm")
	chartsDir.EXPECT().GetChart("bpm").Return(bpm, nil)
	index.EXPECT().MostRecentVersion("bpm").Return("2.13.0")
	bpm.EXPECT().BumpChartVersion("2.13.0").Return("2.14.0", nil)
	chartsDir.EXPECT().UpdateDependentVersionConstraints(bpm, "*").Return(nil)
	bpm.EXPECT().GenerateDocs().Return(nil)
	bpm.EXPECT().PackageChart(dir).Return(nil)
	sherlockUpdater.EXPECT().ReportNewChartVersion("bpm", "2.14.0", "2.13.0", description).Return(nil)

	chartsDir.EXPECT().GetCharts("mysql", "foundation", "yale").Return([]source.Chart{
		mysql,
		foundation,
		yale,
	}, nil)

	chartsDir.EXPECT().WithTransitiveDependents([]source.Chart{mysql, foundation, yale}).Return([]source.Chart{
		mysql,
		foundation,
		yale,
		agora,
		sam,
		bpm,
	}, nil)

	chartsDir.EXPECT().GetCharts("mysql", "foundation", "yale", "agora", "sam", "bpm").Return([]source.Chart{
		mysql,
		foundation,
		yale,
		agora,
		sam,
		bpm,
	}, nil)

	chartsDir.EXPECT().RecursivelyUpdateDependencies(mysql, foundation, yale, agora, sam, bpm).Return(nil)

	publisher.EXPECT().Publish().Return(6, nil)

	versionMap, err := releaser.Release([]string{"mysql", "foundation", "yale"}, "my commit message")
	require.NoError(t, err)

	assert.Equal(t, map[string]VersionPair{
		"agora":      {PriorVersion: "11.12.13", NewVersion: "11.13.0"},
		"bpm":        {PriorVersion: "2.13.0", NewVersion: "2.14.0"},
		"foundation": {PriorVersion: "1.30.5", NewVersion: "1.31.0"},
		"mysql":      {PriorVersion: "1.2.3", NewVersion: "1.3.0"},
		"sam":        {PriorVersion: "1.0.0", NewVersion: "1.1.0"},
		"yale":       {PriorVersion: "0.23.4", NewVersion: "0.24.0"},
	}, versionMap)
}
