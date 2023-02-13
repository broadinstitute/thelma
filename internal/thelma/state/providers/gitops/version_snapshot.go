package gitops

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type VersionSnapshot interface {
	// ReleaseDefined returns true if the given release is defined in this snapshot
	ReleaseDefined(releaseName string) bool
	// ChartVersion returns the chartVersion for the given release in this snapshot. If the release is not defined, returns ""
	ChartVersion(releaseName string) string
	// AppVersion returns the chartVersion for the given release in this snapshot. If the release is not defined, returns ""
	AppVersion(releaseName string) string
}

// Represents a version snapshot file in terra-helmfile (eg. "versions/app/dev.yaml")
type snapshot struct {
	filePath string
	data     *snapshotData
}

// struct used for deserializing a version snapshot
type snapshotData struct {
	Releases map[string]struct {
		ChartVersion string `yaml:"chartVersion"`
		AppVersion   string `yaml:"appVersion"`
	} `yaml:"releases"`
}

func loadSnapshot(filePath string) (VersionSnapshot, error) {
	data, err := readSnapshotFile(filePath)
	if err != nil {
		return nil, err
	}

	return &snapshot{
		filePath: filePath,
		data:     data,
	}, nil
}

func (s *snapshot) ReleaseDefined(releaseName string) bool {
	_, exists := s.data.Releases[releaseName]
	return exists
}

func (s *snapshot) ChartVersion(releaseName string) string {
	if !s.ReleaseDefined(releaseName) {
		return ""
	}
	return s.data.Releases[releaseName].ChartVersion
}

func (s *snapshot) AppVersion(releaseName string) string {
	if !s.ReleaseDefined(releaseName) {
		return ""
	}
	return s.data.Releases[releaseName].AppVersion
}

// Reload snapshot data from disk
func (s *snapshot) reload() error {
	data, err := readSnapshotFile(s.filePath)
	if err != nil {
		return fmt.Errorf("error reloading snapshot: %v", err)
	}
	s.data = data
	return nil
}

// Unmarshal the snapshot at the given filePath into a struct
func readSnapshotFile(filePath string) (*snapshotData, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot file %s: %v", filePath, err)
	}

	snapshot := &snapshotData{}
	if err := yaml.Unmarshal(content, snapshot); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot file %s: %v", filePath, err)
	}

	if snapshot.Releases == nil {
		return nil, fmt.Errorf("empty snapshot file %s: %v", filePath, err)
	}

	return snapshot, nil
}
