package object

import (
	"cloud.google.com/go/storage"
	"context"
)

type Object struct {
	Ctx    context.Context
	Handle *storage.ObjectHandle
}
