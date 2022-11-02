// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	google "github.com/broadinstitute/thelma/internal/thelma/clients/google"
	bucket "github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"

	kubectl "github.com/broadinstitute/thelma/internal/thelma/tools/kubectl"

	mock "github.com/stretchr/testify/mock"

	pubsub "cloud.google.com/go/pubsub"

	terraapi "github.com/broadinstitute/thelma/internal/thelma/clients/google/terraapi"
)

// Clients is an autogenerated mock type for the Clients type
type Clients struct {
	mock.Mock
}

// Bucket provides a mock function with given fields: name, options
func (_m *Clients) Bucket(name string, options ...bucket.BucketOption) (bucket.Bucket, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 bucket.Bucket
	if rf, ok := ret.Get(0).(func(string, ...bucket.BucketOption) bucket.Bucket); ok {
		r0 = rf(name, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(bucket.Bucket)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, ...bucket.BucketOption) error); ok {
		r1 = rf(name, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Kubectl provides a mock function with given fields:
func (_m *Clients) Kubectl() (kubectl.Kubectl, error) {
	ret := _m.Called()

	var r0 kubectl.Kubectl
	if rf, ok := ret.Get(0).(func() kubectl.Kubectl); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(kubectl.Kubectl)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PubSub provides a mock function with given fields: projectId
func (_m *Clients) PubSub(projectId string) (*pubsub.Client, error) {
	ret := _m.Called(projectId)

	var r0 *pubsub.Client
	if rf, ok := ret.Get(0).(func(string) *pubsub.Client); ok {
		r0 = rf(projectId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pubsub.Client)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(projectId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetSubject provides a mock function with given fields: subject
func (_m *Clients) SetSubject(subject string) google.Clients {
	ret := _m.Called(subject)

	var r0 google.Clients
	if rf, ok := ret.Get(0).(func(string) google.Clients); ok {
		r0 = rf(subject)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(google.Clients)
		}
	}

	return r0
}

// Terra provides a mock function with given fields:
func (_m *Clients) Terra() (terraapi.TerraClient, error) {
	ret := _m.Called()

	var r0 terraapi.TerraClient
	if rf, ok := ret.Get(0).(func() terraapi.TerraClient); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(terraapi.TerraClient)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewClients interface {
	mock.TestingT
	Cleanup(func())
}

// NewClients creates a new instance of Clients. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewClients(t mockConstructorTestingTNewClients) *Clients {
	mock := &Clients{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
