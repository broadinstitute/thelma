//go:build smoke
// +build smoke

package statebucket

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStateBucketOverrides(t *testing.T) {
	testCases := []struct {
		name  string
		newFn func(t *testing.T) (StateBucket, error)
	}{
		{
			name: "real gcs bucket",
			newFn: func(t *testing.T) (StateBucket, error) {
				b := bucket.NewTestBucket(t)
				tcfg, err := config.NewTestConfig(t)
				require.NoError(t, err)
				cfg, err := loadConfig(tcfg)
				return newWithBucket(b, cfg), nil
			},
		},
		{
			name: "fake bucket backed by filesystem",
			newFn: func(t *testing.T) (StateBucket, error) {
				return NewFake(t.TempDir())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// create some mock releases for use in tests
			leo := new(mocks.Release)
			leo.On("Name").Return("leonardo")

			sam := new(mocks.Release)
			sam.On("Name").Return("sam")

			// create statebucket
			sb, err := tc.newFn(t)
			require.NoError(t, err)

			// add empty statefile to the bucket
			err = sb.initialize()
			require.NoError(t, err)

			// make sure state is empty
			envs, err := sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 0, len(envs))

			// add fake-env-1 to state
			require.NoError(t, sb.Add(DynamicEnvironment{
				Name:     "fake-env-1",
				Template: "fake-template",
			}))

			// verify it was added
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 1, len(envs))
			assert.Empty(t, envs[0].Overrides, "should have no overrides")

			// setting overrides with empty set of releases should have no effect
			require.NoError(t, sb.EnableReleases("fake-env-1", []string{}))
			require.NoError(t, sb.DisableReleases("fake-env-1", []terra.Release{}), "empty list should have no effect")
			require.NoError(t, sb.OverrideVersions("fake-env-1", []terra.Release{}, func(release terra.Release, override terra.VersionOverride) {
				panic("This should never be called")
			}))

			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 1, len(envs))
			assert.Empty(t, envs[0].Overrides, "should have no overrides")

			// set a version override
			err = sb.OverrideVersions("fake-env-1", []terra.Release{sam}, func(release terra.Release, override terra.VersionOverride) {
				assert.Same(t, sam, release)
				override.SetAppVersion("100")
			})
			require.NoError(t, err)

			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 1, len(envs))
			assert.Equal(t, 1, len(envs[0].Overrides), "should have one override")

			assert.Equal(t, "100", envs[0].Overrides["sam"].AppVersion)
			assert.Equal(t, "", envs[0].Overrides["sam"].ChartVersion)
			assert.Equal(t, "", envs[0].Overrides["sam"].TerraHelmfileRef)
			assert.Equal(t, "", envs[0].Overrides["sam"].FirecloudDevelopRef)
			assert.False(t, envs[0].Overrides["sam"].HasEnableOverride())

			// set disable override on a different release
			err = sb.DisableReleases("fake-env-1", []terra.Release{leo})
			require.NoError(t, err)
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 1, len(envs))
			assert.Equal(t, 2, len(envs[0].Overrides), "should have two overrides")

			// make sure leo matches what we expect
			assert.Equal(t, "", envs[0].Overrides["leonardo"].AppVersion)
			assert.Equal(t, "", envs[0].Overrides["leonardo"].ChartVersion)
			assert.Equal(t, "", envs[0].Overrides["leonardo"].TerraHelmfileRef)
			assert.Equal(t, "", envs[0].Overrides["leonardo"].FirecloudDevelopRef)
			assert.True(t, envs[0].Overrides["leonardo"].HasEnableOverride())
			assert.False(t, envs[0].Overrides["leonardo"].IsEnabled())

			// make sure sam hasn't changed
			assert.Equal(t, "100", envs[0].Overrides["sam"].AppVersion)
			assert.Equal(t, "", envs[0].Overrides["sam"].ChartVersion)
			assert.Equal(t, "", envs[0].Overrides["sam"].TerraHelmfileRef)
			assert.Equal(t, "", envs[0].Overrides["sam"].FirecloudDevelopRef)
			assert.False(t, envs[0].Overrides["sam"].HasEnableOverride())

			// set another version override on sam
			err = sb.OverrideVersions("fake-env-1", []terra.Release{sam}, func(release terra.Release, override terra.VersionOverride) {
				assert.Same(t, sam, release)
				override.SetFirecloudDevelopRef("my-fc-branch")
			})
			require.NoError(t, err)
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 1, len(envs))
			assert.Equal(t, 2, len(envs[0].Overrides), "should have two overrides")

			// make sure sam matches what we expect
			assert.Equal(t, "100", envs[0].Overrides["sam"].AppVersion)
			assert.Equal(t, "", envs[0].Overrides["sam"].ChartVersion)
			assert.Equal(t, "", envs[0].Overrides["sam"].TerraHelmfileRef)
			assert.Equal(t, "my-fc-branch", envs[0].Overrides["sam"].FirecloudDevelopRef)
			assert.False(t, envs[0].Overrides["sam"].HasEnableOverride())

			// make sure leo hasn't changed
			assert.Equal(t, "", envs[0].Overrides["leonardo"].AppVersion)
			assert.Equal(t, "", envs[0].Overrides["leonardo"].ChartVersion)
			assert.Equal(t, "", envs[0].Overrides["leonardo"].TerraHelmfileRef)
			assert.Equal(t, "", envs[0].Overrides["leonardo"].FirecloudDevelopRef)
			assert.True(t, envs[0].Overrides["leonardo"].HasEnableOverride())
			assert.False(t, envs[0].Overrides["leonardo"].IsEnabled())

			// enable sam
			err = sb.EnableReleases("fake-env-1", []string{"sam"})
			require.NoError(t, err)
			envs, err = sb.Environments()
			assert.Equal(t, 1, len(envs))
			assert.Equal(t, 2, len(envs[0].Overrides), "should have two overrides")

			// make sure sam matches what we expect
			assert.Equal(t, "100", envs[0].Overrides["sam"].AppVersion)
			assert.Equal(t, "", envs[0].Overrides["sam"].ChartVersion)
			assert.Equal(t, "", envs[0].Overrides["sam"].TerraHelmfileRef)
			assert.Equal(t, "my-fc-branch", envs[0].Overrides["sam"].FirecloudDevelopRef)
			assert.True(t, envs[0].Overrides["sam"].HasEnableOverride())
			assert.True(t, envs[0].Overrides["sam"].IsEnabled())

			// make sure leo hasn't changed
			assert.Equal(t, "", envs[0].Overrides["leonardo"].AppVersion)
			assert.Equal(t, "", envs[0].Overrides["leonardo"].ChartVersion)
			assert.Equal(t, "", envs[0].Overrides["leonardo"].TerraHelmfileRef)
			assert.Equal(t, "", envs[0].Overrides["leonardo"].FirecloudDevelopRef)
			assert.True(t, envs[0].Overrides["leonardo"].HasEnableOverride())
			assert.False(t, envs[0].Overrides["leonardo"].IsEnabled())
		})
	}
}

