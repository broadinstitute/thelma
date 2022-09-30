package statebucket

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func Test_GenerateUniqueName(t *testing.T) {
	name, err := generateUniqueEnvironmentName("myprefix", StateFile{})
	require.NoError(t, err)

	tokens := strings.Split(name, separator)
	assert.Len(t, tokens, 3)
	prefix, adjective, animal := tokens[0], tokens[1], tokens[2]

	assert.Equal(t, "myprefix", prefix)
	assert.Contains(t, adjectives, adjective)
	assert.Contains(t, animals, animal)
}
