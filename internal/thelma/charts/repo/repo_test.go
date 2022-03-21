package repo

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket/lock"
	"github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket/object"
	"github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const testLockObj = ".my-lock-object"
const testLockWaitTimeout = 3 * time.Second
const testLockExpireTimeout = 6 * time.Second
const testIndexCacheControl = "fake-index-cache-control-header"
const testChartCacheControl = "fake-chart-cache-control-header"

const testIndexFile = "index.yaml"

type testState struct {
	bucket *mocks.Bucket
	locker *mocks.Locker
	repo   Repo
}

func TestDownloadIndex(t *testing.T) {
	ts := setupMocks()
	ts.bucket.On("Download", indexObject, testIndexFile).Return(nil)

	err := ts.repo.DownloadIndex(testIndexFile)
	require.NoError(t, err)

	ts.AssertExpectations(t)
}

func TestHasIndex(t *testing.T) {
	ts := setupMocks()
	ts.bucket.On("Exists", indexObject).Return(true, nil)

	result, err := ts.repo.HasIndex()
	assert.NoError(t, err)
	assert.True(t, result)

	ts.AssertExpectations(t)
}

func TestUploadIndex(t *testing.T) {
	ts := setupMocks()
	ts.bucket.On("Upload", testIndexFile, indexObject, mock.MatchedBy(setsCacheControlTo(testIndexCacheControl))).Return(nil)

	assert.NoError(t, ts.repo.UploadIndex(testIndexFile))

	ts.AssertExpectations(t)
}

func TestUploadChart(t *testing.T) {
	chartFile := "path/to/chart.tgz"
	chartObject := "charts/chart.tgz"

	ts := setupMocks()
	ts.bucket.On("Upload", chartFile, chartObject, mock.MatchedBy(setsCacheControlTo(testChartCacheControl))).Return(nil)

	assert.NoError(t, ts.repo.UploadChart(chartFile))

	ts.AssertExpectations(t)
}

func TestLocking(t *testing.T) {
	lockId := int64(1234567890)
	ts := setupMocks()
	ts.locker.On("Lock").Return(lockId, nil)
	ts.locker.On("Unlock", lockId).Return(nil)

	assert.False(t, ts.repo.IsLocked())
	assert.NoError(t, ts.repo.Lock())
	assert.True(t, ts.repo.IsLocked())
	assert.NoError(t, ts.repo.Unlock())
	assert.False(t, ts.repo.IsLocked())

	ts.AssertExpectations(t)
}

func TestRepoURL(t *testing.T) {
	ts := setupMocks()
	ts.bucket.On("Name").Return("test-bucket")

	assert.Equal(t, "https://test-bucket.storage.googleapis.com", ts.repo.RepoURL())

	ts.AssertExpectations(t)
}

func setupMocks() *testState {
	_locker := &mocks.Locker{}
	_bucket := &mocks.Bucket{}

	// argument matcher that verifies the locker option sets the correct expire timeout
	setsLockExpireTimeout := func(option bucket.LockerOption) bool {
		opts := lock.Options{}
		option(&opts)
		return opts.ExpiresAfter == testLockExpireTimeout
	}

	_bucket.On("NewLocker", testLockObj, testLockWaitTimeout, mock.MatchedBy(setsLockExpireTimeout)).Return(_locker).Maybe()

	_repo := NewRepo(_bucket, func(options *Options) {
		options.LockObject = testLockObj
		options.LockWaitTimeout = testLockWaitTimeout
		options.LockExpireTimeout = testLockExpireTimeout
		options.IndexCacheControl = testIndexCacheControl
		options.ChartCacheControl = testChartCacheControl
	})

	return &testState{
		locker: _locker,
		bucket: _bucket,
		repo:   _repo,
	}
}

func (ts *testState) AssertExpectations(t *testing.T) bool {
	return ts.locker.AssertExpectations(t) &&
		ts.bucket.AssertExpectations(t)
}

// returns an argument matcher function that checks cache control header was set to a specific value
func setsCacheControlTo(cacheControl string) func(object.AttrSetter) bool {
	return func(setter object.AttrSetter) bool {
		attrs := object.AttrSet{}
		attrs = setter(attrs)
		value := attrs.GetCacheControl()
		return value != nil && *value == cacheControl
	}
}
