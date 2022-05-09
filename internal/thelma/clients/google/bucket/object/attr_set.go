package object

import (
	"cloud.google.com/go/storage"
	"github.com/rs/zerolog"
)

// AttrSet used to set optional GCS object attributes in upload, write, and update operations
type AttrSet struct {
	cacheControl *string
	// note: if you add a new field here, be sure to update applyToWriter(), asUpdateAttrs(), and asLogFields() as well
	// maybe someday we can get fancy with reflection to reduce the duplication
}

// AttrSetter used to set optional GCS object attributes in upload, write, and update operations
type AttrSetter func(AttrSet) AttrSet

// CacheControl sets the Cache-Control attribute of an object
func (a AttrSet) CacheControl(cacheControl string) AttrSet {
	a.cacheControl = &cacheControl
	return a
}

// GetCacheControl returns the configured Cache-Control object for this AttrSet. (nil if not set)
func (a AttrSet) GetCacheControl() *string {
	return a.cacheControl
}

// ApplyToWriter apply attributes to a storage writer (used in Upload, Write functions)
func (a AttrSet) applyToWriter(writer *storage.Writer) {
	if a.cacheControl != nil {
		writer.CacheControl = *a.cacheControl
	}
}

// AsUpdateAttrs convert to an ObjectAttrsToUpdate
func (a AttrSet) asUpdateAttrs() storage.ObjectAttrsToUpdate {
	var uattrs storage.ObjectAttrsToUpdate
	if a.cacheControl != nil {
		uattrs.CacheControl = *a.cacheControl
	}
	return uattrs
}

// WriteToLogEvent write log a message indicating what attributes this AttrSet will update
func (a AttrSet) writeToLogEvent(event *zerolog.Event) {
	m := a.asMap()
	event.Interface("attrs", m).Msgf("%d attributes will be updated", len(m))
}

// asMap write attributes that have been set to map (used for logging)
func (a AttrSet) asMap() map[string]interface{} {
	result := make(map[string]interface{})
	if a.cacheControl != nil {
		result["cache-control"] = a.cacheControl
	}
	return result
}
