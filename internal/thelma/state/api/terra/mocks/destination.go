// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	mock "github.com/stretchr/testify/mock"
)

// Destination is an autogenerated mock type for the Destination type
type Destination struct {
	mock.Mock
}

type Destination_Expecter struct {
	mock *mock.Mock
}

func (_m *Destination) EXPECT() *Destination_Expecter {
	return &Destination_Expecter{mock: &_m.Mock}
}

// Base provides a mock function with given fields:
func (_m *Destination) Base() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Destination_Base_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Base'
type Destination_Base_Call struct {
	*mock.Call
}

// Base is a helper method to define mock.On call
func (_e *Destination_Expecter) Base() *Destination_Base_Call {
	return &Destination_Base_Call{Call: _e.mock.On("Base")}
}

func (_c *Destination_Base_Call) Run(run func()) *Destination_Base_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Destination_Base_Call) Return(_a0 string) *Destination_Base_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Destination_Base_Call) RunAndReturn(run func() string) *Destination_Base_Call {
	_c.Call.Return(run)
	return _c
}

// IsCluster provides a mock function with given fields:
func (_m *Destination) IsCluster() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Destination_IsCluster_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsCluster'
type Destination_IsCluster_Call struct {
	*mock.Call
}

// IsCluster is a helper method to define mock.On call
func (_e *Destination_Expecter) IsCluster() *Destination_IsCluster_Call {
	return &Destination_IsCluster_Call{Call: _e.mock.On("IsCluster")}
}

func (_c *Destination_IsCluster_Call) Run(run func()) *Destination_IsCluster_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Destination_IsCluster_Call) Return(_a0 bool) *Destination_IsCluster_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Destination_IsCluster_Call) RunAndReturn(run func() bool) *Destination_IsCluster_Call {
	_c.Call.Return(run)
	return _c
}

// IsEnvironment provides a mock function with given fields:
func (_m *Destination) IsEnvironment() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Destination_IsEnvironment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsEnvironment'
type Destination_IsEnvironment_Call struct {
	*mock.Call
}

// IsEnvironment is a helper method to define mock.On call
func (_e *Destination_Expecter) IsEnvironment() *Destination_IsEnvironment_Call {
	return &Destination_IsEnvironment_Call{Call: _e.mock.On("IsEnvironment")}
}

func (_c *Destination_IsEnvironment_Call) Run(run func()) *Destination_IsEnvironment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Destination_IsEnvironment_Call) Return(_a0 bool) *Destination_IsEnvironment_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Destination_IsEnvironment_Call) RunAndReturn(run func() bool) *Destination_IsEnvironment_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with given fields:
func (_m *Destination) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Destination_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type Destination_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *Destination_Expecter) Name() *Destination_Name_Call {
	return &Destination_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *Destination_Name_Call) Run(run func()) *Destination_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Destination_Name_Call) Return(_a0 string) *Destination_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Destination_Name_Call) RunAndReturn(run func() string) *Destination_Name_Call {
	_c.Call.Return(run)
	return _c
}

// ReleaseType provides a mock function with given fields:
func (_m *Destination) ReleaseType() terra.ReleaseType {
	ret := _m.Called()

	var r0 terra.ReleaseType
	if rf, ok := ret.Get(0).(func() terra.ReleaseType); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(terra.ReleaseType)
	}

	return r0
}

// Destination_ReleaseType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReleaseType'
type Destination_ReleaseType_Call struct {
	*mock.Call
}

// ReleaseType is a helper method to define mock.On call
func (_e *Destination_Expecter) ReleaseType() *Destination_ReleaseType_Call {
	return &Destination_ReleaseType_Call{Call: _e.mock.On("ReleaseType")}
}

func (_c *Destination_ReleaseType_Call) Run(run func()) *Destination_ReleaseType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Destination_ReleaseType_Call) Return(_a0 terra.ReleaseType) *Destination_ReleaseType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Destination_ReleaseType_Call) RunAndReturn(run func() terra.ReleaseType) *Destination_ReleaseType_Call {
	_c.Call.Return(run)
	return _c
}

// Releases provides a mock function with given fields:
func (_m *Destination) Releases() []terra.Release {
	ret := _m.Called()

	var r0 []terra.Release
	if rf, ok := ret.Get(0).(func() []terra.Release); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]terra.Release)
		}
	}

	return r0
}

