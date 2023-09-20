// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	mock "github.com/stretchr/testify/mock"
)

// StateWriter is an autogenerated mock type for the StateWriter type
type StateWriter struct {
	mock.Mock
}

type StateWriter_Expecter struct {
	mock *mock.Mock
}

func (_m *StateWriter) EXPECT() *StateWriter_Expecter {
	return &StateWriter_Expecter{mock: &_m.Mock}
}

// DeleteEnvironments provides a mock function with given fields: _a0
func (_m *StateWriter) DeleteEnvironments(_a0 []terra.Environment) ([]string, error) {
	ret := _m.Called(_a0)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func([]terra.Environment) ([]string, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func([]terra.Environment) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func([]terra.Environment) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StateWriter_DeleteEnvironments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteEnvironments'
type StateWriter_DeleteEnvironments_Call struct {
	*mock.Call
}

// DeleteEnvironments is a helper method to define mock.On call
//   - _a0 []terra.Environment
func (_e *StateWriter_Expecter) DeleteEnvironments(_a0 interface{}) *StateWriter_DeleteEnvironments_Call {
	return &StateWriter_DeleteEnvironments_Call{Call: _e.mock.On("DeleteEnvironments", _a0)}
}

func (_c *StateWriter_DeleteEnvironments_Call) Run(run func(_a0 []terra.Environment)) *StateWriter_DeleteEnvironments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]terra.Environment))
	})
	return _c
}

func (_c *StateWriter_DeleteEnvironments_Call) Return(_a0 []string, _a1 error) *StateWriter_DeleteEnvironments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StateWriter_DeleteEnvironments_Call) RunAndReturn(run func([]terra.Environment) ([]string, error)) *StateWriter_DeleteEnvironments_Call {
	_c.Call.Return(run)
	return _c
}

// DisableRelease provides a mock function with given fields: _a0, _a1
func (_m *StateWriter) DisableRelease(_a0 string, _a1 string) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StateWriter_DisableRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DisableRelease'
type StateWriter_DisableRelease_Call struct {
	*mock.Call
}

// DisableRelease is a helper method to define mock.On call
//   - _a0 string
//   - _a1 string
func (_e *StateWriter_Expecter) DisableRelease(_a0 interface{}, _a1 interface{}) *StateWriter_DisableRelease_Call {
	return &StateWriter_DisableRelease_Call{Call: _e.mock.On("DisableRelease", _a0, _a1)}
}

func (_c *StateWriter_DisableRelease_Call) Run(run func(_a0 string, _a1 string)) *StateWriter_DisableRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *StateWriter_DisableRelease_Call) Return(_a0 error) *StateWriter_DisableRelease_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *StateWriter_DisableRelease_Call) RunAndReturn(run func(string, string) error) *StateWriter_DisableRelease_Call {
	_c.Call.Return(run)
	return _c
}

// EnableRelease provides a mock function with given fields: _a0, _a1
func (_m *StateWriter) EnableRelease(_a0 terra.Environment, _a1 string) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(terra.Environment, string) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StateWriter_EnableRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EnableRelease'
type StateWriter_EnableRelease_Call struct {
	*mock.Call
}

// EnableRelease is a helper method to define mock.On call
//   - _a0 terra.Environment
//   - _a1 string
func (_e *StateWriter_Expecter) EnableRelease(_a0 interface{}, _a1 interface{}) *StateWriter_EnableRelease_Call {
	return &StateWriter_EnableRelease_Call{Call: _e.mock.On("EnableRelease", _a0, _a1)}
}

func (_c *StateWriter_EnableRelease_Call) Run(run func(_a0 terra.Environment, _a1 string)) *StateWriter_EnableRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(terra.Environment), args[1].(string))
	})
	return _c
}

func (_c *StateWriter_EnableRelease_Call) Return(_a0 error) *StateWriter_EnableRelease_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *StateWriter_EnableRelease_Call) RunAndReturn(run func(terra.Environment, string) error) *StateWriter_EnableRelease_Call {
	_c.Call.Return(run)
	return _c
}

// WriteClusters provides a mock function with given fields: _a0
func (_m *StateWriter) WriteClusters(_a0 []terra.Cluster) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func([]terra.Cluster) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StateWriter_WriteClusters_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteClusters'
type StateWriter_WriteClusters_Call struct {
	*mock.Call
}

// WriteClusters is a helper method to define mock.On call
//   - _a0 []terra.Cluster
func (_e *StateWriter_Expecter) WriteClusters(_a0 interface{}) *StateWriter_WriteClusters_Call {
	return &StateWriter_WriteClusters_Call{Call: _e.mock.On("WriteClusters", _a0)}
}

func (_c *StateWriter_WriteClusters_Call) Run(run func(_a0 []terra.Cluster)) *StateWriter_WriteClusters_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]terra.Cluster))
	})
	return _c
}

func (_c *StateWriter_WriteClusters_Call) Return(_a0 error) *StateWriter_WriteClusters_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *StateWriter_WriteClusters_Call) RunAndReturn(run func([]terra.Cluster) error) *StateWriter_WriteClusters_Call {
	_c.Call.Return(run)
	return _c
}

// WriteEnvironments provides a mock function with given fields: _a0
func (_m *StateWriter) WriteEnvironments(_a0 []terra.Environment) ([]string, error) {
	ret := _m.Called(_a0)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func([]terra.Environment) ([]string, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func([]terra.Environment) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func([]terra.Environment) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StateWriter_WriteEnvironments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteEnvironments'
type StateWriter_WriteEnvironments_Call struct {
	*mock.Call
}

// WriteEnvironments is a helper method to define mock.On call
//   - _a0 []terra.Environment
func (_e *StateWriter_Expecter) WriteEnvironments(_a0 interface{}) *StateWriter_WriteEnvironments_Call {
	return &StateWriter_WriteEnvironments_Call{Call: _e.mock.On("WriteEnvironments", _a0)}
}

func (_c *StateWriter_WriteEnvironments_Call) Run(run func(_a0 []terra.Environment)) *StateWriter_WriteEnvironments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]terra.Environment))
	})
	return _c
}

func (_c *StateWriter_WriteEnvironments_Call) Return(_a0 []string, _a1 error) *StateWriter_WriteEnvironments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StateWriter_WriteEnvironments_Call) RunAndReturn(run func([]terra.Environment) ([]string, error)) *StateWriter_WriteEnvironments_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewStateWriter interface {
	mock.TestingT
	Cleanup(func())
}

// NewStateWriter creates a new instance of StateWriter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewStateWriter(t mockConstructorTestingTNewStateWriter) *StateWriter {
	mock := &StateWriter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
