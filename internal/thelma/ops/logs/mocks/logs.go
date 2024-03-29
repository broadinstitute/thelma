// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	artifacts "github.com/broadinstitute/thelma/internal/thelma/ops/artifacts"
	logs "github.com/broadinstitute/thelma/internal/thelma/ops/logs"

	mock "github.com/stretchr/testify/mock"

	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

// Logs is an autogenerated mock type for the Logs type
type Logs struct {
	mock.Mock
}

type Logs_Expecter struct {
	mock *mock.Mock
}

func (_m *Logs) EXPECT() *Logs_Expecter {
	return &Logs_Expecter{mock: &_m.Mock}
}

// Export provides a mock function with given fields: releases, option
func (_m *Logs) Export(releases []terra.Release, option ...logs.ExportOption) (map[terra.Release]artifacts.Location, error) {
	_va := make([]interface{}, len(option))
	for _i := range option {
		_va[_i] = option[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, releases)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 map[terra.Release]artifacts.Location
	var r1 error
	if rf, ok := ret.Get(0).(func([]terra.Release, ...logs.ExportOption) (map[terra.Release]artifacts.Location, error)); ok {
		return rf(releases, option...)
	}
	if rf, ok := ret.Get(0).(func([]terra.Release, ...logs.ExportOption) map[terra.Release]artifacts.Location); ok {
		r0 = rf(releases, option...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[terra.Release]artifacts.Location)
		}
	}

	if rf, ok := ret.Get(1).(func([]terra.Release, ...logs.ExportOption) error); ok {
		r1 = rf(releases, option...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Logs_Export_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Export'
type Logs_Export_Call struct {
	*mock.Call
}

// Export is a helper method to define mock.On call
//   - releases []terra.Release
//   - option ...logs.ExportOption
func (_e *Logs_Expecter) Export(releases interface{}, option ...interface{}) *Logs_Export_Call {
	return &Logs_Export_Call{Call: _e.mock.On("Export",
		append([]interface{}{releases}, option...)...)}
}

func (_c *Logs_Export_Call) Run(run func(releases []terra.Release, option ...logs.ExportOption)) *Logs_Export_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]logs.ExportOption, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(logs.ExportOption)
			}
		}
		run(args[0].([]terra.Release), variadicArgs...)
	})
	return _c
}

func (_c *Logs_Export_Call) Return(_a0 map[terra.Release]artifacts.Location, _a1 error) *Logs_Export_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Logs_Export_Call) RunAndReturn(run func([]terra.Release, ...logs.ExportOption) (map[terra.Release]artifacts.Location, error)) *Logs_Export_Call {
	_c.Call.Return(run)
	return _c
}

// Logs provides a mock function with given fields: release, option
func (_m *Logs) Logs(release terra.Release, option ...logs.LogsOption) error {
	_va := make([]interface{}, len(option))
	for _i := range option {
		_va[_i] = option[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, release)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(terra.Release, ...logs.LogsOption) error); ok {
		r0 = rf(release, option...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Logs_Logs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Logs'
type Logs_Logs_Call struct {
	*mock.Call
}

// Logs is a helper method to define mock.On call
//   - release terra.Release
//   - option ...logs.LogsOption
func (_e *Logs_Expecter) Logs(release interface{}, option ...interface{}) *Logs_Logs_Call {
	return &Logs_Logs_Call{Call: _e.mock.On("Logs",
		append([]interface{}{release}, option...)...)}
}

func (_c *Logs_Logs_Call) Run(run func(release terra.Release, option ...logs.LogsOption)) *Logs_Logs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]logs.LogsOption, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(logs.LogsOption)
			}
		}
		run(args[0].(terra.Release), variadicArgs...)
	})
	return _c
}

func (_c *Logs_Logs_Call) Return(_a0 error) *Logs_Logs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Logs_Logs_Call) RunAndReturn(run func(terra.Release, ...logs.LogsOption) error) *Logs_Logs_Call {
	_c.Call.Return(run)
	return _c
}

// NewLogs creates a new instance of Logs. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLogs(t interface {
	mock.TestingT
	Cleanup(func())
}) *Logs {
	mock := &Logs{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
