// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	mock "github.com/stretchr/testify/mock"
)

// Cluster is an autogenerated mock type for the Cluster type
type Cluster struct {
	mock.Mock
}

type Cluster_Expecter struct {
	mock *mock.Mock
}

func (_m *Cluster) EXPECT() *Cluster_Expecter {
	return &Cluster_Expecter{mock: &_m.Mock}
}

// Address provides a mock function with given fields:
func (_m *Cluster) Address() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Cluster_Address_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Address'
type Cluster_Address_Call struct {
	*mock.Call
}

// Address is a helper method to define mock.On call
func (_e *Cluster_Expecter) Address() *Cluster_Address_Call {
	return &Cluster_Address_Call{Call: _e.mock.On("Address")}
}

func (_c *Cluster_Address_Call) Run(run func()) *Cluster_Address_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_Address_Call) Return(_a0 string) *Cluster_Address_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_Address_Call) RunAndReturn(run func() string) *Cluster_Address_Call {
	_c.Call.Return(run)
	return _c
}

// ArtifactBucket provides a mock function with given fields:
func (_m *Cluster) ArtifactBucket() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Cluster_ArtifactBucket_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ArtifactBucket'
type Cluster_ArtifactBucket_Call struct {
	*mock.Call
}

// ArtifactBucket is a helper method to define mock.On call
func (_e *Cluster_Expecter) ArtifactBucket() *Cluster_ArtifactBucket_Call {
	return &Cluster_ArtifactBucket_Call{Call: _e.mock.On("ArtifactBucket")}
}

func (_c *Cluster_ArtifactBucket_Call) Run(run func()) *Cluster_ArtifactBucket_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_ArtifactBucket_Call) Return(_a0 string) *Cluster_ArtifactBucket_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_ArtifactBucket_Call) RunAndReturn(run func() string) *Cluster_ArtifactBucket_Call {
	_c.Call.Return(run)
	return _c
}

// Base provides a mock function with given fields:
func (_m *Cluster) Base() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Cluster_Base_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Base'
type Cluster_Base_Call struct {
	*mock.Call
}

// Base is a helper method to define mock.On call
func (_e *Cluster_Expecter) Base() *Cluster_Base_Call {
	return &Cluster_Base_Call{Call: _e.mock.On("Base")}
}

func (_c *Cluster_Base_Call) Run(run func()) *Cluster_Base_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_Base_Call) Return(_a0 string) *Cluster_Base_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_Base_Call) RunAndReturn(run func() string) *Cluster_Base_Call {
	_c.Call.Return(run)
	return _c
}

// IsCluster provides a mock function with given fields:
func (_m *Cluster) IsCluster() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Cluster_IsCluster_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsCluster'
type Cluster_IsCluster_Call struct {
	*mock.Call
}

// IsCluster is a helper method to define mock.On call
func (_e *Cluster_Expecter) IsCluster() *Cluster_IsCluster_Call {
	return &Cluster_IsCluster_Call{Call: _e.mock.On("IsCluster")}
}

func (_c *Cluster_IsCluster_Call) Run(run func()) *Cluster_IsCluster_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_IsCluster_Call) Return(_a0 bool) *Cluster_IsCluster_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_IsCluster_Call) RunAndReturn(run func() bool) *Cluster_IsCluster_Call {
	_c.Call.Return(run)
	return _c
}

// IsEnvironment provides a mock function with given fields:
func (_m *Cluster) IsEnvironment() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Cluster_IsEnvironment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsEnvironment'
type Cluster_IsEnvironment_Call struct {
	*mock.Call
}

// IsEnvironment is a helper method to define mock.On call
func (_e *Cluster_Expecter) IsEnvironment() *Cluster_IsEnvironment_Call {
	return &Cluster_IsEnvironment_Call{Call: _e.mock.On("IsEnvironment")}
}

func (_c *Cluster_IsEnvironment_Call) Run(run func()) *Cluster_IsEnvironment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_IsEnvironment_Call) Return(_a0 bool) *Cluster_IsEnvironment_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_IsEnvironment_Call) RunAndReturn(run func() bool) *Cluster_IsEnvironment_Call {
	_c.Call.Return(run)
	return _c
}

// Location provides a mock function with given fields:
func (_m *Cluster) Location() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Cluster_Location_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Location'
type Cluster_Location_Call struct {
	*mock.Call
}

// Location is a helper method to define mock.On call
func (_e *Cluster_Expecter) Location() *Cluster_Location_Call {
	return &Cluster_Location_Call{Call: _e.mock.On("Location")}
}

func (_c *Cluster_Location_Call) Run(run func()) *Cluster_Location_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_Location_Call) Return(_a0 string) *Cluster_Location_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_Location_Call) RunAndReturn(run func() string) *Cluster_Location_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with given fields:
func (_m *Cluster) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Cluster_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type Cluster_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *Cluster_Expecter) Name() *Cluster_Name_Call {
	return &Cluster_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *Cluster_Name_Call) Run(run func()) *Cluster_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_Name_Call) Return(_a0 string) *Cluster_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_Name_Call) RunAndReturn(run func() string) *Cluster_Name_Call {
	_c.Call.Return(run)
	return _c
}

// Project provides a mock function with given fields:
func (_m *Cluster) Project() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Cluster_Project_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Project'
type Cluster_Project_Call struct {
	*mock.Call
}

// Project is a helper method to define mock.On call
func (_e *Cluster_Expecter) Project() *Cluster_Project_Call {
	return &Cluster_Project_Call{Call: _e.mock.On("Project")}
}

func (_c *Cluster_Project_Call) Run(run func()) *Cluster_Project_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_Project_Call) Return(_a0 string) *Cluster_Project_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_Project_Call) RunAndReturn(run func() string) *Cluster_Project_Call {
	_c.Call.Return(run)
	return _c
}

