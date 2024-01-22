// Code generated by mockery v2.40.1. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	watch "k8s.io/apimachinery/pkg/watch"
)

// Watch is an autogenerated mock type for the Watch type
type Watch struct {
	mock.Mock
}

type Watch_Expecter struct {
	mock *mock.Mock
}

func (_m *Watch) EXPECT() *Watch_Expecter {
	return &Watch_Expecter{mock: &_m.Mock}
}

// ResultChan provides a mock function with given fields:
func (_m *Watch) ResultChan() <-chan watch.Event {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ResultChan")
	}

	var r0 <-chan watch.Event
	if rf, ok := ret.Get(0).(func() <-chan watch.Event); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan watch.Event)
		}
	}

	return r0
}

// Watch_ResultChan_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResultChan'
type Watch_ResultChan_Call struct {
	*mock.Call
}

// ResultChan is a helper method to define mock.On call
func (_e *Watch_Expecter) ResultChan() *Watch_ResultChan_Call {
	return &Watch_ResultChan_Call{Call: _e.mock.On("ResultChan")}
}

func (_c *Watch_ResultChan_Call) Run(run func()) *Watch_ResultChan_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Watch_ResultChan_Call) Return(_a0 <-chan watch.Event) *Watch_ResultChan_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Watch_ResultChan_Call) RunAndReturn(run func() <-chan watch.Event) *Watch_ResultChan_Call {
	_c.Call.Return(run)
	return _c
}

// Stop provides a mock function with given fields:
func (_m *Watch) Stop() {
	_m.Called()
}

// Watch_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type Watch_Stop_Call struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
func (_e *Watch_Expecter) Stop() *Watch_Stop_Call {
	return &Watch_Stop_Call{Call: _e.mock.On("Stop")}
}

func (_c *Watch_Stop_Call) Run(run func()) *Watch_Stop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Watch_Stop_Call) Return() *Watch_Stop_Call {
	_c.Call.Return()
	return _c
}

func (_c *Watch_Stop_Call) RunAndReturn(run func()) *Watch_Stop_Call {
	_c.Call.Return(run)
	return _c
}

// NewWatch creates a new instance of Watch. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWatch(t interface {
	mock.TestingT
	Cleanup(func())
}) *Watch {
	mock := &Watch{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
