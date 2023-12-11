// Package api contains interfaces for client factories in the clients package.
// Because there are so many interdependencies between clients (eg. we need to load service account keys from Vault
// in order to create Google API clients), keeping interfaces in a separate package avoids dependency cycles.
package api

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
)

type BucketFactory interface {
	Bucket(name string, options ...bucket.BucketOption) (bucket.Bucket, error)
}
