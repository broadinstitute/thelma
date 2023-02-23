// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Bootstrapper is an autogenerated mock type for the Bootstrapper type
type Bootstrapper struct {
	mock.Mock
}

type Bootstrapper_Expecter struct {
	mock *mock.Mock
}

func (_m *Bootstrapper) EXPECT() *Bootstrapper_Expecter {
	return &Bootstrapper_Expecter{mock: &_m.Mock}
}

// Bootstrap provides a mock function with given fields:
func (_m *Bootstrapper) Bootstrap() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Bootstrapper_Bootstrap_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Bootstrap'
type Bootstrapper_Bootstrap_Call struct {
	*mock.Call
}

// Bootstrap is a helper method to define mock.On call
func (_e *Bootstrapper_Expecter) Bootstrap() *Bootstrapper_Bootstrap_Call {
	return &Bootstrapper_Bootstrap_Call{Call: _e.mock.On("Bootstrap")}
}

func (_c *Bootstrapper_Bootstrap_Call) Run(run func()) *Bootstrapper_Bootstrap_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Bootstrapper_Bootstrap_Call) Return(_a0 error) *Bootstrapper_Bootstrap_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewBootstrapper interface {
	mock.TestingT
	Cleanup(func())
}

// NewBootstrapper creates a new instance of Bootstrapper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewBootstrapper(t mockConstructorTestingTNewBootstrapper) *Bootstrapper {
	mock := &Bootstrapper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
