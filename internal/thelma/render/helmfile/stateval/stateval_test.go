package stateval

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_BuildStateValues(t *testing.T) {
	//nolint:staticcheck // SA1019
	fixture, err := statefixtures.LoadFixture(statefixtures.Default)
	require.NoError(t, err)

	chartPath := t.TempDir()

	testCases := []struct {
		name                string
		release             terra.Release
		expectedAppValues   AppValues
		expectedArgoApp     ArgoApp
		expectedArgoProject ArgoProject
	}{
		{
			name:    "static env release",
			release: fixture.Release("sam", "dev"),
			expectedAppValues: AppValues{
				ChartPath: chartPath,
				Release: Release{
					Name:       "sam",
					ChartName:  "sam",
					Type:       "app",
					Namespace:  "terra-dev",
					AppVersion: "2d309b1645a0",
				},
				Destination: Destination{
					Name:         "dev",
					Type:         "environment",
					ConfigBase:   "live",
					ConfigName:   "dev",
					RequiredRole: "all-users",
				},
				Environment: Environment{
					Name:                 "dev",
					UniqueResourcePrefix: "",
					EnableJanitor:        false,
				},
				Cluster: Cluster{
					Name:                "terra-dev",
					GoogleProject:       "broad-dsde-dev",
					GoogleProjectSuffix: "dev",
				},
			},
			expectedArgoApp: ArgoApp{
				ProjectName:    "terra-dev",
				ClusterName:    "terra-dev",
				ClusterAddress: "https://35.238.186.116",
			},
			expectedArgoProject: ArgoProject{
				ProjectName: "terra-dev",
				Generator: Generator{
					Name:             "terra-dev-generator",
					TerraHelmfileRef: "",
				},
			},
		},
		{
			name:    "suitable release",
			release: fixture.Release("sam", "prod"),
			expectedAppValues: AppValues{
				ChartPath: chartPath,
				Release: Release{
					Name:       "sam",
					ChartName:  "sam",
					Type:       "app",
					Namespace:  "terra-prod",
					AppVersion: "8f69c32bd9fe",
				},
				Destination: Destination{
					Name:         "prod",
					Type:         "environment",
					ConfigBase:   "live",
					ConfigName:   "prod",
					RequiredRole: "all-users-suspend-nonsuitable",
				},
				Environment: Environment{
					Name:                 "prod",
					UniqueResourcePrefix: "",
					EnableJanitor:        false,
				},
				Cluster: Cluster{
					Name:                "terra-prod",
					GoogleProject:       "broad-dsde-prod",
					GoogleProjectSuffix: "prod",
				},
			},
			expectedArgoApp: ArgoApp{
				ProjectName:    "terra-prod",
				ClusterName:    "terra-prod",
				ClusterAddress: "https://35.232.149.177",
			},
			expectedArgoProject: ArgoProject{
				ProjectName: "terra-prod",
				Generator: Generator{
					Name:             "terra-prod-generator",
					TerraHelmfileRef: "",
				},
			},
		},
		{
			name:    "template env release",
			release: fixture.Release("sam", "swatomation"),
			expectedAppValues: AppValues{
				ChartPath: chartPath,
				Release: Release{
					Name:       "sam",
					ChartName:  "sam",
					Type:       "app",
					Namespace:  "terra-swatomation",
					AppVersion: "2d309b1645a0",
				},
				Destination: Destination{
					Name:         "swatomation",
					Type:         "environment",
					ConfigBase:   "bee",
					ConfigName:   "swatomation",
					RequiredRole: "all-users",
				},
				Environment: Environment{
					Name:                 "swatomation",
					UniqueResourcePrefix: "",
					EnableJanitor:        true,
				},
				Cluster: Cluster{
					Name:                "terra-qa",
					GoogleProject:       "broad-dsde-qa",
					GoogleProjectSuffix: "qa",
				},
			},
			expectedArgoApp: ArgoApp{
				ProjectName:    "terra-swatomation",
				ClusterName:    "terra-qa",
				ClusterAddress: "https://35.224.175.229",
			},
			expectedArgoProject: ArgoProject{
				ProjectName: "terra-swatomation",
				Generator: Generator{
					Name:             "terra-swatomation-generator",
					TerraHelmfileRef: "",
				},
			},
		},
		{
			name:    "dynamic env release",
			release: fixture.Release("sam", "fiab-funky-chipmunk"),
			expectedAppValues: AppValues{
				ChartPath: chartPath,
				Release: Release{
					Name:       "sam",
					ChartName:  "sam",
					Type:       "app",
					Namespace:  "terra-fiab-funky-chipmunk",
					AppVersion: "2d309b1645a0",
				},
				Destination: Destination{
					Name:         "fiab-funky-chipmunk",
					Type:         "environment",
					ConfigBase:   "bee",
					ConfigName:   "swatomation",
					RequiredRole: "all-users",
				},
				Environment: Environment{
					Name:                 "fiab-funky-chipmunk",
					UniqueResourcePrefix: "e101",
					EnableJanitor:        true,
				},
				Cluster: Cluster{
					Name:                "terra-qa",
					GoogleProject:       "broad-dsde-qa",
					GoogleProjectSuffix: "qa",
				},
			},
			expectedArgoApp: ArgoApp{
				ProjectName:      "terra-fiab-funky-chipmunk",
				ClusterName:      "terra-qa",
				ClusterAddress:   "https://35.224.175.229",
				TerraHelmfileRef: "my-th-branch-1",
			},
			expectedArgoProject: ArgoProject{
				ProjectName: "terra-fiab-funky-chipmunk",
				Generator: Generator{
					Name:             "terra-fiab-funky-chipmunk-generator",
					TerraHelmfileRef: "",
				},
			},
		},
		{
			name:    "cluster release",
			release: fixture.Release("diskmanager", "terra-dev"),
			expectedAppValues: AppValues{
				ChartPath: chartPath,
				Release: Release{
					Name:      "diskmanager",
					ChartName: "diskmanager",
					Type:      "cluster",
					Namespace: "default",
				},
				Destination: Destination{
					Name:         "terra-dev",
					Type:         "cluster",
					ConfigBase:   "terra",
					ConfigName:   "terra-dev",
					RequiredRole: "all-users",
				},
				Cluster: Cluster{
					Name:                "terra-dev",
					GoogleProject:       "broad-dsde-dev",
					GoogleProjectSuffix: "dev",
				},
			},
			expectedArgoApp: ArgoApp{
				ProjectName:    "cluster-terra-dev",
				ClusterName:    "terra-dev",
				ClusterAddress: "https://35.238.186.116",
			},
			expectedArgoProject: ArgoProject{
				ProjectName: "cluster-terra-dev",
				Generator: Generator{
					Name:             "cluster-terra-dev-generator",
					TerraHelmfileRef: "",
				},
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
			expectedArgoProject.ArgoProject = tc.expectedArgoProject
			// copy common settings over from app values
			expectedArgoProject.Destination = tc.expectedAppValues.Destination

			assert.Equal(t, expectedArgoProject, BuildArgoProjectValues(tc.release.Destination()))
		})
	}
}
