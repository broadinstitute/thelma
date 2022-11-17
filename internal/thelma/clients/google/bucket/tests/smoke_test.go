//go:build smoke
// +build smoke

package tests

// Notes:
// * This file contains a smoke test that interacts with a real GCS bucket. By default it is ignored by `go test`.
//   You can run it with: go test ./... -tags smoke
// * This file lives in its own package so that it can use the `bassert` package without creating a dependency cycle

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket/lock"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket/object"
	bassert "github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket/testing/assert"
	brequire "github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket/testing/require"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
	"time"
)

const lockExpiresTimeout = 5 * time.Second
const lockWaitTimeout = 1 * time.Second
const testCacheControlHeader = "public, max-age=1337"

func TestBucket_LockAndUnlock(t *testing.T) {
	_bucket := bucket.NewTestBucket(t)

	locker := _bucket.NewLocker("test.lk", lockWaitTimeout, func(options *lock.Options) {
		options.ExpiresAfter = lockExpiresTimeout
	})

	brequire.NoObjectExists(t, _bucket, locker.ObjectName(), "lock file should not exist at test start")

	lockId, err := locker.Lock()
	require.NoError(t, err, "first attempt to acquire lock should succeed")
	brequire.ObjectExists(t, _bucket, locker.ObjectName(), "lock file should exist if lock succeeds")
	brequire.ObjectHasGeneration(t, _bucket, locker.ObjectName(), lockId)

	_, err = locker.Lock()
	require.Error(t, err, "second attempt to acquire lock should time out")
	assert.Regexp(t, "timed out after.*waiting for lock", err.Error())

	err = locker.Unlock(lockId)
	require.NoError(t, err, "unlock should succeed")
	brequire.NoObjectExists(t, _bucket, locker.ObjectName(), "object should not exist after unlock")
}

func TestBucket_ExpiredLockIsCleanedUp(t *testing.T) {
	_bucket := bucket.NewTestBucket(t)

	locker := _bucket.NewLocker("test.lk", lockWaitTimeout, func(options *lock.Options) {
		options.ExpiresAfter = lockExpiresTimeout
	})

	brequire.NoObjectExists(t, _bucket, locker.ObjectName(), "lock file should not exist at test start")

	firstId, err := locker.Lock()
	require.NoError(t, err, "first attempt to acquire lock should succeed")
	brequire.ObjectHasGeneration(t, _bucket, locker.ObjectName(), firstId)

	time.Sleep(lockExpiresTimeout + lockExpiresTimeout/10)
	secondId, err := locker.Lock()
	require.NoError(t, err, "second attempt to acquire lock should succeed, since lock has expired")
	brequire.ObjectHasGeneration(t, _bucket, locker.ObjectName(), secondId)
	require.NotEqual(t, firstId, secondId, "expect unique generation for second lock")

	err = locker.Unlock(firstId)
	require.NoError(t, err, "unlock with incorrect id should not return error")
	brequire.ObjectExists(t, _bucket, locker.ObjectName(), "unlock with incorrect id should not remove lock")

	err = locker.Unlock(secondId)
	require.NoError(t, err, "unlock with correct id should not return error")
	brequire.NoObjectExists(t, _bucket, locker.ObjectName(), "unlock with correct id should remove lock")
}

func TestBucket_UploadAndDownload(t *testing.T) {
	_bucket := bucket.NewTestBucket(t)

	testContent := "foo\n"
	objectName := "my-test-object"

	testDir := t.TempDir()
	uploadFile := path.Join(testDir, "upload.txt")
	downloadFile := path.Join(testDir, "download.txt")

	require.NoFileExists(t, uploadFile)
	require.NoFileExists(t, downloadFile)
	brequire.NoObjectExists(t, _bucket, objectName)

	err := os.WriteFile(uploadFile, []byte(testContent), 0600)
	require.NoError(t, err)

	err = _bucket.Upload(uploadFile, objectName)
	require.NoError(t, err)

	brequire.ObjectHasContent(t, _bucket, objectName, testContent)

	err = _bucket.Download(objectName, downloadFile)
	require.NoError(t, err)

	fileContent, err := os.ReadFile(downloadFile)
	require.NoError(t, err)
	assert.Equal(t, string(fileContent), testContent)
}

func TestBucket_UploadUpdatesAttributes(t *testing.T) {
	_bucket := bucket.NewTestBucket(t)
	objectName := "my-object"

	file := path.Join(t.TempDir(), "empty.txt")
	err := os.WriteFile(file, []byte(""), 0600)
	require.NoError(t, err)

	err = _bucket.Upload(file, objectName, func(attrs object.AttrSet) object.AttrSet {
		return attrs.CacheControl(testCacheControlHeader)
	})
	require.NoError(t, err)

	brequire.ObjectHasCacheControl(t, _bucket, objectName, testCacheControlHeader)
}

func TestBucket_ReadAndWrite(t *testing.T) {
	_bucket := bucket.NewTestBucket(t)
	objectName := "my-object"
	content := "some data"

	bassert.NoObjectExists(t, _bucket, objectName, "object should not exist at start of test")

	_, err := _bucket.Read(objectName)
	require.Error(t, err, "attempt to read object that does not exist should return error")

	err = _bucket.Write(objectName, []byte(content))
	require.NoError(t, err)
	brequire.ObjectHasContent(t, _bucket, objectName, content)

	readContent, err := _bucket.Read(objectName)
	require.NoError(t, err)

	assert.Equal(t, content, string(readContent))
}

func TestBucket_WriterWritesToObject(t *testing.T) {
	_bucket := bucket.NewTestBucket(t)
	objectName := "my-test-object"
	writer := _bucket.Writer(objectName)

	_, err := writer.Write([]byte("a"))
	require.NoError(t, err)

	_, err = writer.Write([]byte("b"))
	require.NoError(t, err)

	_, err = writer.Write([]byte("c"))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	brequire.ObjectHasContent(t, _bucket, objectName, "abc")
}

func TestBucket_WriterRaisesErrorOnClose(t *testing.T) {
	badBucket, err := bucket.NewBucket("broad-dsp-devops-sdjhk3227-this-bucket-does-not-exist")
	require.NoError(t, err)
	objectName := "my-test-object"
	writer := badBucket.Writer(objectName)

	_, err = writer.Write([]byte("a"))
	require.NoError(t, err)

	err = writer.Close()
	require.ErrorContains(t, err, "googleapi: Error 404: The specified bucket does not exist.")
}

func TestBucket_WriteUpdatesAttributes(t *testing.T) {
	_bucket := bucket.NewTestBucket(t)
	objectName := "my-object"

	err := _bucket.Write(objectName, []byte(""), func(attrs object.AttrSet) object.AttrSet {
		return attrs.CacheControl(testCacheControlHeader)
	})
	require.NoError(t, err)

	brequire.ObjectHasCacheControl(t, _bucket, objectName, testCacheControlHeader)
}

func TestBucket_UpdateAndAttrs(t *testing.T) {
	_bucket := bucket.NewTestBucket(t)
	objectName := "my-object"

	err := _bucket.Write(objectName, []byte(""))
	assert.NoError(t, err)

	attrs, err := _bucket.Attrs(objectName)
	require.NoError(t, err)
	require.NotEqual(t, testCacheControlHeader, attrs.CacheControl)

	err = _bucket.Update(objectName, func(attrs object.AttrSet) object.AttrSet {
		return attrs.CacheControl(testCacheControlHeader)
	})
	require.NoError(t, err)

	attrs, err = _bucket.Attrs(objectName)
	require.NoError(t, err)
	require.Equal(t, testCacheControlHeader, attrs.CacheControl)
}