// Destination_Releases_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Releases'
type Destination_Releases_Call struct {
	*mock.Call
}

// Releases is a helper method to define mock.On call
func (_e *Destination_Expecter) Releases() *Destination_Releases_Call {
	return &Destination_Releases_Call{Call: _e.mock.On("Releases")}
}

func (_c *Destination_Releases_Call) Run(run func()) *Destination_Releases_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Destination_Releases_Call) Return(_a0 []terra.Release) *Destination_Releases_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Destination_Releases_Call) RunAndReturn(run func() []terra.Release) *Destination_Releases_Call {
	_c.Call.Return(run)
	return _c
}

// RequireSuitable provides a mock function with given fields:
func (_m *Destination) RequireSuitable() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Destination_RequireSuitable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RequireSuitable'
type Destination_RequireSuitable_Call struct {
	*mock.Call
}

// RequireSuitable is a helper method to define mock.On call
func (_e *Destination_Expecter) RequireSuitable() *Destination_RequireSuitable_Call {
	return &Destination_RequireSuitable_Call{Call: _e.mock.On("RequireSuitable")}
}

func (_c *Destination_RequireSuitable_Call) Run(run func()) *Destination_RequireSuitable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Destination_RequireSuitable_Call) Return(_a0 bool) *Destination_RequireSuitable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Destination_RequireSuitable_Call) RunAndReturn(run func() bool) *Destination_RequireSuitable_Call {
	_c.Call.Return(run)
	return _c
}

// RequiredRole provides a mock function with given fields:
func (_m *Destination) RequiredRole() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Destination_RequiredRole_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RequiredRole'
type Destination_RequiredRole_Call struct {
	*mock.Call
}

// RequiredRole is a helper method to define mock.On call
func (_e *Destination_Expecter) RequiredRole() *Destination_RequiredRole_Call {
	return &Destination_RequiredRole_Call{Call: _e.mock.On("RequiredRole")}
}

func (_c *Destination_RequiredRole_Call) Run(run func()) *Destination_RequiredRole_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Destination_RequiredRole_Call) Return(_a0 string) *Destination_RequiredRole_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Destination_RequiredRole_Call) RunAndReturn(run func() string) *Destination_RequiredRole_Call {
	_c.Call.Return(run)
	return _c
}

// TerraHelmfileRef provides a mock function with given fields:
func (_m *Destination) TerraHelmfileRef() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Destination_TerraHelmfileRef_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TerraHelmfileRef'
type Destination_TerraHelmfileRef_Call struct {
	*mock.Call
}

// TerraHelmfileRef is a helper method to define mock.On call
func (_e *Destination_Expecter) TerraHelmfileRef() *Destination_TerraHelmfileRef_Call {
	return &Destination_TerraHelmfileRef_Call{Call: _e.mock.On("TerraHelmfileRef")}
}

func (_c *Destination_TerraHelmfileRef_Call) Run(run func()) *Destination_TerraHelmfileRef_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Destination_TerraHelmfileRef_Call) Return(_a0 string) *Destination_TerraHelmfileRef_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Destination_TerraHelmfileRef_Call) RunAndReturn(run func() string) *Destination_TerraHelmfileRef_Call {
	_c.Call.Return(run)
	return _c
}

// Type provides a mock function with given fields:
func (_m *Destination) Type() terra.DestinationType {
	ret := _m.Called()

	var r0 terra.DestinationType
	if rf, ok := ret.Get(0).(func() terra.DestinationType); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(terra.DestinationType)
	}

	return r0
}

// Destination_Type_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Type'
type Destination_Type_Call struct {
	*mock.Call
}

// Type is a helper method to define mock.On call
func (_e *Destination_Expecter) Type() *Destination_Type_Call {
	return &Destination_Type_Call{Call: _e.mock.On("Type")}
}

func (_c *Destination_Type_Call) Run(run func()) *Destination_Type_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Destination_Type_Call) Return(_a0 terra.DestinationType) *Destination_Type_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Destination_Type_Call) RunAndReturn(run func() terra.DestinationType) *Destination_Type_Call {
	_c.Call.Return(run)
	return _c
}

// NewDestination creates a new instance of Destination. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDestination(t interface {
	mock.TestingT
	Cleanup(func())
}) *Destination {
	mock := &Destination{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
