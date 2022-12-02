package labels

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
