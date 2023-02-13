package root

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/version"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func Test_ToolsDir(t *testing.T) {
	rootdir := t.TempDir()
	_root := New(rootdir)
	_, err := _root.ToolsDir()
	require.Error(t, err)
	assert.ErrorContains(t, err, "tools/bin/helm does not exist")

	bindir := path.Join(rootdir, "releases", version.Version, "tools", "bin")
	require.NoError(t, os.MkdirAll(bindir, 0755))
	require.NoError(t, os.WriteFile(path.Join(bindir, helm.ProgName), []byte("#!/bin/sh\necho fake helm"), 0755))
	_tools, err := _root.ToolsDir()
	require.NoError(t, err)
	assert.Equal(t, bindir, _tools.Bin())
}
