package statebucket

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_GenerateUnique(t *testing.T) {
	urp, err := generateUniqueResourcePrefix("fiab-choover-funky-sapien", StateFile{})
	require.NoError(t, err)
	assert.Equal(t, "aac1", urp)
}

func Test_GenerateUnique_NoReuse(t *testing.T) {
	urp, err := generateUniqueResourcePrefix("fiab-choover-funky-sapien", StateFile{
		Environments: map[string]DynamicEnvironment{
			"e1": {
				Name:                 "e1",
				UniqueResourcePrefix: "aac1",
			},
			"e2": {
				Name:                 "e2",
				UniqueResourcePrefix: "cac1",
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "dac1", urp, "Should skip letter `b` and proceed to `d`")
}
