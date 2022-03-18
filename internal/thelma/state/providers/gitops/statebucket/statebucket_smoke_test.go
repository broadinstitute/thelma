//go:build smoke
// +build smoke

package statebucket

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStateBucket(t *testing.T) {
	testCases := []struct {
		name  string
		newFn func(t *testing.T) (StateBucket, error)
	}{
		{
			name: "real gcs bucket",
			newFn: func(t *testing.T) (StateBucket, error) {
				b := bucket.NewTestBucket(t)
				return newWithBucket(b), nil
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
				VersionPins: map[string]string{
					"foo": "1.2.3",
				},
			}))

			// verify it was added
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 2, len(envs))

			// verify fake-env-1 attributes
			assert.NotNil(t, envs[0].VersionPins)
			assert.Equal(t, "fake-env-1", envs[0].Name)
			assert.Equal(t, "fake-template", envs[0].Template)
			assert.Equal(t, 0, len(envs[0].VersionPins))
			assert.False(t, envs[0].Hybrid)
			assert.Equal(t, "", envs[0].Fiab.Name)
			assert.Equal(t, "", envs[0].Fiab.IP)

			// verify fake-env-2 attributes
			assert.NotNil(t, envs[1].VersionPins)
			assert.Equal(t, "fake-env-2", envs[1].Name)
			assert.Equal(t, "fake-template", envs[1].Template)
			assert.Equal(t, 1, len(envs[1].VersionPins))
			assert.Equal(t, "1.2.3", envs[1].VersionPins["foo"])
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
			assert.Equal(t, 0, len(envs[2].VersionPins))
			assert.True(t, envs[2].Hybrid)
			assert.Equal(t, "fake-fiab-name", envs[2].Fiab.Name)
			assert.Equal(t, "10.11.12.13", envs[2].Fiab.IP)

			// test version pinning and unpinning on fake-env-1
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 1, len(envs[1].VersionPins))
			assert.Equal(t, "1.2.3", envs[1].VersionPins["foo"])

			// pin empty map should have no effect
			require.NoError(t, sb.PinVersions("fake-env-2", map[string]string{}), "empty map should have no effect")
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 1, len(envs[1].VersionPins))
			assert.Equal(t, "1.2.3", envs[1].VersionPins["foo"])

			// test pins merge on top of each other
			require.NoError(t, sb.PinVersions("fake-env-2", map[string]string{"bar": "100", "baz": "4.5.6"}))
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 3, len(envs[1].VersionPins))
			assert.Equal(t, "1.2.3", envs[1].VersionPins["foo"])
			assert.Equal(t, "100", envs[1].VersionPins["bar"])
			assert.Equal(t, "4.5.6", envs[1].VersionPins["baz"])

			require.NoError(t, sb.PinVersions("fake-env-2", map[string]string{"foo": "1234"}))
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 3, len(envs[1].VersionPins))
			assert.Equal(t, "1234", envs[1].VersionPins["foo"])
			assert.Equal(t, "100", envs[1].VersionPins["bar"])
			assert.Equal(t, "4.5.6", envs[1].VersionPins["baz"])

			// test unpin removes all pins
			require.NoError(t, sb.UnpinVersions("fake-env-2"))
			envs, err = sb.Environments()
			require.NoError(t, err)
			assert.Equal(t, 0, len(envs[1].VersionPins))

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
