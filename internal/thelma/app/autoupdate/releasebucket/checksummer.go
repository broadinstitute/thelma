package releasebucket

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"io"
	"os"
	"strings"
)

// checksummer verifies sha256sums for Thelma release archives
type checksummer interface {
	// verify returns an error if the local copy of a release archive file does not match its published sha256sum
	verify(archive Archive, file string) error
}

type checksummerImpl struct {
	bucket bucket.Bucket
}

func newChecksummer(bucket bucket.Bucket) checksummer {
	return &checksummerImpl{
		bucket: bucket,
	}
}

func (s *checksummerImpl) verify(archive Archive, file string) error {
	archiveSha245Sum, err := s.getReleaseArchiveSha256Sum(archive)
	if err != nil {
		return fmt.Errorf("error identifying sha256sum for Thelma version %s: %v", archive.Version(), err)
	}

	localSha256Sum, err := computeSha256Sum(file)
	if err != nil {
		return err
	}
	if localSha256Sum != archiveSha245Sum {
		return fmt.Errorf("downloaded release archive %s has incorrect sha256sum (has %s, should be %s)", file, localSha256Sum, archiveSha245Sum)
	}
	return nil
}

// fetch the published sha256sum for a given release archive and return it.
// eg. "5b0fac41f493099924dbcbcc40ac7b2d61d342e5044b2e2ffd4b771863756a65"
//
// checksums are recorded in a file in the thelma release archive directory at
//
//	releases/<version>/thelma_<version>_SHA256SUMS
//
// and the file looks like:
//
// 5b0fac41f493099924dbcbcc40ac7b2d61d342e5044b2e2ffd4b771863756a65  thelma_<version>_darwin_amd64.tar.gz
// e50372e48fa3750a917a4e61c67c456015a016f1beac1e55d241349eb44d266d  thelma_<version>_darwin_arm64.tar.gz
// a9afa31857e0b9b8206202a3f31f6e968602e08c324fd270822ae824744cb1c4  thelma_<version>_linux_amd64.tar.gz
func (s *checksummerImpl) getReleaseArchiveSha256Sum(archive Archive) (string, error) {
	checksumsContent, err := s.bucket.Read(archive.Sha256SumObjectPath())
	if err != nil {
		return "", fmt.Errorf("error reading checksum object gs://%s/%s: %v", s.bucket.Name(), archive.Sha256SumObjectPath(), err)
	}

	sc := bufio.NewScanner(bytes.NewReader(checksumsContent))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasSuffix(line, archive.Filename()) {
			return strings.Fields(line)[0], nil
		}
	}
	return "", fmt.Errorf("found no matching checksum for %s in gs://%s/%s", archive.Filename(), s.bucket.Name(), archive.Sha256SumObjectPath())
}

// compute sha256sum for a local file
func computeSha256Sum(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("error computing sha256sum for %s: %v", file, err)
	}

	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", fmt.Errorf("error computing sha256sum for %s: %v", file, err)
	}

	if err = f.Close(); err != nil {
		return "", fmt.Errorf("error computing sha256sum for %s: %v", file, err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
