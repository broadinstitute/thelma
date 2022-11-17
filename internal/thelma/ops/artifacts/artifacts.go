package artifacts

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/api"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

const timestampFormatString = "20060102.150405"

type Options struct {
	// Dir local directory where operational artifacts should be written
	Dir string
	// Upload if true, upload exported operational artifacts to the release's cluster artifact bucket
	Upload bool
}

type Location struct {
	FilesystemPath  string `yaml:"path,omitempty" json:"url,omitempty"`
	CloudConsoleURL string `yaml:"url,omitempty" json:"path,omitempty"`
}

type Artifacts interface {
	NewManager(artifactType Type, options Options) Manager
}

func New(bucketFactory api.BucketFactory) Artifacts {
	return &artifacts{bucketFactory: bucketFactory}
}

// DefaultArtifactsURL returns the artifacts base URL for this environment's default cluster.
// If the environment spans multiple clusters, artifacts for the releases outside the default cluster will not be
// found at this URL
func DefaultArtifactsURL(env terra.Environment) string {
	return bucket.CloudConsoleObjectListURL(env.DefaultCluster().ArtifactBucket(), env.Name())
}

type artifacts struct {
	bucketFactory api.BucketFactory
}

func (a *artifacts) NewManager(artifactType Type, options Options) Manager {
	return NewManager(artifactType, a.bucketFactory, options)
}