// ProjectSuffix provides a mock function with given fields:
func (_m *Cluster) ProjectSuffix() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Cluster_ProjectSuffix_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ProjectSuffix'
type Cluster_ProjectSuffix_Call struct {
	*mock.Call
}

// ProjectSuffix is a helper method to define mock.On call
func (_e *Cluster_Expecter) ProjectSuffix() *Cluster_ProjectSuffix_Call {
	return &Cluster_ProjectSuffix_Call{Call: _e.mock.On("ProjectSuffix")}
}

func (_c *Cluster_ProjectSuffix_Call) Run(run func()) *Cluster_ProjectSuffix_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_ProjectSuffix_Call) Return(_a0 string) *Cluster_ProjectSuffix_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_ProjectSuffix_Call) RunAndReturn(run func() string) *Cluster_ProjectSuffix_Call {
	_c.Call.Return(run)
	return _c
}

// ReleaseType provides a mock function with given fields:
func (_m *Cluster) ReleaseType() terra.ReleaseType {
	ret := _m.Called()

	var r0 terra.ReleaseType
	if rf, ok := ret.Get(0).(func() terra.ReleaseType); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(terra.ReleaseType)
	}

	return r0
}

// Cluster_ReleaseType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReleaseType'
type Cluster_ReleaseType_Call struct {
	*mock.Call
}

// ReleaseType is a helper method to define mock.On call
func (_e *Cluster_Expecter) ReleaseType() *Cluster_ReleaseType_Call {
	return &Cluster_ReleaseType_Call{Call: _e.mock.On("ReleaseType")}
}

func (_c *Cluster_ReleaseType_Call) Run(run func()) *Cluster_ReleaseType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_ReleaseType_Call) Return(_a0 terra.ReleaseType) *Cluster_ReleaseType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_ReleaseType_Call) RunAndReturn(run func() terra.ReleaseType) *Cluster_ReleaseType_Call {
	_c.Call.Return(run)
	return _c
}

// Releases provides a mock function with given fields:
func (_m *Cluster) Releases() []terra.Release {
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

// Cluster_Releases_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Releases'
type Cluster_Releases_Call struct {
	*mock.Call
}

// Releases is a helper method to define mock.On call
func (_e *Cluster_Expecter) Releases() *Cluster_Releases_Call {
	return &Cluster_Releases_Call{Call: _e.mock.On("Releases")}
}

func (_c *Cluster_Releases_Call) Run(run func()) *Cluster_Releases_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_Releases_Call) Return(_a0 []terra.Release) *Cluster_Releases_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_Releases_Call) RunAndReturn(run func() []terra.Release) *Cluster_Releases_Call {
	_c.Call.Return(run)
	return _c
}

// RequireSuitable provides a mock function with given fields:
func (_m *Cluster) RequireSuitable() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Cluster_RequireSuitable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RequireSuitable'
type Cluster_RequireSuitable_Call struct {
	*mock.Call
}

// RequireSuitable is a helper method to define mock.On call
func (_e *Cluster_Expecter) RequireSuitable() *Cluster_RequireSuitable_Call {
	return &Cluster_RequireSuitable_Call{Call: _e.mock.On("RequireSuitable")}
}

func (_c *Cluster_RequireSuitable_Call) Run(run func()) *Cluster_RequireSuitable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_RequireSuitable_Call) Return(_a0 bool) *Cluster_RequireSuitable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_RequireSuitable_Call) RunAndReturn(run func() bool) *Cluster_RequireSuitable_Call {
	_c.Call.Return(run)
	return _c
}

// TerraHelmfileRef provides a mock function with given fields:
func (_m *Cluster) TerraHelmfileRef() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Cluster_TerraHelmfileRef_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TerraHelmfileRef'
type Cluster_TerraHelmfileRef_Call struct {
	*mock.Call
}

// TerraHelmfileRef is a helper method to define mock.On call
func (_e *Cluster_Expecter) TerraHelmfileRef() *Cluster_TerraHelmfileRef_Call {
	return &Cluster_TerraHelmfileRef_Call{Call: _e.mock.On("TerraHelmfileRef")}
}

func (_c *Cluster_TerraHelmfileRef_Call) Run(run func()) *Cluster_TerraHelmfileRef_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_TerraHelmfileRef_Call) Return(_a0 string) *Cluster_TerraHelmfileRef_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_TerraHelmfileRef_Call) RunAndReturn(run func() string) *Cluster_TerraHelmfileRef_Call {
	_c.Call.Return(run)
	return _c
}

// Type provides a mock function with given fields:
func (_m *Cluster) Type() terra.DestinationType {
	ret := _m.Called()

	var r0 terra.DestinationType
	if rf, ok := ret.Get(0).(func() terra.DestinationType); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(terra.DestinationType)
	}

	return r0
}

// Cluster_Type_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Type'
type Cluster_Type_Call struct {
	*mock.Call
}

// Type is a helper method to define mock.On call
func (_e *Cluster_Expecter) Type() *Cluster_Type_Call {
	return &Cluster_Type_Call{Call: _e.mock.On("Type")}
}

func (_c *Cluster_Type_Call) Run(run func()) *Cluster_Type_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Cluster_Type_Call) Return(_a0 terra.DestinationType) *Cluster_Type_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Cluster_Type_Call) RunAndReturn(run func() terra.DestinationType) *Cluster_Type_Call {
	_c.Call.Return(run)
	return _c
}

// NewCluster creates a new instance of Cluster. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCluster(t interface {
	mock.TestingT
	Cleanup(func())
}) *Cluster {
	mock := &Cluster{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
