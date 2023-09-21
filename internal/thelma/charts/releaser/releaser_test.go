package releaser

import (
	publishmocks "github.com/broadinstitute/thelma/internal/thelma/charts/publish/mocks"
	indexmocks "github.com/broadinstitute/thelma/internal/thelma/charts/repo/index/mocks"

	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	sourcemocks "github.com/broadinstitute/thelma/internal/thelma/charts/source/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

func Test_ChartReleaser(t *testing.T) {
	chartsDir := sourcemocks.NewChartsDir(t)
	index := indexmocks.NewIndex(t)
	dir := t.TempDir()
	publisher := publishmocks.NewPublisher(t)
	publisher.EXPECT().Index().Return(index)
	publisher.EXPECT().ChartDir().Return(dir)
	updater := &DeployedVersionUpdater{}
	releaser := NewChartReleaser(chartsDir, publisher, updater)

	chartsDir.EXPECT().Exists("mysql").Return(true)
	chartsDir.EXPECT().Exists("foundation").Return(true)
	chartsDir.EXPECT().Exists("yale").Return(true)

	mysql := sourcemocks.NewChart(t)
	mysql.EXPECT().Name().Return("mysql")
	mysql.EXPECT().Path().Return("charts/mysql")
	chartsDir.EXPECT().GetChart("mysql").Return(mysql, nil)
	index.EXPECT().MostRecentVersion("mysql").Return("1.2.3")
	mysql.EXPECT().BumpChartVersion("1.2.3").Return("1.3.0", nil)
	chartsDir.EXPECT().UpdateDependentVersionConstraints(mysql, "1.3.0").Return(nil)
	mysql.EXPECT().GenerateDocs().Return(nil)
	mysql.EXPECT().PackageChart(dir).Return(nil)

	yale := sourcemocks.NewChart(t)
	yale.EXPECT().Name().Return("yale")
	yale.EXPECT().Path().Return("charts/yale")
	chartsDir.EXPECT().GetChart("yale").Return(yale, nil)
	index.EXPECT().MostRecentVersion("yale").Return("0.23.4")
	yale.EXPECT().BumpChartVersion("0.23.4").Return("0.24.0", nil)
	chartsDir.EXPECT().UpdateDependentVersionConstraints(yale, "0.24.0").Return(nil)
	yale.EXPECT().GenerateDocs().Return(nil)
	yale.EXPECT().PackageChart(dir).Return(nil)

	foundation := sourcemocks.NewChart(t)
	foundation.EXPECT().Name().Return("foundation")
	foundation.EXPECT().Path().Return("charts/foundation")
	chartsDir.EXPECT().GetChart("foundation").Return(foundation, nil)
	index.EXPECT().MostRecentVersion("foundation").Return("1.30.5")
	foundation.EXPECT().BumpChartVersion("1.30.5").Return("1.31.0", nil)
	chartsDir.EXPECT().UpdateDependentVersionConstraints(foundation, "1.31.0").Return(nil)
	foundation.EXPECT().GenerateDocs().Return(nil)
	foundation.EXPECT().PackageChart(dir).Return(nil)

	agora := sourcemocks.NewChart(t)
	agora.EXPECT().Name().Return("agora")
	agora.EXPECT().Path().Return("charts/agora")
	chartsDir.EXPECT().GetChart("agora").Return(agora, nil)
	index.EXPECT().MostRecentVersion("agora").Return("11.12.13")
	agora.EXPECT().BumpChartVersion("11.12.13").Return("11.13.0", nil)
	chartsDir.EXPECT().UpdateDependentVersionConstraints(agora, "11.13.0").Return(nil)
	agora.EXPECT().GenerateDocs().Return(nil)
	agora.EXPECT().PackageChart(dir).Return(nil)

	sam := sourcemocks.NewChart(t)
	sam.EXPECT().Name().Return("sam")
	sam.EXPECT().Path().Return("charts/sam")
	chartsDir.EXPECT().GetChart("sam").Return(sam, nil)
	index.EXPECT().MostRecentVersion("sam").Return("1.0.0")
	sam.EXPECT().BumpChartVersion("1.0.0").Return("1.1.0", nil)
	chartsDir.EXPECT().UpdateDependentVersionConstraints(sam, "1.1.0").Return(nil)
	sam.EXPECT().GenerateDocs().Return(nil)
	sam.EXPECT().PackageChart(dir).Return(nil)

	bpm := sourcemocks.NewChart(t)
	bpm.EXPECT().Name().Return("bpm")
	bpm.EXPECT().Path().Return("charts/bpm")
	chartsDir.EXPECT().GetChart("bpm").Return(bpm, nil)
	index.EXPECT().MostRecentVersion("bpm").Return("2.13.0")
	bpm.EXPECT().BumpChartVersion("2.13.0").Return("2.14.0", nil)
	chartsDir.EXPECT().UpdateDependentVersionConstraints(bpm, "2.14.0").Return(nil)
	bpm.EXPECT().GenerateDocs().Return(nil)
	bpm.EXPECT().PackageChart(dir).Return(nil)

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

	versionMap, err := releaser.Release([]string{"mysql", "foundation", "yale"}, "my change description")
	require.NoError(t, err)

	assert.Equal(t, map[string]string{
		"agora":      "11.13.0",
		"bpm":        "2.14.0",
		"foundation": "1.31.0",
		"mysql":      "1.3.0",
		"sam":        "1.1.0",
		"yale":       "0.24.0",
	}, versionMap)
}
