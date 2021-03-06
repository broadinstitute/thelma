// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	bucket "github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	mock "github.com/stretchr/testify/mock"

	object "github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket/object"

	storage "cloud.google.com/go/storage"

	time "time"
)

// Bucket is an autogenerated mock type for the Bucket type
type Bucket struct {
	mock.Mock
}

// Attrs provides a mock function with given fields: objectName
func (_m *Bucket) Attrs(objectName string) (*storage.ObjectAttrs, error) {
	ret := _m.Called(objectName)

	var r0 *storage.ObjectAttrs
	if rf, ok := ret.Get(0).(func(string) *storage.ObjectAttrs); ok {
		r0 = rf(objectName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage.ObjectAttrs)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(objectName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Close provides a mock function with given fields:
func (_m *Bucket) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: objectName
func (_m *Bucket) Delete(objectName string) error {
	ret := _m.Called(objectName)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(objectName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Download provides a mock function with given fields: objectName, localPath
func (_m *Bucket) Download(objectName string, localPath string) error {
	ret := _m.Called(objectName, localPath)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(objectName, localPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: objectName
func (_m *Bucket) Exists(objectName string) (bool, error) {
	ret := _m.Called(objectName)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(objectName)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(objectName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Name provides a mock function with given fields:
func (_m *Bucket) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// NewLocker provides a mock function with given fields: objectName, maxWait, options
func (_m *Bucket) NewLocker(objectName string, maxWait time.Duration, options ...bucket.LockerOption) bucket.Locker {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, objectName, maxWait)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 bucket.Locker
	if rf, ok := ret.Get(0).(func(string, time.Duration, ...bucket.LockerOption) bucket.Locker); ok {
		r0 = rf(objectName, maxWait, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(bucket.Locker)
		}
	}

	return r0
}

// Read provides a mock function with given fields: objectName
func (_m *Bucket) Read(objectName string) ([]byte, error) {
	ret := _m.Called(objectName)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string) []byte); ok {
		r0 = rf(objectName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(objectName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: objectName, attrs
func (_m *Bucket) Update(objectName string, attrs ...object.AttrSetter) error {
	_va := make([]interface{}, len(attrs))
	for _i := range attrs {
		_va[_i] = attrs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, objectName)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, ...object.AttrSetter) error); ok {
		r0 = rf(objectName, attrs...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Upload provides a mock function with given fields: localPath, objectName, attrs
func (_m *Bucket) Upload(localPath string, objectName string, attrs ...object.AttrSetter) error {
	_va := make([]interface{}, len(attrs))
	for _i := range attrs {
		_va[_i] = attrs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, localPath, objectName)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, ...object.AttrSetter) error); ok {
		r0 = rf(localPath, objectName, attrs...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Write provides a mock function with given fields: objectName, content, attrs
func (_m *Bucket) Write(objectName string, content []byte, attrs ...object.AttrSetter) error {
	_va := make([]interface{}, len(attrs))
	for _i := range attrs {
		_va[_i] = attrs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, objectName, content)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []byte, ...object.AttrSetter) error); ok {
		r0 = rf(objectName, content, attrs...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
