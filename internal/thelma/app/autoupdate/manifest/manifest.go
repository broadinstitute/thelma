// Package manifest contains logic for parsing Thelma build manifests
package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

const filename = "build.json"

// EnsureMatches ensures the build manifest for the given release directory matches the given version
func EnsureMatches(releaseDirectory string, version string) error {
	manifestVersion, err := Version(releaseDirectory)
	if err != nil {
		return err
	}

	if manifestVersion != version {
		return fmt.Errorf("error verifying release directory %s: build manifest version %s does not match desired Thelma version %s", releaseDirectory, manifestVersion, version)
	}

	return nil
}

// Version returns the version recorded in a Thelma build manifest
func Version(releaseDirectory string) (string, error) {
	manifestFile := path.Join(releaseDirectory, filename)
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return "", fmt.Errorf("error reading build manifest %s: %v", manifestFile, err)
	}

	type manifest struct {
		Version string `json:"version"`
	}
	var m manifest
	if err = json.Unmarshal(content, &m); err != nil {
		return "", fmt.Errorf("error parsing build manifest %s: %v", manifestFile, err)
	}
	if m.Version == "" {
		return "", fmt.Errorf("error parsing build manifest %s (version field not found): %v", manifestFile, err)
	}

	return m.Version, nil
}
