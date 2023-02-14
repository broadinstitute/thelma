package gitops

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"path"
)

const versionsDir = "versions"

// Versions is for manipulating chart release versions in terra-helmfile (eg. versions/app/dev.yaml, versions/cluster/dev.yaml)
type Versions interface {
	// GetSnapshot returns a VersionSnapshot reference for the given release type & version set
	GetSnapshot(releaseType terra.ReleaseType, versionSet VersionSet) VersionSnapshot
}

// Implements public Versions interface
type versions struct {
	thelmaHome string
	snapshots  map[terra.ReleaseType]map[VersionSet]VersionSnapshot
}

// NewVersions returns a new Versions instance
func NewVersions(thelmaHome string) (Versions, error) {
	snapshots, err := loadSnapshots(thelmaHome)
	if err != nil {
		return nil, err
	}

	return &versions{
		thelmaHome: thelmaHome,
		snapshots:  snapshots,
	}, nil
}

func (v *versions) GetSnapshot(releaseType terra.ReleaseType, versionSet VersionSet) VersionSnapshot {
	return v.snapshots[releaseType][versionSet]
}

// Returns filePath to snapshot file release type / version set. (eg "versions/app/dev.yaml")
func snapshotPath(thelmaHome string, releaseType terra.ReleaseType, set VersionSet) string {
	fileName := fmt.Sprintf("%s.yaml", set)
	return path.Join(thelmaHome, versionsDir, releaseType.String(), fileName)
}

func loadSnapshots(thelmaHome string) (map[terra.ReleaseType]map[VersionSet]VersionSnapshot, error) {
	snapshots := make(map[terra.ReleaseType]map[VersionSet]VersionSnapshot)

	for _, releaseType := range terra.ReleaseTypes() {
		_, exists := snapshots[releaseType]
		if !exists {
			snapshots[releaseType] = make(map[VersionSet]VersionSnapshot)
		}

		for _, versionSet := range VersionSets() {
			fileName := snapshotPath(thelmaHome, releaseType, versionSet)
			snap, err := loadSnapshot(fileName)
			if err != nil {
				return nil, fmt.Errorf("error loading %s %s snapshot from %s: %v", releaseType, versionSet, fileName, err)
			}
			snapshots[releaseType][versionSet] = snap
		}
	}

	return snapshots, nil
}
