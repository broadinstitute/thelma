package stateval

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statefixtures"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_BuildStateValues(t *testing.T) {
	state := statefixtures.LoadFixture(statefixtures.Default, t)

	chartPath := t.TempDir()

	testCases := []struct {
		name              string
		release           terra.Release
		expectedAppValues AppValues
		expectedArgoApp   ArgoApp
	}{
		{
			name:    "static env release",
			release: state.Release("sam", "dev"),
			expectedAppValues: AppValues{
				ChartPath: chartPath,
				Release: Release{
					Name:       "sam",
					Type:       "app",
					Namespace:  "terra-dev",
					AppVersion: "2d309b1645a0",
				},
				Destination: Destination{
					Name:       "dev",
					Type:       "environment",
					ConfigBase: "live",
					ConfigName: "dev",
				},
				Environment: Environment{
					Name:     "dev",
					IsHybrid: false,
				},
			},
			expectedArgoApp: ArgoApp{
				ProjectName:    "terra-dev",
				ClusterName:    "terra-dev",
				ClusterAddress: "https://35.238.186.116",
			},
		},
		{
			name:    "template env release",
			release: state.Release("sam", "swatomation"),
			expectedAppValues: AppValues{
				ChartPath: chartPath,
				Release: Release{
					Name:       "sam",
					Type:       "app",
					Namespace:  "terra-swatomation",
					AppVersion: "2d309b1645a0",
				},
				Destination: Destination{
					Name:       "swatomation",
					Type:       "environment",
					ConfigBase: "bee",
					ConfigName: "swatomation",
				},
				Environment: Environment{
					Name:     "swatomation",
					IsHybrid: false,
				},
			},
			expectedArgoApp: ArgoApp{
				ProjectName:    "terra-swatomation",
				ClusterName:    "terra-qa",
				ClusterAddress: "https://35.224.175.229",
			},
		},
		{
			name:    "dynamic env release",
			release: state.Release("sam", "fiab-funky-chipmunk"),
			expectedAppValues: AppValues{
				ChartPath: chartPath,
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
			expectedArgoApp: ArgoApp{
				ProjectName:         "terra-fiab-funky-chipmunk",
				ClusterName:         "terra-qa",
				ClusterAddress:      "https://35.224.175.229",
				TerraHelmfileRef:    "my-th-branch-1",
				FirecloudDevelopRef: "my-fc-branch-1",
			},
		},
		{
			name:    "cluster release",
			release: state.Release("diskmanager", "terra-dev"),
			expectedAppValues: AppValues{
				ChartPath: chartPath,
				Release: Release{
					Name:      "diskmanager",
					Type:      "cluster",
					Namespace: "default",
				},
				Destination: Destination{
					Name:       "terra-dev",
					Type:       "cluster",
					ConfigBase: "terra",
					ConfigName: "terra-dev",
				},
				Cluster: Cluster{
					Name: "terra-dev",
				},
			},
			expectedArgoApp: ArgoApp{
				ProjectName:    "cluster-terra-dev",
				ClusterName:    "terra-dev",
				ClusterAddress: "https://35.238.186.116",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedAppValues, BuildAppValues(tc.release, chartPath))

			var expectedArgoApp ArgoAppValues
			expectedArgoApp.ArgoApp = tc.expectedArgoApp
			// copy common settings over from app values
			expectedArgoApp.Release = tc.expectedAppValues.Release
			expectedArgoApp.Destination = tc.expectedAppValues.Destination
			expectedArgoApp.Environment = tc.expectedAppValues.Environment
			expectedArgoApp.Cluster = tc.expectedAppValues.Cluster

			assert.Equal(t, expectedArgoApp, BuildArgoAppValues(tc.release))

			var expectedArgoProject ArgoProjectValues
			// copy project name over from expected argo app
			expectedArgoProject.ArgoProject = ArgoProject{
				ProjectName: expectedArgoApp.ArgoApp.ProjectName,
			}
			// copy common settings over from app values
			expectedArgoProject.Destination = tc.expectedAppValues.Destination

			assert.Equal(t, expectedArgoProject, BuildArgoProjectValues(tc.release.Destination()))
		})
	}
}
