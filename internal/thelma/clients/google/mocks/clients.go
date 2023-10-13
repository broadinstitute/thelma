// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	container "cloud.google.com/go/container/apiv1"
	bucket "github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"

	google "github.com/broadinstitute/thelma/internal/thelma/clients/google"

	mock "github.com/stretchr/testify/mock"

	oauth2 "golang.org/x/oauth2"

	pubsub "cloud.google.com/go/pubsub"

	sqladmin "github.com/broadinstitute/thelma/internal/thelma/clients/google/sqladmin"

	terraapi "github.com/broadinstitute/thelma/internal/thelma/clients/google/terraapi"
)

// Clients is an autogenerated mock type for the Clients type
type Clients struct {
	mock.Mock
}

type Clients_Expecter struct {
	mock *mock.Mock
}

func (_m *Clients) EXPECT() *Clients_Expecter {
	return &Clients_Expecter{mock: &_m.Mock}
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
	var r1 error
	if rf, ok := ret.Get(0).(func(string, ...bucket.BucketOption) (bucket.Bucket, error)); ok {
		return rf(name, options...)
	}
	if rf, ok := ret.Get(0).(func(string, ...bucket.BucketOption) bucket.Bucket); ok {
		r0 = rf(name, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(bucket.Bucket)
		}
	}

	if rf, ok := ret.Get(1).(func(string, ...bucket.BucketOption) error); ok {
		r1 = rf(name, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Clients_Bucket_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Bucket'
type Clients_Bucket_Call struct {
	*mock.Call
}

// Bucket is a helper method to define mock.On call
//   - name string
//   - options ...bucket.BucketOption
func (_e *Clients_Expecter) Bucket(name interface{}, options ...interface{}) *Clients_Bucket_Call {
	return &Clients_Bucket_Call{Call: _e.mock.On("Bucket",
		append([]interface{}{name}, options...)...)}
}

func (_c *Clients_Bucket_Call) Run(run func(name string, options ...bucket.BucketOption)) *Clients_Bucket_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]bucket.BucketOption, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(bucket.BucketOption)
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Clients_Bucket_Call) Return(_a0 bucket.Bucket, _a1 error) *Clients_Bucket_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Clients_Bucket_Call) RunAndReturn(run func(string, ...bucket.BucketOption) (bucket.Bucket, error)) *Clients_Bucket_Call {
	_c.Call.Return(run)
	return _c
}

// ClusterManager provides a mock function with given fields:
func (_m *Clients) ClusterManager() (*container.ClusterManagerClient, error) {
	ret := _m.Called()

	var r0 *container.ClusterManagerClient
	var r1 error
	if rf, ok := ret.Get(0).(func() (*container.ClusterManagerClient, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *container.ClusterManagerClient); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*container.ClusterManagerClient)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Clients_ClusterManager_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClusterManager'
type Clients_ClusterManager_Call struct {
	*mock.Call
}

// ClusterManager is a helper method to define mock.On call
func (_e *Clients_Expecter) ClusterManager() *Clients_ClusterManager_Call {
	return &Clients_ClusterManager_Call{Call: _e.mock.On("ClusterManager")}
}

func (_c *Clients_ClusterManager_Call) Run(run func()) *Clients_ClusterManager_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Clients_ClusterManager_Call) Return(_a0 *container.ClusterManagerClient, _a1 error) *Clients_ClusterManager_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Clients_ClusterManager_Call) RunAndReturn(run func() (*container.ClusterManagerClient, error)) *Clients_ClusterManager_Call {
	_c.Call.Return(run)
	return _c
}

