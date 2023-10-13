// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Kubectx is an autogenerated mock type for the Kubectx type
type Kubectx struct {
	mock.Mock
}

type Kubectx_Expecter struct {
	mock *mock.Mock
}

func (_m *Kubectx) EXPECT() *Kubectx_Expecter {
	return &Kubectx_Expecter{mock: &_m.Mock}
}

// ContextName provides a mock function with given fields:
func (_m *Kubectx) ContextName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Kubectx_ContextName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ContextName'
type Kubectx_ContextName_Call struct {
	*mock.Call
}

// ContextName is a helper method to define mock.On call
func (_e *Kubectx_Expecter) ContextName() *Kubectx_ContextName_Call {
	return &Kubectx_ContextName_Call{Call: _e.mock.On("ContextName")}
}

func (_c *Kubectx_ContextName_Call) Run(run func()) *Kubectx_ContextName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Kubectx_ContextName_Call) Return(_a0 string) *Kubectx_ContextName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Kubectx_ContextName_Call) RunAndReturn(run func() string) *Kubectx_ContextName_Call {
	_c.Call.Return(run)
	return _c
}

// Namespace provides a mock function with given fields:
func (_m *Kubectx) Namespace() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Kubectx_Namespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Namespace'
type Kubectx_Namespace_Call struct {
	*mock.Call
}

// Namespace is a helper method to define mock.On call
func (_e *Kubectx_Expecter) Namespace() *Kubectx_Namespace_Call {
	return &Kubectx_Namespace_Call{Call: _e.mock.On("Namespace")}
}

func (_c *Kubectx_Namespace_Call) Run(run func()) *Kubectx_Namespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Kubectx_Namespace_Call) Return(_a0 string) *Kubectx_Namespace_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Kubectx_Namespace_Call) RunAndReturn(run func() string) *Kubectx_Namespace_Call {
	_c.Call.Return(run)
	return _c
}

// NewKubectx creates a new instance of Kubectx. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewKubectx(t interface {
	mock.TestingT
	Cleanup(func())
}) *Kubectx {
	mock := &Kubectx{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
