// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	index "github.com/broadinstitute/thelma/internal/thelma/charts/repo/index"
	mock "github.com/stretchr/testify/mock"
)

// Publisher is an autogenerated mock type for the Publisher type
type Publisher struct {
	mock.Mock
}

type Publisher_Expecter struct {
	mock *mock.Mock
}

func (_m *Publisher) EXPECT() *Publisher_Expecter {
	return &Publisher_Expecter{mock: &_m.Mock}
}

// ChartDir provides a mock function with given fields:
func (_m *Publisher) ChartDir() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Publisher_ChartDir_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChartDir'
type Publisher_ChartDir_Call struct {
	*mock.Call
}

// ChartDir is a helper method to define mock.On call
func (_e *Publisher_Expecter) ChartDir() *Publisher_ChartDir_Call {
	return &Publisher_ChartDir_Call{Call: _e.mock.On("ChartDir")}
}

func (_c *Publisher_ChartDir_Call) Run(run func()) *Publisher_ChartDir_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Publisher_ChartDir_Call) Return(_a0 string) *Publisher_ChartDir_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Publisher_ChartDir_Call) RunAndReturn(run func() string) *Publisher_ChartDir_Call {
	_c.Call.Return(run)
	return _c
}

// Close provides a mock function with given fields:
func (_m *Publisher) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Publisher_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type Publisher_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *Publisher_Expecter) Close() *Publisher_Close_Call {
	return &Publisher_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *Publisher_Close_Call) Run(run func()) *Publisher_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Publisher_Close_Call) Return(_a0 error) *Publisher_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Publisher_Close_Call) RunAndReturn(run func() error) *Publisher_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Index provides a mock function with given fields:
func (_m *Publisher) Index() index.Index {
	ret := _m.Called()

	var r0 index.Index
	if rf, ok := ret.Get(0).(func() index.Index); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(index.Index)
		}
	}

	return r0
}

// Publisher_Index_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Index'
type Publisher_Index_Call struct {
	*mock.Call
}

// Index is a helper method to define mock.On call
func (_e *Publisher_Expecter) Index() *Publisher_Index_Call {
	return &Publisher_Index_Call{Call: _e.mock.On("Index")}
}

func (_c *Publisher_Index_Call) Run(run func()) *Publisher_Index_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Publisher_Index_Call) Return(_a0 index.Index) *Publisher_Index_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Publisher_Index_Call) RunAndReturn(run func() index.Index) *Publisher_Index_Call {
	_c.Call.Return(run)
	return _c
}

// Publish provides a mock function with given fields:
func (_m *Publisher) Publish() (int, error) {
	ret := _m.Called()

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func() (int, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Publisher_Publish_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Publish'
type Publisher_Publish_Call struct {
	*mock.Call
}

// Publish is a helper method to define mock.On call
func (_e *Publisher_Expecter) Publish() *Publisher_Publish_Call {
	return &Publisher_Publish_Call{Call: _e.mock.On("Publish")}
}

func (_c *Publisher_Publish_Call) Run(run func()) *Publisher_Publish_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Publisher_Publish_Call) Return(count int, err error) *Publisher_Publish_Call {
	_c.Call.Return(count, err)
	return _c
}

func (_c *Publisher_Publish_Call) RunAndReturn(run func() (int, error)) *Publisher_Publish_Call {
	_c.Call.Return(run)
	return _c
}

// NewPublisher creates a new instance of Publisher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPublisher(t interface {
	mock.TestingT
	Cleanup(func())
}) *Publisher {
	mock := &Publisher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
