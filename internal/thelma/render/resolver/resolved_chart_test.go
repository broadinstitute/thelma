package resolver

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestResolvedChart_SourceDescription(t *testing.T) {
	fakeChartPath, err := utils.ExpandAndVerifyExists(path.Join("testdata", "charts", fakeChart1Name), "fake chart path")
	assert.NoError(t, err)

	testCases := []struct {
		name          string
		resolvedChart ResolvedChart
		expected      string
	}{
		{
			name:          "local charts should be relative to cwd even when fully-qualified path is supplied",
			resolvedChart: NewLocallyResolvedChart(fakeChartPath, fakeChart1Version),
			expected:      fmt.Sprintf("./testdata/charts/%s", fakeChart1Name),
		},
		{
			name:          "remote charts source should be repo name",
			resolvedChart: NewRemotelyResolvedChart(fakeChartPath, fakeChart1Version, fakeChart1Repo),
			expected:      fakeChart1Repo,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.resolvedChart.SourceDescription())
		})
	}
}
