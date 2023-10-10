// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	mock "github.com/stretchr/testify/mock"
)

// Releases is an autogenerated mock type for the Releases type
type Releases struct {
	mock.Mock
}

type Releases_Expecter struct {
	mock *mock.Mock
}

func (_m *Releases) EXPECT() *Releases_Expecter {
	return &Releases_Expecter{mock: &_m.Mock}
}

// All provides a mock function with given fields:
func (_m *Releases) All() ([]terra.Release, error) {
	ret := _m.Called()

	var r0 []terra.Release
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]terra.Release, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []terra.Release); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]terra.Release)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Releases_All_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'All'
type Releases_All_Call struct {
	*mock.Call
}

// All is a helper method to define mock.On call
func (_e *Releases_Expecter) All() *Releases_All_Call {
	return &Releases_All_Call{Call: _e.mock.On("All")}
}

func (_c *Releases_All_Call) Run(run func()) *Releases_All_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Releases_All_Call) Return(_a0 []terra.Release, _a1 error) *Releases_All_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Releases_All_Call) RunAndReturn(run func() ([]terra.Release, error)) *Releases_All_Call {
	_c.Call.Return(run)
	return _c
}

// Filter provides a mock function with given fields: filter
func (_m *Releases) Filter(filter terra.ReleaseFilter) ([]terra.Release, error) {
	ret := _m.Called(filter)

	var r0 []terra.Release
	var r1 error
	if rf, ok := ret.Get(0).(func(terra.ReleaseFilter) ([]terra.Release, error)); ok {
		return rf(filter)
	}
	if rf, ok := ret.Get(0).(func(terra.ReleaseFilter) []terra.Release); ok {
		r0 = rf(filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]terra.Release)
		}
	}

	if rf, ok := ret.Get(1).(func(terra.ReleaseFilter) error); ok {
		r1 = rf(filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Releases_Filter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Filter'
type Releases_Filter_Call struct {
	*mock.Call
}

// Filter is a helper method to define mock.On call
//   - filter terra.ReleaseFilter
func (_e *Releases_Expecter) Filter(filter interface{}) *Releases_Filter_Call {
	return &Releases_Filter_Call{Call: _e.mock.On("Filter", filter)}
}

func (_c *Releases_Filter_Call) Run(run func(filter terra.ReleaseFilter)) *Releases_Filter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(terra.ReleaseFilter))
	})
	return _c
}

func (_c *Releases_Filter_Call) Return(_a0 []terra.Release, _a1 error) *Releases_Filter_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Releases_Filter_Call) RunAndReturn(run func(terra.ReleaseFilter) ([]terra.Release, error)) *Releases_Filter_Call {
	_c.Call.Return(run)
	return _c
}

// NewReleases creates a new instance of Releases. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewReleases(t interface {
	mock.TestingT
	Cleanup(func())
}) *Releases {
	mock := &Releases{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