func TestStateBucket(t *testing.T) {
	testCases := []struct {
		name  string
		newFn func(t *testing.T) (StateBucket, error)
	}{
		{
			name: "real gcs bucket",
			newFn: func(t *testing.T) (StateBucket, error) {
				b := bucket.NewTestBucket(t)
				tcfg, err := config.NewTestConfig(t)
				require.NoError(t, err)
				cfg, err := loadConfig(tcfg)
				return newWithBucket(b, cfg), nil
			},
		},
		{
			name: "fake bucket backed by filesystem",
			newFn: func(t *testing.T) (StateBucket, error) {
				return NewFake(t.TempDir())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// create some mock releases for use in tests
			leo := new(mocks.Release)
			leo.On("Name").Return("leonardo")

			sam := new(mocks.Release)
			sam.On("Name").Return("sam")

			sb, err := tc.newFn(t)
			require.NoError(t, err)

			// add empty statefile to the bucket
			err = sb.initialize()
			require.NoError(t, err)

			// make sure state is empty
			envs, err := sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 0, len(envs))

			// add fake-env-1 to state
			require.NoError(t, sb.Add(DynamicEnvironment{
				Name:     "fake-env-1",
				Template: "fake-template",
			}))

			// verify it was added
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 1, len(envs))

			// add fake-env-2 to state
			require.NoError(t, sb.Add(DynamicEnvironment{
				Name:     "fake-env-2",
				Template: "fake-template",
				Overrides: map[string]*Override{
					"sam": {
						AppVersion: "1.2.3",
					},
				},
			}))

			// verify it was added
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 2, len(envs))

			// verify fake-env-1 attributes
			assert.NotNil(t, envs[0].Overrides)
			assert.Equal(t, "fake-env-1", envs[0].Name)
			assert.Equal(t, "fake-template", envs[0].Template)
			assert.Empty(t, envs[0].Overrides)
			assert.False(t, envs[0].Hybrid)
			assert.Equal(t, "", envs[0].Fiab.Name)
			assert.Equal(t, "", envs[0].Fiab.IP)

			// verify fake-env-2 attributes
			assert.NotNil(t, envs[1].Overrides)
			assert.Equal(t, "fake-env-2", envs[1].Name)
			assert.Equal(t, "fake-template", envs[1].Template)
			assert.Equal(t, 1, len(envs[1].Overrides))
			assert.Equal(t, "1.2.3", envs[1].Overrides["sam"].AppVersion)
			assert.False(t, envs[1].Hybrid)
			assert.Equal(t, "", envs[1].Fiab.Name)
			assert.Equal(t, "", envs[1].Fiab.IP)

			// make sure dupe env name raises error
			require.Error(t, sb.Add(DynamicEnvironment{
				Name:     "fake-env-2",
				Template: "fake-template",
			}), "duplicate env name should raise error")

			// add fake-env-3
			require.NoError(t, sb.Add(DynamicEnvironment{
				Name:     "fake-env-3",
				Template: "other-fake-template",
				Hybrid:   true,
				Fiab: Fiab{
					Name: "fake-fiab-name",
					IP:   "10.11.12.13",
				},
			}))

			envs, err = sb.Environments()

			// verify fake-env-3 has expected attributes
			require.NoError(t, err)
			assert.Equal(t, 3, len(envs))
			assert.Equal(t, "fake-env-3", envs[2].Name)
			assert.Equal(t, "other-fake-template", envs[2].Template)
			assert.Empty(t, envs[2].Overrides)
			assert.True(t, envs[2].Hybrid)
			assert.Equal(t, "fake-fiab-name", envs[2].Fiab.Name)
			assert.Equal(t, "10.11.12.13", envs[2].Fiab.IP)

			// test overrides on fake-env-1 match what was in bucket
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 1, len(envs[1].Overrides))
			assert.Equal(t, "1.2.3", envs[1].Overrides["sam"].AppVersion)
			assert.Equal(t, "", envs[1].Overrides["sam"].ChartVersion)
			assert.Equal(t, "", envs[1].Overrides["sam"].TerraHelmfileRef)
			assert.Equal(t, "", envs[1].Overrides["sam"].FirecloudDevelopRef)
			assert.False(t, envs[1].Overrides["sam"].HasEnableOverride())

			// test environment deletion
			require.NoError(t, sb.Delete("fake-env-1"))
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 2, len(envs))
			assert.Equal(t, envs[0].Name, "fake-env-2")
			assert.Equal(t, envs[1].Name, "fake-env-3")

			require.NoError(t, sb.Delete("fake-env-3"))
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 1, len(envs))
			assert.Equal(t, envs[0].Name, "fake-env-2")

			require.NoError(t, sb.Delete("fake-env-2"))
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 0, len(envs))
		})
	}
}
