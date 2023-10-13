// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	mock "github.com/stretchr/testify/mock"
)

// Release is an autogenerated mock type for the Release type
type Release struct {
	mock.Mock
}

type Release_Expecter struct {
	mock *mock.Mock
}

func (_m *Release) EXPECT() *Release_Expecter {
	return &Release_Expecter{mock: &_m.Mock}
}

// AppVersion provides a mock function with given fields:
func (_m *Release) AppVersion() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Release_AppVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AppVersion'
type Release_AppVersion_Call struct {
	*mock.Call
}

// AppVersion is a helper method to define mock.On call
func (_e *Release_Expecter) AppVersion() *Release_AppVersion_Call {
	return &Release_AppVersion_Call{Call: _e.mock.On("AppVersion")}
}

func (_c *Release_AppVersion_Call) Run(run func()) *Release_AppVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_AppVersion_Call) Return(_a0 string) *Release_AppVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_AppVersion_Call) RunAndReturn(run func() string) *Release_AppVersion_Call {
	_c.Call.Return(run)
	return _c
}

// ChartName provides a mock function with given fields:
func (_m *Release) ChartName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Release_ChartName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChartName'
type Release_ChartName_Call struct {
	*mock.Call
}

// ChartName is a helper method to define mock.On call
func (_e *Release_Expecter) ChartName() *Release_ChartName_Call {
	return &Release_ChartName_Call{Call: _e.mock.On("ChartName")}
}

func (_c *Release_ChartName_Call) Run(run func()) *Release_ChartName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_ChartName_Call) Return(_a0 string) *Release_ChartName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_ChartName_Call) RunAndReturn(run func() string) *Release_ChartName_Call {
	_c.Call.Return(run)
	return _c
}

// ChartVersion provides a mock function with given fields:
func (_m *Release) ChartVersion() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Release_ChartVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChartVersion'
type Release_ChartVersion_Call struct {
	*mock.Call
}

// ChartVersion is a helper method to define mock.On call
func (_e *Release_Expecter) ChartVersion() *Release_ChartVersion_Call {
	return &Release_ChartVersion_Call{Call: _e.mock.On("ChartVersion")}
}

func (_c *Release_ChartVersion_Call) Run(run func()) *Release_ChartVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_ChartVersion_Call) Return(_a0 string) *Release_ChartVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_ChartVersion_Call) RunAndReturn(run func() string) *Release_ChartVersion_Call {
	_c.Call.Return(run)
	return _c
}

// Cluster provides a mock function with given fields:
func (_m *Release) Cluster() terra.Cluster {
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

// Release_Cluster_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Cluster'
type Release_Cluster_Call struct {
	*mock.Call
}

// Cluster is a helper method to define mock.On call
func (_e *Release_Expecter) Cluster() *Release_Cluster_Call {
	return &Release_Cluster_Call{Call: _e.mock.On("Cluster")}
}

func (_c *Release_Cluster_Call) Run(run func()) *Release_Cluster_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_Cluster_Call) Return(_a0 terra.Cluster) *Release_Cluster_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_Cluster_Call) RunAndReturn(run func() terra.Cluster) *Release_Cluster_Call {
	_c.Call.Return(run)
	return _c
}

// ClusterAddress provides a mock function with given fields:
func (_m *Release) ClusterAddress() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Release_ClusterAddress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClusterAddress'
type Release_ClusterAddress_Call struct {
	*mock.Call
}

// ClusterAddress is a helper method to define mock.On call
func (_e *Release_Expecter) ClusterAddress() *Release_ClusterAddress_Call {
	return &Release_ClusterAddress_Call{Call: _e.mock.On("ClusterAddress")}
}

func (_c *Release_ClusterAddress_Call) Run(run func()) *Release_ClusterAddress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_ClusterAddress_Call) Return(_a0 string) *Release_ClusterAddress_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_ClusterAddress_Call) RunAndReturn(run func() string) *Release_ClusterAddress_Call {
	_c.Call.Return(run)
	return _c
}

// ClusterName provides a mock function with given fields:
func (_m *Release) ClusterName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Release_ClusterName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClusterName'
type Release_ClusterName_Call struct {
	*mock.Call
}

// ClusterName is a helper method to define mock.On call
func (_e *Release_Expecter) ClusterName() *Release_ClusterName_Call {
	return &Release_ClusterName_Call{Call: _e.mock.On("ClusterName")}
}

