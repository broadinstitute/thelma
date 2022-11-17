// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	mock "github.com/stretchr/testify/mock"
)

// ClusterRelease is an autogenerated mock type for the ClusterRelease type
type ClusterRelease struct {
	mock.Mock
}

type ClusterRelease_Expecter struct {
	mock *mock.Mock
}

func (_m *ClusterRelease) EXPECT() *ClusterRelease_Expecter {
	return &ClusterRelease_Expecter{mock: &_m.Mock}
}

// ChartName provides a mock function with given fields:
func (_m *ClusterRelease) ChartName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ClusterRelease_ChartName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChartName'
type ClusterRelease_ChartName_Call struct {
	*mock.Call
}

// ChartName is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) ChartName() *ClusterRelease_ChartName_Call {
	return &ClusterRelease_ChartName_Call{Call: _e.mock.On("ChartName")}
}

func (_c *ClusterRelease_ChartName_Call) Run(run func()) *ClusterRelease_ChartName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_ChartName_Call) Return(_a0 string) *ClusterRelease_ChartName_Call {
	_c.Call.Return(_a0)
	return _c
}

// ChartVersion provides a mock function with given fields:
func (_m *ClusterRelease) ChartVersion() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ClusterRelease_ChartVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChartVersion'
type ClusterRelease_ChartVersion_Call struct {
	*mock.Call
}

// ChartVersion is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) ChartVersion() *ClusterRelease_ChartVersion_Call {
	return &ClusterRelease_ChartVersion_Call{Call: _e.mock.On("ChartVersion")}
}

func (_c *ClusterRelease_ChartVersion_Call) Run(run func()) *ClusterRelease_ChartVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_ChartVersion_Call) Return(_a0 string) *ClusterRelease_ChartVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

// Cluster provides a mock function with given fields:
func (_m *ClusterRelease) Cluster() terra.Cluster {
	ret := _m.Called()

	var r0 terra.Cluster
	if rf, ok := ret.Get(0).(func() terra.Cluster); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(terra.Cluster)
		}
	}

	return r0
}

// ClusterRelease_Cluster_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Cluster'
type ClusterRelease_Cluster_Call struct {
	*mock.Call
}

// Cluster is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) Cluster() *ClusterRelease_Cluster_Call {
	return &ClusterRelease_Cluster_Call{Call: _e.mock.On("Cluster")}
}

func (_c *ClusterRelease_Cluster_Call) Run(run func()) *ClusterRelease_Cluster_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_Cluster_Call) Return(_a0 terra.Cluster) *ClusterRelease_Cluster_Call {
	_c.Call.Return(_a0)
	return _c
}

// ClusterAddress provides a mock function with given fields:
func (_m *ClusterRelease) ClusterAddress() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ClusterRelease_ClusterAddress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClusterAddress'
type ClusterRelease_ClusterAddress_Call struct {
	*mock.Call
}

// ClusterAddress is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) ClusterAddress() *ClusterRelease_ClusterAddress_Call {
	return &ClusterRelease_ClusterAddress_Call{Call: _e.mock.On("ClusterAddress")}
}

func (_c *ClusterRelease_ClusterAddress_Call) Run(run func()) *ClusterRelease_ClusterAddress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_ClusterAddress_Call) Return(_a0 string) *ClusterRelease_ClusterAddress_Call {
	_c.Call.Return(_a0)
	return _c
}

// ClusterName provides a mock function with given fields:
func (_m *ClusterRelease) ClusterName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ClusterRelease_ClusterName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClusterName'
type ClusterRelease_ClusterName_Call struct {
	*mock.Call
}

// ClusterName is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) ClusterName() *ClusterRelease_ClusterName_Call {
	return &ClusterRelease_ClusterName_Call{Call: _e.mock.On("ClusterName")}
}

func (_c *ClusterRelease_ClusterName_Call) Run(run func()) *ClusterRelease_ClusterName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_ClusterName_Call) Return(_a0 string) *ClusterRelease_ClusterName_Call {
	_c.Call.Return(_a0)
	return _c
}

// Destination provides a mock function with given fields:
func (_m *ClusterRelease) Destination() terra.Destination {
	ret := _m.Called()

	var r0 terra.Destination
	if rf, ok := ret.Get(0).(func() terra.Destination); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(terra.Destination)
		}
	}

	return r0
}

// ClusterRelease_Destination_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Destination'
type ClusterRelease_Destination_Call struct {
	*mock.Call
}

// Destination is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) Destination() *ClusterRelease_Destination_Call {
	return &ClusterRelease_Destination_Call{Call: _e.mock.On("Destination")}
}

func (_c *ClusterRelease_Destination_Call) Run(run func()) *ClusterRelease_Destination_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_Destination_Call) Return(_a0 terra.Destination) *ClusterRelease_Destination_Call {
	_c.Call.Return(_a0)
	return _c
}

// FirecloudDevelopRef provides a mock function with given fields:
func (_m *ClusterRelease) FirecloudDevelopRef() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ClusterRelease_FirecloudDevelopRef_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FirecloudDevelopRef'
type ClusterRelease_FirecloudDevelopRef_Call struct {
	*mock.Call
}

// FirecloudDevelopRef is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) FirecloudDevelopRef() *ClusterRelease_FirecloudDevelopRef_Call {
	return &ClusterRelease_FirecloudDevelopRef_Call{Call: _e.mock.On("FirecloudDevelopRef")}
}

func (_c *ClusterRelease_FirecloudDevelopRef_Call) Run(run func()) *ClusterRelease_FirecloudDevelopRef_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_FirecloudDevelopRef_Call) Return(_a0 string) *ClusterRelease_FirecloudDevelopRef_Call {
	_c.Call.Return(_a0)
	return _c
}

