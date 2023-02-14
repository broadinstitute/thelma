// Package statefixtures provides a fake state provider for use in unit tests
package statefixtures

import (
	"embed"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statebucket"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"
)

var fixtureDirs = struct {
	thelmaHome  string
	statebucket string
}{
	thelmaHome:  "thelma-home",
	statebucket: "statebucket",
}

// We can't use testdata because tests that use this code live outside this package. So, we use embedded fs instead
//go:embed fixtures
var fixturesFS embed.FS

// NewFakeStateLoader (FOR USE IN TESTS ONLY) returns a state loader that loads fake state from test fixtures.
func NewFakeStateLoader(fixture FixtureName, t *testing.T, thelmaHome string) (terra.StateLoader, error) {
	// copy fixture files into state bucket dir and thelma home
	statebucketDir := t.TempDir()
	if err := copyFixture(fixture, fixtureDirs.statebucket, statebucketDir); err != nil {
		return nil, err
	}
	if err := copyFixture(fixture, fixtureDirs.thelmaHome, thelmaHome); err != nil {
		return nil, err
	}

	// create new state bucket
	sb, err := statebucket.NewFake(statebucketDir)
	if err != nil {
		return nil, err
	}

	// create new underlying loader
	loader := gitops.NewStateLoader(thelmaHome, sb)

	if err != nil {
		return nil, err
	}

	return loader, nil
}

// copyFixture copies all fixture files out of the embedded filesystem into the target directory
func copyFixture(fixture FixtureName, subDir string, targetDir string) error {
	// it seems we have to use "." here. "./fixtures/default" and other relative paths that should exist apparently do not work?
	root := "."

	return fs.WalkDir(fixturesFS, root, func(srcPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// compute path prefix for fixture dir, eg.
		//   "fixtures/default/thelma-home/"
		pathPrefix := path.Join("fixtures", fixture.String(), subDir, "")

		if !strings.HasPrefix(srcPath, pathPrefix) {
			// ignore files outside subDir
			return nil
		}

		relativePath := strings.TrimPrefix(srcPath, pathPrefix)
		targetPath := path.Join(targetDir, relativePath)

		// if dir, create it
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0700)
		}

		// safeguard to make sure we don't accidentally overwrite files in someone's real terra-helmfile clone
		_, err = os.Stat(targetPath)
		if err == nil {
			return fmt.Errorf("file %s should not exist, but it does", targetPath)
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("unexpected error checking if file %s exists: %v", targetPath, err)
		}

		// write file
		content, err := fs.ReadFile(fixturesFS, srcPath)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", srcPath, err)
		}

		return os.WriteFile(targetPath, content, 0755)
	})
}
