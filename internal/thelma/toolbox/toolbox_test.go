package toolbox

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func Test_FindToolsDir(t *testing.T) {
	rootdir := t.TempDir()
	// expand symlinks when constructing test paths so that comparisons succeed later
	rootdir, err := filepath.EvalSymlinks(rootdir)
	require.NoError(t, err)

	// create fake thelma executable in the fake thelma root
	thelmaPath := path.Join(rootdir, "bin", "thelma")
	require.NoError(t, os.MkdirAll(path.Join(rootdir, "bin"), 0755))
	require.NoError(t, os.WriteFile(thelmaPath, []byte("#!/bin/sh\necho fake thelma"), 0755))

	// compute paths we care about - eg. $ROOT/tools/bin/helm
	toolsDir := path.Join(rootdir, toolsDirName)
	exeDir := path.Join(toolsDir, executableDirName)
	helmExecutablePath := path.Join(exeDir, verifyTool)

	// make sure that error is returned if tools dir not found
	_, err = findToolsDir(thelmaPath)
	require.Error(t, err)
	assert.ErrorContains(t, err, fmt.Sprintf("%s does not exist", helmExecutablePath))

	// create a fake executable so that our check passes and verify
	// the expected path is returned
	exedir := path.Join(rootdir, toolsDirName, executableDirName)
	require.NoError(t, os.MkdirAll(exedir, 0755))
	require.NoError(t, os.WriteFile(helmExecutablePath, []byte("#!/bin/sh\necho fake tool"), 0755))

	// make sure correct path was resolved
	resolvedPath, err := findToolsDir(thelmaPath)
	require.NoError(t, err)
	assert.Equal(t, toolsDir, resolvedPath)
}
