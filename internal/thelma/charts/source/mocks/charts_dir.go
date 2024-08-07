// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	source "github.com/broadinstitute/thelma/internal/thelma/charts/source"
	mock "github.com/stretchr/testify/mock"
)

// ChartsDir is an autogenerated mock type for the ChartsDir type
type ChartsDir struct {
	mock.Mock
}

type ChartsDir_Expecter struct {
	mock *mock.Mock
}

func (_m *ChartsDir) EXPECT() *ChartsDir_Expecter {
	return &ChartsDir_Expecter{mock: &_m.Mock}
}

// Exists provides a mock function with given fields: name
func (_m *ChartsDir) Exists(name string) bool {
	ret := _m.Called(name)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ChartsDir_Exists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Exists'
type ChartsDir_Exists_Call struct {
	*mock.Call
}

// Exists is a helper method to define mock.On call
//   - name string
func (_e *ChartsDir_Expecter) Exists(name interface{}) *ChartsDir_Exists_Call {
	return &ChartsDir_Exists_Call{Call: _e.mock.On("Exists", name)}
}

func (_c *ChartsDir_Exists_Call) Run(run func(name string)) *ChartsDir_Exists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ChartsDir_Exists_Call) Return(_a0 bool) *ChartsDir_Exists_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ChartsDir_Exists_Call) RunAndReturn(run func(string) bool) *ChartsDir_Exists_Call {
	_c.Call.Return(run)
	return _c
}

// GetChart provides a mock function with given fields: name
func (_m *ChartsDir) GetChart(name string) (source.Chart, error) {
	ret := _m.Called(name)

	var r0 source.Chart
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (source.Chart, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) source.Chart); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(source.Chart)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChartsDir_GetChart_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetChart'
type ChartsDir_GetChart_Call struct {
	*mock.Call
}

// GetChart is a helper method to define mock.On call
//   - name string
func (_e *ChartsDir_Expecter) GetChart(name interface{}) *ChartsDir_GetChart_Call {
	return &ChartsDir_GetChart_Call{Call: _e.mock.On("GetChart", name)}
}

func (_c *ChartsDir_GetChart_Call) Run(run func(name string)) *ChartsDir_GetChart_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ChartsDir_GetChart_Call) Return(_a0 source.Chart, _a1 error) *ChartsDir_GetChart_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ChartsDir_GetChart_Call) RunAndReturn(run func(string) (source.Chart, error)) *ChartsDir_GetChart_Call {
	_c.Call.Return(run)
	return _c
}

// GetCharts provides a mock function with given fields: name
func (_m *ChartsDir) GetCharts(name ...string) ([]source.Chart, error) {
	_va := make([]interface{}, len(name))
	for _i := range name {
		_va[_i] = name[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []source.Chart
	var r1 error
	if rf, ok := ret.Get(0).(func(...string) ([]source.Chart, error)); ok {
		return rf(name...)
	}
	if rf, ok := ret.Get(0).(func(...string) []source.Chart); ok {
		r0 = rf(name...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]source.Chart)
		}
	}

	if rf, ok := ret.Get(1).(func(...string) error); ok {
		r1 = rf(name...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChartsDir_GetCharts_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCharts'
type ChartsDir_GetCharts_Call struct {
	*mock.Call
}

// GetCharts is a helper method to define mock.On call
//   - name ...string
func (_e *ChartsDir_Expecter) GetCharts(name ...interface{}) *ChartsDir_GetCharts_Call {
	return &ChartsDir_GetCharts_Call{Call: _e.mock.On("GetCharts",
		append([]interface{}{}, name...)...)}
}

func (_c *ChartsDir_GetCharts_Call) Run(run func(name ...string)) *ChartsDir_GetCharts_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *ChartsDir_GetCharts_Call) Return(_a0 []source.Chart, _a1 error) *ChartsDir_GetCharts_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ChartsDir_GetCharts_Call) RunAndReturn(run func(...string) ([]source.Chart, error)) *ChartsDir_GetCharts_Call {
	_c.Call.Return(run)
	return _c
}

// Path provides a mock function with given fields:
func (_m *ChartsDir) Path() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ChartsDir_Path_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Path'
type ChartsDir_Path_Call struct {
	*mock.Call
}

// Path is a helper method to define mock.On call
func (_e *ChartsDir_Expecter) Path() *ChartsDir_Path_Call {
	return &ChartsDir_Path_Call{Call: _e.mock.On("Path")}
}

func (_c *ChartsDir_Path_Call) Run(run func()) *ChartsDir_Path_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ChartsDir_Path_Call) Return(_a0 string) *ChartsDir_Path_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ChartsDir_Path_Call) RunAndReturn(run func() string) *ChartsDir_Path_Call {
	_c.Call.Return(run)
	return _c
}

