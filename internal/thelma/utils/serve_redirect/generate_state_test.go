package serve_redirect

import (
	"encoding/base64"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_generateState(t *testing.T) {
	state, err := generateState()
	require.NoError(t, err)

	decodedBytes, err := base64.URLEncoding.DecodeString(state)
	require.NoError(t, err)
	require.Len(t, decodedBytes, 32)
}
