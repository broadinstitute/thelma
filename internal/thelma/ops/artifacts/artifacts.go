package artifacts

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/api"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

const timestampFormatString = "2006-01-02T15:04:05"

type Options struct {
	// Dir local directory where operational artifacts should be written
	Dir string
	// Upload if true, upload exported operational artifacts to the release's cluster artifact bucket
	Upload bool
}

type Location struct {
	FilesystemPath  string `yaml:",omitempty" json:",omitempty"`
	CloudConsoleURL string `yaml:",omitempty" json:",omitempty"`
}

// Manager is for writing operational artifacts to standard locations on the local filesystem or a GCS bucket
type Manager interface {
	// Writer returns a writer that can be used to write artifact data to the given destination(s)
	// Note: Callers are responsible for closing the writer once data has been written!
	Writer(release terra.Release, path string) (io.WriteCloser, error)

	// GetArtifactLocation returns links to an individual artifact path.
	// Either of the fields in the returned Links struct may be empty, depending on how the Manager is configured.
	// For example if the Manager is not configured to upload artifacts to GCS, the CloudConsoleLink field will be empty
	GetArtifactLocation(release terra.Release, path string) Location

	// GetArtifactsBaseLocation returns links to the base directory/path for all artifacts for this Release
	// Either of the fields in the returned Links struct may be empty, depending on how the manager is configured
	// For example if the Manager is not configured to upload artifacts to GCS, the CloudConsoleLink field will be empty
	GetArtifactsBaseLocation(release terra.Release) Location
}

type Type int64

const (
	ContainerLog Type = iota
)

func (t Type) pathPrefix() string {
	switch t {
	case ContainerLog:
		return "container-log"
	default:
		panic(fmt.Sprintf("Unknown artifact type: %v", t))
	}
}

func NewManager(artifactType Type, bucketFactory api.BucketFactory, options Options) Manager {
	return &manager{
		options:       options,
		artifactType:  artifactType,
		bucketFactory: bucketFactory,
		timestamp:     time.Now(),
	}
}

type manager struct {
	options       Options
	artifactType  Type
	bucketFactory api.BucketFactory
	timestamp     time.Time
}

func (m *manager) GetArtifactLocation(release terra.Release, path string) Location {
	return Location{
		CloudConsoleURL: m.cloudConsoleLink(release, path),
		FilesystemPath:  m.filesytemPath(release, path),
	}
}

func (m *manager) GetArtifactsBaseLocation(release terra.Release) Location {
	return m.GetArtifactLocation(release, "")
}

func (m *manager) Writer(release terra.Release, _path string) (io.WriteCloser, error) {
	var writeClosers []io.WriteCloser

	relativePath := m.relativePath(release, _path)

	if m.options.Dir != "" {
		outputFile := path.Join(m.options.Dir, relativePath)
		parentDir := path.Dir(outputFile)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return nil, fmt.Errorf("error creating artifact dir %s: %v", parentDir, err)
		}
		file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, fmt.Errorf("error creating artifact file %s: %v", outputFile, err)
		}
		writeClosers = append(writeClosers, file)
	}

	if m.options.Upload {
		_bucket, err := m.bucketFactory.Bucket(bucketName(release))
		if err != nil {
			return nil, err
		}
		writeClosers = append(writeClosers, _bucket.Writer(relativePath))
	}

	return newMultiWriteCloser(writeClosers...), nil
}

func (m *manager) relativePath(release terra.Release, _path string) string {
	return path.Join(m.basePath(release), _path)
}

func (m *manager) cloudConsoleLink(release terra.Release, _path string) string {
	if !m.options.Upload {
		return ""
	}
	return bucket.CloudConsoleURL(bucketName(release), m.relativePath(release, _path))
}

func (m *manager) filesytemPath(release terra.Release, _path string) string {
	if m.options.Dir == "" {
		return ""
	}
	fullPath := path.Join(m.options.Dir, m.relativePath(release, _path))

	abs, err := filepath.Abs(fullPath)
	if err != nil {
		log.Warn().Err(err).Msgf("failed to generate absolute path for: %s", fullPath)
		return ""
	}
	return abs
}

// compute the base subdirectory where this manager will upload files for the given release.
// Eg.
// "my-bee/agora/container-logs/2022-10-07T05:36:02"
func (m *manager) basePath(release terra.Release) string {
	return path.Join(release.Destination().Name(), release.Name(), m.artifactType.pathPrefix(), m.formattedTimestamp())
}

// return the name of the bucket where artifacts should be uploaded for a release
func bucketName(release terra.Release) string {
	return release.Cluster().ArtifactBucket()
}

// return formatted timestamp for use artifact upload paths
func (m *manager) formattedTimestamp() string {
	return m.timestamp.UTC().Format(timestampFormatString)
}
