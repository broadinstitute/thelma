package labels

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ForReleaseDestination(t *testing.T) {
	devEnv := mocks.NewEnvironment(t)
	devEnv.EXPECT().Name().Return("dev")
	devEnv.EXPECT().IsEnvironment().Return(true)
	devEnv.EXPECT().IsCluster().Return(false)

	devCluster := mocks.NewCluster(t)
	devCluster.EXPECT().Name().Return("terra-dev")
	devCluster.EXPECT().IsEnvironment().Return(false)
	devCluster.EXPECT().IsCluster().Return(true)

	leoDev := mocks.NewAppRelease(t)
	leoDev.EXPECT().Name().Return("leonardo")
	leoDev.EXPECT().Destination().Return(devEnv)
	leoDev.EXPECT().Cluster().Return(devCluster)

	yaleDev := mocks.NewClusterRelease(t)
	yaleDev.EXPECT().Name().Return("yale")
	yaleDev.EXPECT().Destination().Return(devCluster)
	yaleDev.EXPECT().Cluster().Return(devCluster)

	t.Run("for destination: dev env", func(t *testing.T) {
		assert.Equal(t, map[string]string{
			"env":     "dev",
			"cluster": "",
		}, ForDestination(devEnv))
	})

	t.Run("for destination: dev cluster", func(t *testing.T) {
		assert.Equal(t, map[string]string{
			"env":     "",
			"cluster": "terra-dev",
		}, ForDestination(devCluster))
	})

	t.Run("for release: leo dev", func(t *testing.T) {
		assert.Equal(t, map[string]string{
			"release": "leonardo",
			"env":     "dev",
			"cluster": "terra-dev",
		}, ForRelease(leoDev))
	})

	t.Run("for release: yale dev cluster", func(t *testing.T) {
		assert.Equal(t, map[string]string{
			"release": "yale",
			"env":     "",
			"cluster": "terra-dev",
		}, ForRelease(yaleDev))
	})

	t.Run("for release or destination: dev env", func(t *testing.T) {
		assert.Equal(t, map[string]string{
			"release": "",
			"env":     "dev",
			"cluster": "",
		}, ForReleaseOrDestination(devEnv))
	})

	t.Run("for release or destination: dev cluster", func(t *testing.T) {
		assert.Equal(t, map[string]string{
			"release": "",
			"env":     "",
			"cluster": "terra-dev",
		}, ForReleaseOrDestination(devCluster))
	})

	t.Run("for release or destination: leonardo dev", func(t *testing.T) {
		assert.Equal(t, map[string]string{
			"release": "leonardo",
			"env":     "dev",
			"cluster": "terra-dev",
		}, ForReleaseOrDestination(leoDev))
	})

	t.Run("for release or destination: yale terra-dev", func(t *testing.T) {
		assert.Equal(t, map[string]string{
			"release": "yale",
			"env":     "",
			"cluster": "terra-dev",
		}, ForReleaseOrDestination(yaleDev))
	})
}

func Test_Normalized(t *testing.T) {
	assert.Equal(t, map[string]string{}, Normalize(map[string]string{}))
	assert.Equal(t, map[string]string{"a": "b"}, Normalize(map[string]string{"a": "b"}))
	assert.Equal(t,
		map[string]string{
			"a":    "b",
			"_job": "foo",
		},
		Normalize(
			map[string]string{
				"a":   "b",
				"job": "foo",
			},
		),
	)
}

func Test_Merge(t *testing.T) {
	assert.Equal(t, map[string]string{}, Merge())
	assert.Equal(t, map[string]string{"a": "b"}, Merge(map[string]string{"a": "b"}))
	assert.Equal(t, map[string]string{"a": "c"}, Merge(map[string]string{"a": "b"}, map[string]string{"a": "c"}))
	assert.Equal(t, map[string]string{"a": "b"}, Merge(map[string]string{"a": "c"}, map[string]string{"a": "b"}))
	assert.Equal(t, map[string]string{"a": "d"}, Merge(map[string]string{"a": "b"}, map[string]string{"a": "c"}, map[string]string{"a": "d"}))

	assert.Equal(t, map[string]string{"a": "d"}, Merge(nil, map[string]string{"a": "c"}, map[string]string{"a": "d"}))
	assert.Equal(t, map[string]string{"a": "c"}, Merge(nil, map[string]string{"a": "c"}))
}