func (_c *Release_ClusterName_Call) Run(run func()) *Release_ClusterName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_ClusterName_Call) Return(_a0 string) *Release_ClusterName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_ClusterName_Call) RunAndReturn(run func() string) *Release_ClusterName_Call {
	_c.Call.Return(run)
	return _c
}

// Destination provides a mock function with given fields:
func (_m *Release) Destination() terra.Destination {
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

// Release_Destination_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Destination'
type Release_Destination_Call struct {
	*mock.Call
}

// Destination is a helper method to define mock.On call
func (_e *Release_Expecter) Destination() *Release_Destination_Call {
	return &Release_Destination_Call{Call: _e.mock.On("Destination")}
}

func (_c *Release_Destination_Call) Run(run func()) *Release_Destination_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_Destination_Call) Return(_a0 terra.Destination) *Release_Destination_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_Destination_Call) RunAndReturn(run func() terra.Destination) *Release_Destination_Call {
	_c.Call.Return(run)
	return _c
}

// FirecloudDevelopRef provides a mock function with given fields:
func (_m *Release) FirecloudDevelopRef() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Release_FirecloudDevelopRef_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FirecloudDevelopRef'
type Release_FirecloudDevelopRef_Call struct {
	*mock.Call
}

// FirecloudDevelopRef is a helper method to define mock.On call
func (_e *Release_Expecter) FirecloudDevelopRef() *Release_FirecloudDevelopRef_Call {
	return &Release_FirecloudDevelopRef_Call{Call: _e.mock.On("FirecloudDevelopRef")}
}

func (_c *Release_FirecloudDevelopRef_Call) Run(run func()) *Release_FirecloudDevelopRef_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_FirecloudDevelopRef_Call) Return(_a0 string) *Release_FirecloudDevelopRef_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_FirecloudDevelopRef_Call) RunAndReturn(run func() string) *Release_FirecloudDevelopRef_Call {
	_c.Call.Return(run)
	return _c
}

// FullName provides a mock function with given fields:
func (_m *Release) FullName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Release_FullName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FullName'
type Release_FullName_Call struct {
	*mock.Call
}

// FullName is a helper method to define mock.On call
func (_e *Release_Expecter) FullName() *Release_FullName_Call {
	return &Release_FullName_Call{Call: _e.mock.On("FullName")}
}

func (_c *Release_FullName_Call) Run(run func()) *Release_FullName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_FullName_Call) Return(_a0 string) *Release_FullName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_FullName_Call) RunAndReturn(run func() string) *Release_FullName_Call {
	_c.Call.Return(run)
	return _c
}

// HelmfileOverlays provides a mock function with given fields:
func (_m *Release) HelmfileOverlays() []string {
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

// Release_HelmfileOverlays_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HelmfileOverlays'
type Release_HelmfileOverlays_Call struct {
	*mock.Call
}

// HelmfileOverlays is a helper method to define mock.On call
func (_e *Release_Expecter) HelmfileOverlays() *Release_HelmfileOverlays_Call {
	return &Release_HelmfileOverlays_Call{Call: _e.mock.On("HelmfileOverlays")}
}

func (_c *Release_HelmfileOverlays_Call) Run(run func()) *Release_HelmfileOverlays_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_HelmfileOverlays_Call) Return(_a0 []string) *Release_HelmfileOverlays_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_HelmfileOverlays_Call) RunAndReturn(run func() []string) *Release_HelmfileOverlays_Call {
	_c.Call.Return(run)
	return _c
}

// IsAppRelease provides a mock function with given fields:
func (_m *Release) IsAppRelease() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Release_IsAppRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsAppRelease'
type Release_IsAppRelease_Call struct {
	*mock.Call
}

// IsAppRelease is a helper method to define mock.On call
func (_e *Release_Expecter) IsAppRelease() *Release_IsAppRelease_Call {
	return &Release_IsAppRelease_Call{Call: _e.mock.On("IsAppRelease")}
}

func (_c *Release_IsAppRelease_Call) Run(run func()) *Release_IsAppRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_IsAppRelease_Call) Return(_a0 bool) *Release_IsAppRelease_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_IsAppRelease_Call) RunAndReturn(run func() bool) *Release_IsAppRelease_Call {
	_c.Call.Return(run)
	return _c
}