// FullName provides a mock function with given fields:
func (_m *ClusterRelease) FullName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ClusterRelease_FullName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FullName'
type ClusterRelease_FullName_Call struct {
	*mock.Call
}

// FullName is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) FullName() *ClusterRelease_FullName_Call {
	return &ClusterRelease_FullName_Call{Call: _e.mock.On("FullName")}
}

func (_c *ClusterRelease_FullName_Call) Run(run func()) *ClusterRelease_FullName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_FullName_Call) Return(_a0 string) *ClusterRelease_FullName_Call {
	_c.Call.Return(_a0)
	return _c
}

// IsAppRelease provides a mock function with given fields:
func (_m *ClusterRelease) IsAppRelease() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ClusterRelease_IsAppRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsAppRelease'
type ClusterRelease_IsAppRelease_Call struct {
	*mock.Call
}

// IsAppRelease is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) IsAppRelease() *ClusterRelease_IsAppRelease_Call {
	return &ClusterRelease_IsAppRelease_Call{Call: _e.mock.On("IsAppRelease")}
}

func (_c *ClusterRelease_IsAppRelease_Call) Run(run func()) *ClusterRelease_IsAppRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_IsAppRelease_Call) Return(_a0 bool) *ClusterRelease_IsAppRelease_Call {
	_c.Call.Return(_a0)
	return _c
}

// IsClusterRelease provides a mock function with given fields:
func (_m *ClusterRelease) IsClusterRelease() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ClusterRelease_IsClusterRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsClusterRelease'
type ClusterRelease_IsClusterRelease_Call struct {
	*mock.Call
}

// IsClusterRelease is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) IsClusterRelease() *ClusterRelease_IsClusterRelease_Call {
	return &ClusterRelease_IsClusterRelease_Call{Call: _e.mock.On("IsClusterRelease")}
}

func (_c *ClusterRelease_IsClusterRelease_Call) Run(run func()) *ClusterRelease_IsClusterRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_IsClusterRelease_Call) Return(_a0 bool) *ClusterRelease_IsClusterRelease_Call {
	_c.Call.Return(_a0)
	return _c
}

// Name provides a mock function with given fields:
func (_m *ClusterRelease) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ClusterRelease_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type ClusterRelease_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) Name() *ClusterRelease_Name_Call {
	return &ClusterRelease_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *ClusterRelease_Name_Call) Run(run func()) *ClusterRelease_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_Name_Call) Return(_a0 string) *ClusterRelease_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

// Namespace provides a mock function with given fields:
func (_m *ClusterRelease) Namespace() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ClusterRelease_Namespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Namespace'
type ClusterRelease_Namespace_Call struct {
	*mock.Call
}

// Namespace is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) Namespace() *ClusterRelease_Namespace_Call {
	return &ClusterRelease_Namespace_Call{Call: _e.mock.On("Namespace")}
}

func (_c *ClusterRelease_Namespace_Call) Run(run func()) *ClusterRelease_Namespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_Namespace_Call) Return(_a0 string) *ClusterRelease_Namespace_Call {
	_c.Call.Return(_a0)
	return _c
}

// Repo provides a mock function with given fields:
func (_m *ClusterRelease) Repo() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ClusterRelease_Repo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Repo'
type ClusterRelease_Repo_Call struct {
	*mock.Call
}

// Repo is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) Repo() *ClusterRelease_Repo_Call {
	return &ClusterRelease_Repo_Call{Call: _e.mock.On("Repo")}
}

func (_c *ClusterRelease_Repo_Call) Run(run func()) *ClusterRelease_Repo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_Repo_Call) Return(_a0 string) *ClusterRelease_Repo_Call {
	_c.Call.Return(_a0)
	return _c
}

// TerraHelmfileRef provides a mock function with given fields:
func (_m *ClusterRelease) TerraHelmfileRef() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ClusterRelease_TerraHelmfileRef_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TerraHelmfileRef'
type ClusterRelease_TerraHelmfileRef_Call struct {
	*mock.Call
}

// TerraHelmfileRef is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) TerraHelmfileRef() *ClusterRelease_TerraHelmfileRef_Call {
	return &ClusterRelease_TerraHelmfileRef_Call{Call: _e.mock.On("TerraHelmfileRef")}
}

func (_c *ClusterRelease_TerraHelmfileRef_Call) Run(run func()) *ClusterRelease_TerraHelmfileRef_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_TerraHelmfileRef_Call) Return(_a0 string) *ClusterRelease_TerraHelmfileRef_Call {
	_c.Call.Return(_a0)
	return _c
}

// Type provides a mock function with given fields:
func (_m *ClusterRelease) Type() terra.ReleaseType {
	ret := _m.Called()

	var r0 terra.ReleaseType
	if rf, ok := ret.Get(0).(func() terra.ReleaseType); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(terra.ReleaseType)
	}

	return r0
}

// ClusterRelease_Type_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Type'
type ClusterRelease_Type_Call struct {
	*mock.Call
}

// Type is a helper method to define mock.On call
func (_e *ClusterRelease_Expecter) Type() *ClusterRelease_Type_Call {
	return &ClusterRelease_Type_Call{Call: _e.mock.On("Type")}
}

func (_c *ClusterRelease_Type_Call) Run(run func()) *ClusterRelease_Type_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ClusterRelease_Type_Call) Return(_a0 terra.ReleaseType) *ClusterRelease_Type_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewClusterRelease interface {
	mock.TestingT
	Cleanup(func())
}

// NewClusterRelease creates a new instance of ClusterRelease. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewClusterRelease(t mockConstructorTestingTNewClusterRelease) *ClusterRelease {
	mock := &ClusterRelease{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
