// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	api "github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	dbms "github.com/broadinstitute/thelma/internal/thelma/ops/sql/dbms"

	mock "github.com/stretchr/testify/mock"

	podrun "github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"
)

// DBMS is an autogenerated mock type for the DBMS type
type DBMS struct {
	mock.Mock
}

type DBMS_Expecter struct {
	mock *mock.Mock
}

func (_m *DBMS) EXPECT() *DBMS_Expecter {
	return &DBMS_Expecter{mock: &_m.Mock}
}

// InitCommand provides a mock function with given fields:
func (_m *DBMS) InitCommand() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// DBMS_InitCommand_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InitCommand'
type DBMS_InitCommand_Call struct {
	*mock.Call
}

// InitCommand is a helper method to define mock.On call
func (_e *DBMS_Expecter) InitCommand() *DBMS_InitCommand_Call {
	return &DBMS_InitCommand_Call{Call: _e.mock.On("InitCommand")}
}

func (_c *DBMS_InitCommand_Call) Run(run func()) *DBMS_InitCommand_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DBMS_InitCommand_Call) Return(_a0 []string) *DBMS_InitCommand_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DBMS_InitCommand_Call) RunAndReturn(run func() []string) *DBMS_InitCommand_Call {
	_c.Call.Return(run)
	return _c
}

// PodSpec provides a mock function with given fields: _a0
func (_m *DBMS) PodSpec(_a0 dbms.ClientSettings) (podrun.DBMSSpec, error) {
	ret := _m.Called(_a0)

	var r0 podrun.DBMSSpec
	var r1 error
	if rf, ok := ret.Get(0).(func(dbms.ClientSettings) (podrun.DBMSSpec, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(dbms.ClientSettings) podrun.DBMSSpec); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(podrun.DBMSSpec)
	}

	if rf, ok := ret.Get(1).(func(dbms.ClientSettings) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DBMS_PodSpec_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PodSpec'
type DBMS_PodSpec_Call struct {
	*mock.Call
}

// PodSpec is a helper method to define mock.On call
//   - _a0 dbms.ClientSettings
func (_e *DBMS_Expecter) PodSpec(_a0 interface{}) *DBMS_PodSpec_Call {
	return &DBMS_PodSpec_Call{Call: _e.mock.On("PodSpec", _a0)}
}

func (_c *DBMS_PodSpec_Call) Run(run func(_a0 dbms.ClientSettings)) *DBMS_PodSpec_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(dbms.ClientSettings))
	})
	return _c
}

func (_c *DBMS_PodSpec_Call) Return(_a0 podrun.DBMSSpec, _a1 error) *DBMS_PodSpec_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DBMS_PodSpec_Call) RunAndReturn(run func(dbms.ClientSettings) (podrun.DBMSSpec, error)) *DBMS_PodSpec_Call {
	_c.Call.Return(run)
	return _c
}

// ShellCommand provides a mock function with given fields:
func (_m *DBMS) ShellCommand() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// DBMS_ShellCommand_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ShellCommand'
type DBMS_ShellCommand_Call struct {
	*mock.Call
}

// ShellCommand is a helper method to define mock.On call
func (_e *DBMS_Expecter) ShellCommand() *DBMS_ShellCommand_Call {
	return &DBMS_ShellCommand_Call{Call: _e.mock.On("ShellCommand")}
}

func (_c *DBMS_ShellCommand_Call) Run(run func()) *DBMS_ShellCommand_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DBMS_ShellCommand_Call) Return(_a0 []string) *DBMS_ShellCommand_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DBMS_ShellCommand_Call) RunAndReturn(run func() []string) *DBMS_ShellCommand_Call {
	_c.Call.Return(run)
	return _c
}

// Type provides a mock function with given fields:
func (_m *DBMS) Type() api.DBMS {
	ret := _m.Called()

	var r0 api.DBMS
	if rf, ok := ret.Get(0).(func() api.DBMS); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(api.DBMS)
	}

	return r0
}

// DBMS_Type_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Type'
type DBMS_Type_Call struct {
	*mock.Call
}

// Type is a helper method to define mock.On call
func (_e *DBMS_Expecter) Type() *DBMS_Type_Call {
	return &DBMS_Type_Call{Call: _e.mock.On("Type")}
}

func (_c *DBMS_Type_Call) Run(run func()) *DBMS_Type_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DBMS_Type_Call) Return(_a0 api.DBMS) *DBMS_Type_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DBMS_Type_Call) RunAndReturn(run func() api.DBMS) *DBMS_Type_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewDBMS interface {
	mock.TestingT
	Cleanup(func())
}

// NewDBMS creates a new instance of DBMS. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDBMS(t mockConstructorTestingTNewDBMS) *DBMS {
	mock := &DBMS{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
