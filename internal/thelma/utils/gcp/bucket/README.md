# bucket

This package contains some helpful abstractions for interacting with GCS buckets.

## Features

### Basic GCS Operations

```go
var err error // error handling omitted

b := bucket.NewBucket("my-bucket")

// upload file to bucket
err = b.Upload("/tmp/my-file.json", "my-object.json")

// download file from bucket
err = b.Upload("my-object.json", "/tmp/my-file.json")

// write to bucket
err = b.Write("my-object.json", []byte(`{"foo":"bar"}`))

// read from bucket
data, err := b.Read("my-object.json")

// delete object
err = b.Delete("my-object.json")
```

See the `Bucket` interface in `bucket.go` for a full list of supported operations.

### Distributed Locking

```go
var err error // error handling omitted

b := bucket.NewBucket("my-bucket")

// create a new Locker associated with `gs://my-bucket/my-lock-object`
locker := b.NewLocker("my-lock-object", 30 * time.Second, func(options *lock.Options) {
  options.ExpiresAfter = 300 * time.Second // if left unset, the lock will never expire
})

// Wait up to 30 seconds to acquire the lock, returning an error if we time out.
// If there is an existing lock, but it's more than 300 seconds old, 
// it will be expired (deleted so that we can claim the lock).
lockId, err := locker.Lock()

// do stuff while we hold the lock
// ...

// release the lock
err = locker.Unlock(lockId)

```

### Smoke Testing

Often, it's useful to test code that depends on GCS against a real GCS bucket. This package includes a TestBucket feature that creates virtual Buckets, backed by the shared `thelma-gcs-integration-test` bucket (in the `dsp-tools-k8s` project).

Every TestBucket uses a **random path prefix**, so that `Read()` and `Write()` calls for the path `my-object.json` will translate to `<random-prefix>/my-object.json`. This prevents different tests (and even multiple concurrent runs of the same test) from stepping on each other.

Any objects written to the TestBucket are **automatically cleaned up** when the test finishes.

Two packages, `assert` and `require`, are provided, to **support assertions on objects in the bucket**. (`require` includes all the same assertions as `assert` but causes the test to fail immediately if an assertion fails, just like Testify's `require` package).

```go

func TestSomething(t *testing.T) {
  // create a new TestBucket
  b := bucket.NewTestBucket(t)

  // do something with bucket, eg. b.Upload("/tmp/my-file.json", "my-object.json")
  // ...
  
  // use the testing/assert and testing
  assert.ObjectExists(t, b, "my-object.json")
  
  // no need to clean up my-object.json, it will be deleted automatically
}
```

See `tests/example_smoke_test.go` for a complete, runnable example.

### Instrumentation

* Every GCS API call is logged at debug level, with relevant contextual fields.
* Every GCS API call is assigned a random id which is included in all related log messages for the call.
* Durations for API calls are logged.

Example message:

```json
{
  "level": "debug",
  "bucket": {
    "name": "my-bucket",
    "prefix": "my-prefix/"
  },
  "object": {
    "name": "my-object",
    "url": "gs://my-bucket/my-prefix/my-object"
  },
  "call": {
    "kind": "write",
    "id": "ba7b5c"
  },
  "attrs":{
    "cache-control": "public, max-age=1337"
  },
  "time": "2022-02-25T13:31:14-05:00",
  "message": "1 attributes will be updated"
}
```
