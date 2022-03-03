//go:build smoke
// +build smoke

package statebucket

import (
	"github.com/broadinstitute/thelma/internal/thelma/utils/gcp/bucket"
	brequire "github.com/broadinstitute/thelma/internal/thelma/utils/gcp/bucket/testing/assert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
)

func TestStateBucket(t *testing.T) {
	b := bucket.NewTestBucket(t)
	sb := newWithBucket(b)

	// add empty statefile to the bucket
	err := sb.initialize()
	require.NoError(t, err)
	brequire.ObjectHasContent(t, b, stateObject, `{"environments":null}`)

	envs, err := sb.Environments()
	require.NoError(t, err)
	assert.Equal(t, 0, len(envs))

	require.NoError(t, sb.Add(DynamicEnvironment{
		Name:     "fake-env-1",
		Template: "fake-template",
	}))

	envs, err = sb.Environments()
	require.NoError(t, err)
	assert.Equal(t, 1, len(envs))

	require.NoError(t, sb.Add(DynamicEnvironment{
		Name:     "fake-env-2",
		Template: "fake-template",
	}))

	envs, err = sb.Environments()
	sort.Slice(envs, func(i, j int) bool {
		return envs[i].Name < envs[j].Name
	})

	require.NoError(t, err)
	assert.Equal(t, 2, len(envs))

	assert.NotNil(t, envs[0].VersionPins)
	assert.Equal(t, "fake-env-1", envs[0].Name)
	assert.Equal(t, "fake-template", envs[0].Template)
	assert.Equal(t, 0, len(envs[0].VersionPins))
	assert.False(t, envs[0].Hybrid)
	assert.Equal(t, "", envs[0].Fiab.Name)
	assert.Equal(t, "", envs[0].Fiab.IP)

	assert.NotNil(t, envs[1].VersionPins)
	assert.Equal(t, "fake-env-2", envs[1].Name)
	assert.Equal(t, "fake-template", envs[1].Template)
	assert.Equal(t, 0, len(envs[1].VersionPins))
	assert.False(t, envs[1].Hybrid)
	assert.Equal(t, "", envs[1].Fiab.Name)
	assert.Equal(t, "", envs[1].Fiab.IP)

	require.Error(t, sb.Add(DynamicEnvironment{
		Name:     "fake-env-2",
		Template: "fake-template",
	}), "duplicate env name should raise error")

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
	sort.Slice(envs, func(i, j int) bool {
		return envs[i].Name < envs[j].Name
	})

	require.NoError(t, err)
	assert.Equal(t, 3, len(envs))
	assert.Equal(t, "fake-env-3", envs[2].Name)
	assert.Equal(t, "other-fake-template", envs[2].Template)
	assert.Equal(t, 0, len(envs[2].VersionPins))
	assert.True(t, envs[2].Hybrid)
	assert.Equal(t, "fake-fiab-name", envs[2].Fiab.Name)
	assert.Equal(t, "10.11.12.13", envs[2].Fiab.IP)

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
}