// RecursivelyUpdateDependencies provides a mock function with given fields: chart
func (_m *ChartsDir) RecursivelyUpdateDependencies(chart ...source.Chart) error {
	_va := make([]interface{}, len(chart))
	for _i := range chart {
		_va[_i] = chart[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(...source.Chart) error); ok {
		r0 = rf(chart...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ChartsDir_RecursivelyUpdateDependencies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RecursivelyUpdateDependencies'
type ChartsDir_RecursivelyUpdateDependencies_Call struct {
	*mock.Call
}

// RecursivelyUpdateDependencies is a helper method to define mock.On call
//   - chart ...source.Chart
func (_e *ChartsDir_Expecter) RecursivelyUpdateDependencies(chart ...interface{}) *ChartsDir_RecursivelyUpdateDependencies_Call {
	return &ChartsDir_RecursivelyUpdateDependencies_Call{Call: _e.mock.On("RecursivelyUpdateDependencies",
		append([]interface{}{}, chart...)...)}
}

func (_c *ChartsDir_RecursivelyUpdateDependencies_Call) Run(run func(chart ...source.Chart)) *ChartsDir_RecursivelyUpdateDependencies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]source.Chart, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(source.Chart)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *ChartsDir_RecursivelyUpdateDependencies_Call) Return(_a0 error) *ChartsDir_RecursivelyUpdateDependencies_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ChartsDir_RecursivelyUpdateDependencies_Call) RunAndReturn(run func(...source.Chart) error) *ChartsDir_RecursivelyUpdateDependencies_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateDependentVersionConstraints provides a mock function with given fields: chart, newVersionConstraint
func (_m *ChartsDir) UpdateDependentVersionConstraints(chart source.Chart, newVersionConstraint string) error {
	ret := _m.Called(chart, newVersionConstraint)

	var r0 error
	if rf, ok := ret.Get(0).(func(source.Chart, string) error); ok {
		r0 = rf(chart, newVersionConstraint)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ChartsDir_UpdateDependentVersionConstraints_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateDependentVersionConstraints'
type ChartsDir_UpdateDependentVersionConstraints_Call struct {
	*mock.Call
}

// UpdateDependentVersionConstraints is a helper method to define mock.On call
//   - chart source.Chart
//   - newVersionConstraint string
func (_e *ChartsDir_Expecter) UpdateDependentVersionConstraints(chart interface{}, newVersionConstraint interface{}) *ChartsDir_UpdateDependentVersionConstraints_Call {
	return &ChartsDir_UpdateDependentVersionConstraints_Call{Call: _e.mock.On("UpdateDependentVersionConstraints", chart, newVersionConstraint)}
}

func (_c *ChartsDir_UpdateDependentVersionConstraints_Call) Run(run func(chart source.Chart, newVersionConstraint string)) *ChartsDir_UpdateDependentVersionConstraints_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(source.Chart), args[1].(string))
	})
	return _c
}

func (_c *ChartsDir_UpdateDependentVersionConstraints_Call) Return(_a0 error) *ChartsDir_UpdateDependentVersionConstraints_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ChartsDir_UpdateDependentVersionConstraints_Call) RunAndReturn(run func(source.Chart, string) error) *ChartsDir_UpdateDependentVersionConstraints_Call {
	_c.Call.Return(run)
	return _c
}

// WithTransitiveDependents provides a mock function with given fields: chart
func (_m *ChartsDir) WithTransitiveDependents(chart []source.Chart) ([]source.Chart, error) {
	ret := _m.Called(chart)

	var r0 []source.Chart
	var r1 error
	if rf, ok := ret.Get(0).(func([]source.Chart) ([]source.Chart, error)); ok {
		return rf(chart)
	}
	if rf, ok := ret.Get(0).(func([]source.Chart) []source.Chart); ok {
		r0 = rf(chart)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]source.Chart)
		}
	}

	if rf, ok := ret.Get(1).(func([]source.Chart) error); ok {
		r1 = rf(chart)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ChartsDir_WithTransitiveDependents_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithTransitiveDependents'
type ChartsDir_WithTransitiveDependents_Call struct {
	*mock.Call
}

// WithTransitiveDependents is a helper method to define mock.On call
//   - chart []source.Chart
func (_e *ChartsDir_Expecter) WithTransitiveDependents(chart interface{}) *ChartsDir_WithTransitiveDependents_Call {
	return &ChartsDir_WithTransitiveDependents_Call{Call: _e.mock.On("WithTransitiveDependents", chart)}
}

func (_c *ChartsDir_WithTransitiveDependents_Call) Run(run func(chart []source.Chart)) *ChartsDir_WithTransitiveDependents_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]source.Chart))
	})
	return _c
}

func (_c *ChartsDir_WithTransitiveDependents_Call) Return(_a0 []source.Chart, _a1 error) *ChartsDir_WithTransitiveDependents_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ChartsDir_WithTransitiveDependents_Call) RunAndReturn(run func([]source.Chart) ([]source.Chart, error)) *ChartsDir_WithTransitiveDependents_Call {
	_c.Call.Return(run)
	return _c
}

// NewChartsDir creates a new instance of ChartsDir. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewChartsDir(t interface {
	mock.TestingT
	Cleanup(func())
}) *ChartsDir {
	mock := &ChartsDir{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
