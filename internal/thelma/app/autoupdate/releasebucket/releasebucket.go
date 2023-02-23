package releasebucket

import (
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/manifest"
	"github.com/broadinstitute/thelma/internal/thelma/app/scratch"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
	"golang.org/x/mod/semver"
	"os"
	"path"
	"strings"
)

const tagsFile = "tags.json"
const tmpdirName = "releasebucket"

// ReleaseBucket represents the Thelma releases bucket
type ReleaseBucket interface {
	// ResolveTagOrVersion given a tag like "latest", or a semantic version
	// string like "1.2.3", resolve to a proper Thelma semantic version like "v1.2.3"
	ResolveTagOrVersion(tagOrVersion string) (string, error)
	// DownloadAndUnpack will download and unpack a release archive in a tmp directory,
	// returning the path to the unpacked release archive
	DownloadAndUnpack(archive Archive) (path string, err error)
}

// New returns a new ReleaseBucket
func New(releasesBucket bucket.Bucket, runner shell.Runner, scratch scratch.Scratch) ReleaseBucket {
	return &releaseBucket{
		bucket:      releasesBucket,
		runner:      runner,
		scratch:     scratch,
		checksummer: newChecksummer(releasesBucket),
	}
}

type releaseBucket struct {
	bucket      bucket.Bucket
	runner      shell.Runner
	scratch     scratch.Scratch
	checksummer checksummer
}

func (r *releaseBucket) ResolveTagOrVersion(versionOrTag string) (string, error) {
	tags, err := r.fetchTags()
	if err != nil {
		return "", err
	}
	v, exists := tags[versionOrTag]
	if exists {
		log.Debug().Msgf("Tag %q is an alias for %q", versionOrTag, v)
		return v, nil
	}

	normalizedVersion := versionOrTag
	if !strings.HasPrefix(normalizedVersion, "v") {
		normalizedVersion = fmt.Sprintf("v%s", versionOrTag)
	}
	if !semver.IsValid(normalizedVersion) {
		return "", fmt.Errorf("%q is not a valid Thelma tag or semantic version", versionOrTag)
	}
	return normalizedVersion, nil
}

// fetch list of tags (version aliases) from the thelma releases bucket
func (r *releaseBucket) fetchTags() (map[string]string, error) {
	content, err := r.bucket.Read(tagsFile)
	if err != nil {
		return nil, fmt.Errorf("error loading %s from %s: %v", tagsFile, r.bucket.Name(), err)
	}
	tags := make(map[string]string)
	if err = json.Unmarshal(content, &tags); err != nil {
		return nil, fmt.Errorf("error parsing %s from %s: %v", tagsFile, r.bucket.Name(), err)
	}
	return tags, nil
}

func (r *releaseBucket) DownloadAndUnpack(archive Archive) (string, error) {
	if err := r.verifyExistsInBucket(archive); err != nil {
		return "", err
	}

	tmpDir, err := r.scratch.Mkdir(tmpdirName)
	if err != nil {
		return "", err
	}

	localArchive := path.Join(tmpDir, archive.Filename())

	if err = r.bucket.Download(archive.ObjectPath(), localArchive); err != nil {
		return "", fmt.Errorf("error downloading release archive gs://%s/%s to %s: %v", r.bucket.Name(), archive.ObjectPath(), localArchive, err)
	}

	// verify archive checksum
	if err = r.checksummer.verify(archive, localArchive); err != nil {
		return "", err
	}

	// unpack release archive .tar.gz with `tar`
	unpackDir := path.Join(tmpDir, archive.Version())
	if err = r.unpack(localArchive, unpackDir); err != nil {
		return "", err
	}

	// make sure build.json file in unpackDir contains the version we expect
	if err = manifest.EnsureMatches(unpackDir, archive.Version()); err != nil {
		return "", err
	}

	return unpackDir, nil
}

func (r *releaseBucket) unpack(localArchive string, unpackDir string) error {
	if err := os.MkdirAll(unpackDir, 0755); err != nil {
		return fmt.Errorf("error unpacking release archive %s: %v", localArchive, err)
	}

	// this is a lot less verbose than trying to accomplish the same thing in go
	err := r.runner.Run(shell.Command{
		Prog: "tar",
		Args: []string{
			"-xz",
			"-C",
			unpackDir,
			"-f",
			localArchive,
		},
	})
	if err != nil {
		return fmt.Errorf("error unpacking release archive %s: %v", localArchive, err)
	}

	return nil
}

func (r *releaseBucket) verifyExistsInBucket(archive Archive) error {
	exists, err := r.bucket.Exists(archive.ObjectPath())
	if err != nil {
		return fmt.Errorf("error validating release version %s: %v", archive.Version(), err)
	}
	if !exists {
		return fmt.Errorf("release archive gs://%s/%s does not exist", r.bucket.Name(), archive.ObjectPath())
	}
	return nil
}
