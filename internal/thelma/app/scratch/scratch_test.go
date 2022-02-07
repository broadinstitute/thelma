package scratch

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func Test_Scratch(t *testing.T) {
	type testState struct {
		scratch Scratch
		tmpdir  string
	}

	testCases := []struct {
		name           string
		thelmaSettings map[string]interface{}
		testFn         func(*testing.T, testState)
	}{
		{
			name: "should create multiple scratch directories and clean them all up when cleanup is enabled",
			testFn: func(t *testing.T, ts testState) {
				_scratch := ts.scratch

				d1, err := _scratch.Mkdir("testdir")
				assert.NoError(t, err)
				assert.DirExists(t, d1)
				assert.Regexp(t, "testdir", d1)

				root := path.Dir(d1)
				assert.Equal(t, ts.tmpdir, path.Dir(root), "Scratch dirs should be created in the configured tmp directory")

				d2, err := _scratch.Mkdir("testdir")
				assert.NoError(t, err)
				assert.DirExists(t, d2)
				assert.Regexp(t, "testdir", d2)
				assert.NotEqual(t, d1, d2, "Same nickname should produce different directories")
				assert.Equal(t, root, path.Dir(d2), "Scratch dirs should be created in a common root directory")

				d3, err := _scratch.Mkdir("foobar")
				assert.NoError(t, err)
				assert.DirExists(t, d3)
				assert.Regexp(t, "foobar", d3)
				assert.Equal(t, root, path.Dir(d3), "Scratch dirs should be created in a common root directory")

				// make sure we can create a file inside our tmp dir
				err = os.WriteFile(path.Join(d1, "test.txt"), []byte("here is some content!"), 0644)
				assert.NoError(t, err, "Should be able to create new files inside scratch dir")

				err = _scratch.Cleanup()
				if !assert.NoError(t, err) {
					return
				}
				assert.NoDirExists(t, d1)
				assert.NoDirExists(t, d2)
				assert.NoDirExists(t, d3)
				assert.NoDirExists(t, root)
				assert.DirExists(t, ts.tmpdir, "Cleanup should not delete tmp dir")
			},
		},
		{
			name: "should not clean up when cleanup is disabled",
			thelmaSettings: map[string]interface{}{
				"scratch.cleanuponexit": false,
			},
			testFn: func(t *testing.T, ts testState) {
				_scratch := ts.scratch

				d, err := _scratch.Mkdir("testdir")
				assert.NoError(t, err)
				assert.DirExists(t, d)

				err = _scratch.Cleanup()
				if !assert.NoError(t, err) {
					return
				}
				assert.DirExists(t, d, "Cleanup should not delete scratch directories when it is disabled")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpdir := t.TempDir()
			thelmaSettings := map[string]interface{}{
				"scratch.tmpdir": tmpdir,
			}
			if tc.thelmaSettings != nil {
				for k, v := range tc.thelmaSettings {
					thelmaSettings[k] = v
				}
			}
			thelmaConfig, err := config.NewTestConfig(thelmaSettings)
			if !assert.NoError(t, err) {
				return
			}
			_scratch, err := NewScratch(thelmaConfig)
			if !assert.NoError(t, err) {
				return
			}
			if tc.testFn != nil {
				tc.testFn(t, testState{_scratch, tmpdir})
			}
		})
	}
}