// PubSub provides a mock function with given fields: projectId
func (_m *Clients) PubSub(projectId string) (*pubsub.Client, error) {
	ret := _m.Called(projectId)

	var r0 *pubsub.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*pubsub.Client, error)); ok {
		return rf(projectId)
	}
	if rf, ok := ret.Get(0).(func(string) *pubsub.Client); ok {
		r0 = rf(projectId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pubsub.Client)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(projectId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Clients_PubSub_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PubSub'
type Clients_PubSub_Call struct {
	*mock.Call
}

// PubSub is a helper method to define mock.On call
//   - projectId string
func (_e *Clients_Expecter) PubSub(projectId interface{}) *Clients_PubSub_Call {
	return &Clients_PubSub_Call{Call: _e.mock.On("PubSub", projectId)}
}

func (_c *Clients_PubSub_Call) Run(run func(projectId string)) *Clients_PubSub_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Clients_PubSub_Call) Return(_a0 *pubsub.Client, _a1 error) *Clients_PubSub_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Clients_PubSub_Call) RunAndReturn(run func(string) (*pubsub.Client, error)) *Clients_PubSub_Call {
	_c.Call.Return(run)
	return _c
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

// Clients_SetSubject_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetSubject'
type Clients_SetSubject_Call struct {
	*mock.Call
}

// SetSubject is a helper method to define mock.On call
//   - subject string
func (_e *Clients_Expecter) SetSubject(subject interface{}) *Clients_SetSubject_Call {
	return &Clients_SetSubject_Call{Call: _e.mock.On("SetSubject", subject)}
}

func (_c *Clients_SetSubject_Call) Run(run func(subject string)) *Clients_SetSubject_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Clients_SetSubject_Call) Return(_a0 google.Clients) *Clients_SetSubject_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Clients_SetSubject_Call) RunAndReturn(run func(string) google.Clients) *Clients_SetSubject_Call {
	_c.Call.Return(run)
	return _c
}

// SqlAdmin provides a mock function with given fields:
func (_m *Clients) SqlAdmin() (sqladmin.Client, error) {
	ret := _m.Called()

	var r0 sqladmin.Client
	var r1 error
	if rf, ok := ret.Get(0).(func() (sqladmin.Client, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() sqladmin.Client); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sqladmin.Client)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Clients_SqlAdmin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SqlAdmin'
type Clients_SqlAdmin_Call struct {
	*mock.Call
}

// SqlAdmin is a helper method to define mock.On call
func (_e *Clients_Expecter) SqlAdmin() *Clients_SqlAdmin_Call {
	return &Clients_SqlAdmin_Call{Call: _e.mock.On("SqlAdmin")}
}

func (_c *Clients_SqlAdmin_Call) Run(run func()) *Clients_SqlAdmin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Clients_SqlAdmin_Call) Return(_a0 sqladmin.Client, _a1 error) *Clients_SqlAdmin_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Clients_SqlAdmin_Call) RunAndReturn(run func() (sqladmin.Client, error)) *Clients_SqlAdmin_Call {
	_c.Call.Return(run)
	return _c
}

// Terra provides a mock function with given fields:
func (_m *Clients) Terra() (terraapi.TerraClient, error) {
	ret := _m.Called()

	var r0 terraapi.TerraClient
	var r1 error
	if rf, ok := ret.Get(0).(func() (terraapi.TerraClient, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() terraapi.TerraClient); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(terraapi.TerraClient)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Clients_Terra_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Terra'
type Clients_Terra_Call struct {
	*mock.Call
}

// Terra is a helper method to define mock.On call
func (_e *Clients_Expecter) Terra() *Clients_Terra_Call {
	return &Clients_Terra_Call{Call: _e.mock.On("Terra")}
}

func (_c *Clients_Terra_Call) Run(run func()) *Clients_Terra_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Clients_Terra_Call) Return(_a0 terraapi.TerraClient, _a1 error) *Clients_Terra_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Clients_Terra_Call) RunAndReturn(run func() (terraapi.TerraClient, error)) *Clients_Terra_Call {
	_c.Call.Return(run)
	return _c
}

// TokenSource provides a mock function with given fields:
func (_m *Clients) TokenSource() (oauth2.TokenSource, error) {
	ret := _m.Called()

	var r0 oauth2.TokenSource
	var r1 error
	if rf, ok := ret.Get(0).(func() (oauth2.TokenSource, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() oauth2.TokenSource); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(oauth2.TokenSource)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Clients_TokenSource_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TokenSource'
type Clients_TokenSource_Call struct {
	*mock.Call
}

// TokenSource is a helper method to define mock.On call
func (_e *Clients_Expecter) TokenSource() *Clients_TokenSource_Call {
	return &Clients_TokenSource_Call{Call: _e.mock.On("TokenSource")}
}

func (_c *Clients_TokenSource_Call) Run(run func()) *Clients_TokenSource_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Clients_TokenSource_Call) Return(_a0 oauth2.TokenSource, _a1 error) *Clients_TokenSource_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Clients_TokenSource_Call) RunAndReturn(run func() (oauth2.TokenSource, error)) *Clients_TokenSource_Call {
	_c.Call.Return(run)
	return _c
}

// NewClients creates a new instance of Clients. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClients(t interface {
	mock.TestingT
	Cleanup(func())
}) *Clients {
	mock := &Clients{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
