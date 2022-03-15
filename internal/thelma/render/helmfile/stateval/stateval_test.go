package stateval

import (
	"github.com/broadinstitute/thelma/internal/thelma/gitops/statefixtures"
	"github.com/broadinstitute/thelma/internal/thelma/terra"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_BuildArgoAppValues(t *testing.T) {
	state := statefixtures.LoadFixture(statefixtures.Default, t)

	testCases := []struct {
		name     string
		release  terra.Release
		expected ArgoAppValues
	}{
		{
			name:    "dynamic env release",
			release: state.Release("sam", "fiab-funky-chipmunk"),
			expected: ArgoAppValues{
				Release: Release{
					Name:       "sam",
					Type:       "app",
					Namespace:  "terra-fiab-funky-chipmunk",
					AppVersion: "2d309b1645a0",
				},
				Destination: Destination{
					Name:       "fiab-funky-chipmunk",
					Type:       "environment",
					ConfigBase: "bee",
					ConfigName: "swatomation",
				},
				ArgoApp: ArgoApp{
					ProjectName:    "terra-fiab-funky-chipmunk",
					ClusterName:    "terra-qa",
					ClusterAddress: "https://35.224.175.229",
				},
				Environment: Environment{
					Name:     "fiab-funky-chipmunk",
					IsHybrid: true,
					Fiab: struct {
						Name string `yaml:"Name,omitempty"`
						IP   string `yaml:"IP,omitempty"`
					}{
						Name: "fiab-automation-funky-chipmunk",
						IP:   "10.0.0.2",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, BuildArgoAppValues(tc.release))
		})
	}
}
