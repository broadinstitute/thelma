package resolver

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestResolvedChart_SourceDescription(t *testing.T) {
	fakeChartPath, err := utils.ExpandAndVerifyExists(path.Join("testdata", "charts", fakeChartName), "fake chart path")
	assert.NoError(t, err)

	testCases := []struct {
		name     string
		resType  ResolutionType
		expected string
	}{
		{
			name:     "local charts should be relative to cwd even when fully-qualified path is supplied",
			resType:  Local,
			expected: fmt.Sprintf("./testdata/charts/%s", fakeChartName),
		},
		{
			name:     "remote charts source should be repo name",
			resType:  Remote,
			expected: fakeChartRepo,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, err)

			rc := NewResolvedChart(
				fakeChartPath,
				fakeChartVersion,
				tc.resType,
				ChartRelease{
					Name:    fakeChartName,
					Repo:    fakeChartRepo,
					Version: fakeChartVersion,
				})

			assert.Equal(t, tc.expected, rc.SourceDescription())
		})
	}
}
