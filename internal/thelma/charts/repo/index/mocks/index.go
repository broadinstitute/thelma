// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Index is an autogenerated mock type for the Index type
type Index struct {
	mock.Mock
}

type Index_Expecter struct {
	mock *mock.Mock
}

func (_m *Index) EXPECT() *Index_Expecter {
	return &Index_Expecter{mock: &_m.Mock}
}

// HasVersion provides a mock function with given fields: chartName, version
func (_m *Index) HasVersion(chartName string, version string) bool {
	ret := _m.Called(chartName, version)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(chartName, version)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Index_HasVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasVersion'
type Index_HasVersion_Call struct {
	*mock.Call
}

// HasVersion is a helper method to define mock.On call
//   - chartName string
//   - version string
func (_e *Index_Expecter) HasVersion(chartName interface{}, version interface{}) *Index_HasVersion_Call {
	return &Index_HasVersion_Call{Call: _e.mock.On("HasVersion", chartName, version)}
}

func (_c *Index_HasVersion_Call) Run(run func(chartName string, version string)) *Index_HasVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Index_HasVersion_Call) Return(_a0 bool) *Index_HasVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Index_HasVersion_Call) RunAndReturn(run func(string, string) bool) *Index_HasVersion_Call {
	_c.Call.Return(run)
	return _c
}

// MostRecentVersion provides a mock function with given fields: chartName
func (_m *Index) MostRecentVersion(chartName string) string {
	ret := _m.Called(chartName)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(chartName)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Index_MostRecentVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MostRecentVersion'
type Index_MostRecentVersion_Call struct {
	*mock.Call
}

// MostRecentVersion is a helper method to define mock.On call
//   - chartName string
func (_e *Index_Expecter) MostRecentVersion(chartName interface{}) *Index_MostRecentVersion_Call {
	return &Index_MostRecentVersion_Call{Call: _e.mock.On("MostRecentVersion", chartName)}
}

func (_c *Index_MostRecentVersion_Call) Run(run func(chartName string)) *Index_MostRecentVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Index_MostRecentVersion_Call) Return(_a0 string) *Index_MostRecentVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Index_MostRecentVersion_Call) RunAndReturn(run func(string) string) *Index_MostRecentVersion_Call {
	_c.Call.Return(run)
	return _c
}

// Versions provides a mock function with given fields: chartName
func (_m *Index) Versions(chartName string) []string {
	ret := _m.Called(chartName)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(chartName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Index_Versions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Versions'
type Index_Versions_Call struct {
	*mock.Call
}

// Versions is a helper method to define mock.On call
//   - chartName string
func (_e *Index_Expecter) Versions(chartName interface{}) *Index_Versions_Call {
	return &Index_Versions_Call{Call: _e.mock.On("Versions", chartName)}
}

func (_c *Index_Versions_Call) Run(run func(chartName string)) *Index_Versions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Index_Versions_Call) Return(_a0 []string) *Index_Versions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Index_Versions_Call) RunAndReturn(run func(string) []string) *Index_Versions_Call {
	_c.Call.Return(run)
	return _c
}

// NewIndex creates a new instance of Index. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIndex(t interface {
	mock.TestingT
	Cleanup(func())
}) *Index {
	mock := &Index{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