// IsClusterRelease provides a mock function with given fields:
func (_m *Release) IsClusterRelease() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Release_IsClusterRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsClusterRelease'
type Release_IsClusterRelease_Call struct {
	*mock.Call
}

// IsClusterRelease is a helper method to define mock.On call
func (_e *Release_Expecter) IsClusterRelease() *Release_IsClusterRelease_Call {
	return &Release_IsClusterRelease_Call{Call: _e.mock.On("IsClusterRelease")}
}

func (_c *Release_IsClusterRelease_Call) Run(run func()) *Release_IsClusterRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_IsClusterRelease_Call) Return(_a0 bool) *Release_IsClusterRelease_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_IsClusterRelease_Call) RunAndReturn(run func() bool) *Release_IsClusterRelease_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with given fields:
func (_m *Release) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Release_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type Release_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *Release_Expecter) Name() *Release_Name_Call {
	return &Release_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *Release_Name_Call) Run(run func()) *Release_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_Name_Call) Return(_a0 string) *Release_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_Name_Call) RunAndReturn(run func() string) *Release_Name_Call {
	_c.Call.Return(run)
	return _c
}

// Namespace provides a mock function with given fields:
func (_m *Release) Namespace() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Release_Namespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Namespace'
type Release_Namespace_Call struct {
	*mock.Call
}

// Namespace is a helper method to define mock.On call
func (_e *Release_Expecter) Namespace() *Release_Namespace_Call {
	return &Release_Namespace_Call{Call: _e.mock.On("Namespace")}
}

func (_c *Release_Namespace_Call) Run(run func()) *Release_Namespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_Namespace_Call) Return(_a0 string) *Release_Namespace_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_Namespace_Call) RunAndReturn(run func() string) *Release_Namespace_Call {
	_c.Call.Return(run)
	return _c
}

// Repo provides a mock function with given fields:
func (_m *Release) Repo() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Release_Repo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Repo'
type Release_Repo_Call struct {
	*mock.Call
}

// Repo is a helper method to define mock.On call
func (_e *Release_Expecter) Repo() *Release_Repo_Call {
	return &Release_Repo_Call{Call: _e.mock.On("Repo")}
}

func (_c *Release_Repo_Call) Run(run func()) *Release_Repo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_Repo_Call) Return(_a0 string) *Release_Repo_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_Repo_Call) RunAndReturn(run func() string) *Release_Repo_Call {
	_c.Call.Return(run)
	return _c
}

// TerraHelmfileRef provides a mock function with given fields:
func (_m *Release) TerraHelmfileRef() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Release_TerraHelmfileRef_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TerraHelmfileRef'
type Release_TerraHelmfileRef_Call struct {
	*mock.Call
}

// TerraHelmfileRef is a helper method to define mock.On call
func (_e *Release_Expecter) TerraHelmfileRef() *Release_TerraHelmfileRef_Call {
	return &Release_TerraHelmfileRef_Call{Call: _e.mock.On("TerraHelmfileRef")}
}

func (_c *Release_TerraHelmfileRef_Call) Run(run func()) *Release_TerraHelmfileRef_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_TerraHelmfileRef_Call) Return(_a0 string) *Release_TerraHelmfileRef_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_TerraHelmfileRef_Call) RunAndReturn(run func() string) *Release_TerraHelmfileRef_Call {
	_c.Call.Return(run)
	return _c
}

// Type provides a mock function with given fields:
func (_m *Release) Type() terra.ReleaseType {
	ret := _m.Called()

	var r0 terra.ReleaseType
	if rf, ok := ret.Get(0).(func() terra.ReleaseType); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(terra.ReleaseType)
	}

	return r0
}

// Release_Type_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Type'
type Release_Type_Call struct {
	*mock.Call
}

// Type is a helper method to define mock.On call
func (_e *Release_Expecter) Type() *Release_Type_Call {
	return &Release_Type_Call{Call: _e.mock.On("Type")}
}

func (_c *Release_Type_Call) Run(run func()) *Release_Type_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Release_Type_Call) Return(_a0 terra.ReleaseType) *Release_Type_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Release_Type_Call) RunAndReturn(run func() terra.ReleaseType) *Release_Type_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewRelease interface {
	mock.TestingT
	Cleanup(func())
}

// NewRelease creates a new instance of Release. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRelease(t mockConstructorTestingTNewRelease) *Release {
	mock := &Release{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
