// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	mock "github.com/stretchr/testify/mock"
)

// AppRelease is an autogenerated mock type for the AppRelease type
type AppRelease struct {
	mock.Mock
}

type AppRelease_Expecter struct {
	mock *mock.Mock
}

func (_m *AppRelease) EXPECT() *AppRelease_Expecter {
	return &AppRelease_Expecter{mock: &_m.Mock}
}

// AppVersion provides a mock function with given fields:
func (_m *AppRelease) AppVersion() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_AppVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AppVersion'
type AppRelease_AppVersion_Call struct {
	*mock.Call
}

// AppVersion is a helper method to define mock.On call
func (_e *AppRelease_Expecter) AppVersion() *AppRelease_AppVersion_Call {
	return &AppRelease_AppVersion_Call{Call: _e.mock.On("AppVersion")}
}

func (_c *AppRelease_AppVersion_Call) Run(run func()) *AppRelease_AppVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_AppVersion_Call) Return(_a0 string) *AppRelease_AppVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_AppVersion_Call) RunAndReturn(run func() string) *AppRelease_AppVersion_Call {
	_c.Call.Return(run)
	return _c
}

// ChartName provides a mock function with given fields:
func (_m *AppRelease) ChartName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_ChartName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChartName'
type AppRelease_ChartName_Call struct {
	*mock.Call
}

// ChartName is a helper method to define mock.On call
func (_e *AppRelease_Expecter) ChartName() *AppRelease_ChartName_Call {
	return &AppRelease_ChartName_Call{Call: _e.mock.On("ChartName")}
}

func (_c *AppRelease_ChartName_Call) Run(run func()) *AppRelease_ChartName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_ChartName_Call) Return(_a0 string) *AppRelease_ChartName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_ChartName_Call) RunAndReturn(run func() string) *AppRelease_ChartName_Call {
	_c.Call.Return(run)
	return _c
}

// ChartVersion provides a mock function with given fields:
func (_m *AppRelease) ChartVersion() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_ChartVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChartVersion'
type AppRelease_ChartVersion_Call struct {
	*mock.Call
}

// ChartVersion is a helper method to define mock.On call
func (_e *AppRelease_Expecter) ChartVersion() *AppRelease_ChartVersion_Call {
	return &AppRelease_ChartVersion_Call{Call: _e.mock.On("ChartVersion")}
}

func (_c *AppRelease_ChartVersion_Call) Run(run func()) *AppRelease_ChartVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_ChartVersion_Call) Return(_a0 string) *AppRelease_ChartVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_ChartVersion_Call) RunAndReturn(run func() string) *AppRelease_ChartVersion_Call {
	_c.Call.Return(run)
	return _c
}

// Cluster provides a mock function with given fields:
func (_m *AppRelease) Cluster() terra.Cluster {
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

// AppRelease_Cluster_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Cluster'
type AppRelease_Cluster_Call struct {
	*mock.Call
}

// Cluster is a helper method to define mock.On call
func (_e *AppRelease_Expecter) Cluster() *AppRelease_Cluster_Call {
	return &AppRelease_Cluster_Call{Call: _e.mock.On("Cluster")}
}

func (_c *AppRelease_Cluster_Call) Run(run func()) *AppRelease_Cluster_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_Cluster_Call) Return(_a0 terra.Cluster) *AppRelease_Cluster_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_Cluster_Call) RunAndReturn(run func() terra.Cluster) *AppRelease_Cluster_Call {
	_c.Call.Return(run)
	return _c
}

// ClusterAddress provides a mock function with given fields:
func (_m *AppRelease) ClusterAddress() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_ClusterAddress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClusterAddress'
type AppRelease_ClusterAddress_Call struct {
	*mock.Call
}

// ClusterAddress is a helper method to define mock.On call
func (_e *AppRelease_Expecter) ClusterAddress() *AppRelease_ClusterAddress_Call {
	return &AppRelease_ClusterAddress_Call{Call: _e.mock.On("ClusterAddress")}
}

func (_c *AppRelease_ClusterAddress_Call) Run(run func()) *AppRelease_ClusterAddress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_ClusterAddress_Call) Return(_a0 string) *AppRelease_ClusterAddress_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_ClusterAddress_Call) RunAndReturn(run func() string) *AppRelease_ClusterAddress_Call {
	_c.Call.Return(run)
	return _c
}

// ClusterName provides a mock function with given fields:
func (_m *AppRelease) ClusterName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_ClusterName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClusterName'
type AppRelease_ClusterName_Call struct {
	*mock.Call
}

// ClusterName is a helper method to define mock.On call
func (_e *AppRelease_Expecter) ClusterName() *AppRelease_ClusterName_Call {
	return &AppRelease_ClusterName_Call{Call: _e.mock.On("ClusterName")}
}

func (_c *AppRelease_ClusterName_Call) Run(run func()) *AppRelease_ClusterName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_ClusterName_Call) Return(_a0 string) *AppRelease_ClusterName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_ClusterName_Call) RunAndReturn(run func() string) *AppRelease_ClusterName_Call {
	_c.Call.Return(run)
	return _c
}

// Destination provides a mock function with given fields:
func (_m *AppRelease) Destination() terra.Destination {
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

// AppRelease_Destination_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Destination'
type AppRelease_Destination_Call struct {
	*mock.Call
}

// Destination is a helper method to define mock.On call
func (_e *AppRelease_Expecter) Destination() *AppRelease_Destination_Call {
	return &AppRelease_Destination_Call{Call: _e.mock.On("Destination")}
}

func (_c *AppRelease_Destination_Call) Run(run func()) *AppRelease_Destination_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_Destination_Call) Return(_a0 terra.Destination) *AppRelease_Destination_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_Destination_Call) RunAndReturn(run func() terra.Destination) *AppRelease_Destination_Call {
	_c.Call.Return(run)
	return _c
}

// Environment provides a mock function with given fields:
func (_m *AppRelease) Environment() terra.Environment {
	ret := _m.Called()

	var r0 terra.Environment
	if rf, ok := ret.Get(0).(func() terra.Environment); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(terra.Environment)
		}
	}

	return r0
}

// AppRelease_Environment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Environment'
type AppRelease_Environment_Call struct {
	*mock.Call
}

// Environment is a helper method to define mock.On call
func (_e *AppRelease_Expecter) Environment() *AppRelease_Environment_Call {
	return &AppRelease_Environment_Call{Call: _e.mock.On("Environment")}
}

func (_c *AppRelease_Environment_Call) Run(run func()) *AppRelease_Environment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_Environment_Call) Return(_a0 terra.Environment) *AppRelease_Environment_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_Environment_Call) RunAndReturn(run func() terra.Environment) *AppRelease_Environment_Call {
	_c.Call.Return(run)
	return _c
}

// FirecloudDevelopRef provides a mock function with given fields:
func (_m *AppRelease) FirecloudDevelopRef() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_FirecloudDevelopRef_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FirecloudDevelopRef'
type AppRelease_FirecloudDevelopRef_Call struct {
	*mock.Call
}

// FirecloudDevelopRef is a helper method to define mock.On call
func (_e *AppRelease_Expecter) FirecloudDevelopRef() *AppRelease_FirecloudDevelopRef_Call {
	return &AppRelease_FirecloudDevelopRef_Call{Call: _e.mock.On("FirecloudDevelopRef")}
}

func (_c *AppRelease_FirecloudDevelopRef_Call) Run(run func()) *AppRelease_FirecloudDevelopRef_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_FirecloudDevelopRef_Call) Return(_a0 string) *AppRelease_FirecloudDevelopRef_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_FirecloudDevelopRef_Call) RunAndReturn(run func() string) *AppRelease_FirecloudDevelopRef_Call {
	_c.Call.Return(run)
	return _c
}

// FullName provides a mock function with given fields:
func (_m *AppRelease) FullName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_FullName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FullName'
type AppRelease_FullName_Call struct {
	*mock.Call
}

// FullName is a helper method to define mock.On call
func (_e *AppRelease_Expecter) FullName() *AppRelease_FullName_Call {
	return &AppRelease_FullName_Call{Call: _e.mock.On("FullName")}
}

func (_c *AppRelease_FullName_Call) Run(run func()) *AppRelease_FullName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_FullName_Call) Return(_a0 string) *AppRelease_FullName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_FullName_Call) RunAndReturn(run func() string) *AppRelease_FullName_Call {
	_c.Call.Return(run)
	return _c
}

// HelmfileOverlays provides a mock function with given fields:
func (_m *AppRelease) HelmfileOverlays() []string {
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

// AppRelease_HelmfileOverlays_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HelmfileOverlays'
type AppRelease_HelmfileOverlays_Call struct {
	*mock.Call
}

// HelmfileOverlays is a helper method to define mock.On call
func (_e *AppRelease_Expecter) HelmfileOverlays() *AppRelease_HelmfileOverlays_Call {
	return &AppRelease_HelmfileOverlays_Call{Call: _e.mock.On("HelmfileOverlays")}
}

func (_c *AppRelease_HelmfileOverlays_Call) Run(run func()) *AppRelease_HelmfileOverlays_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_HelmfileOverlays_Call) Return(_a0 []string) *AppRelease_HelmfileOverlays_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_HelmfileOverlays_Call) RunAndReturn(run func() []string) *AppRelease_HelmfileOverlays_Call {
	_c.Call.Return(run)
	return _c
}

// Host provides a mock function with given fields:
func (_m *AppRelease) Host() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_Host_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Host'
type AppRelease_Host_Call struct {
	*mock.Call
}

// Host is a helper method to define mock.On call
func (_e *AppRelease_Expecter) Host() *AppRelease_Host_Call {
	return &AppRelease_Host_Call{Call: _e.mock.On("Host")}
}

func (_c *AppRelease_Host_Call) Run(run func()) *AppRelease_Host_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_Host_Call) Return(_a0 string) *AppRelease_Host_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_Host_Call) RunAndReturn(run func() string) *AppRelease_Host_Call {
	_c.Call.Return(run)
	return _c
}

// IsAppRelease provides a mock function with given fields:
func (_m *AppRelease) IsAppRelease() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// AppRelease_IsAppRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsAppRelease'
type AppRelease_IsAppRelease_Call struct {
	*mock.Call
}

// IsAppRelease is a helper method to define mock.On call
func (_e *AppRelease_Expecter) IsAppRelease() *AppRelease_IsAppRelease_Call {
	return &AppRelease_IsAppRelease_Call{Call: _e.mock.On("IsAppRelease")}
}

func (_c *AppRelease_IsAppRelease_Call) Run(run func()) *AppRelease_IsAppRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_IsAppRelease_Call) Return(_a0 bool) *AppRelease_IsAppRelease_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_IsAppRelease_Call) RunAndReturn(run func() bool) *AppRelease_IsAppRelease_Call {
	_c.Call.Return(run)
	return _c
}

// IsClusterRelease provides a mock function with given fields:
func (_m *AppRelease) IsClusterRelease() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// AppRelease_IsClusterRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsClusterRelease'
type AppRelease_IsClusterRelease_Call struct {
	*mock.Call
}

// IsClusterRelease is a helper method to define mock.On call
func (_e *AppRelease_Expecter) IsClusterRelease() *AppRelease_IsClusterRelease_Call {
	return &AppRelease_IsClusterRelease_Call{Call: _e.mock.On("IsClusterRelease")}
}

func (_c *AppRelease_IsClusterRelease_Call) Run(run func()) *AppRelease_IsClusterRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_IsClusterRelease_Call) Return(_a0 bool) *AppRelease_IsClusterRelease_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_IsClusterRelease_Call) RunAndReturn(run func() bool) *AppRelease_IsClusterRelease_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with given fields:
func (_m *AppRelease) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type AppRelease_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *AppRelease_Expecter) Name() *AppRelease_Name_Call {
	return &AppRelease_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *AppRelease_Name_Call) Run(run func()) *AppRelease_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_Name_Call) Return(_a0 string) *AppRelease_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_Name_Call) RunAndReturn(run func() string) *AppRelease_Name_Call {
	_c.Call.Return(run)
	return _c
}

// Namespace provides a mock function with given fields:
func (_m *AppRelease) Namespace() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_Namespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Namespace'
type AppRelease_Namespace_Call struct {
	*mock.Call
}

// Namespace is a helper method to define mock.On call
func (_e *AppRelease_Expecter) Namespace() *AppRelease_Namespace_Call {
	return &AppRelease_Namespace_Call{Call: _e.mock.On("Namespace")}
}

func (_c *AppRelease_Namespace_Call) Run(run func()) *AppRelease_Namespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_Namespace_Call) Return(_a0 string) *AppRelease_Namespace_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_Namespace_Call) RunAndReturn(run func() string) *AppRelease_Namespace_Call {
	_c.Call.Return(run)
	return _c
}

// Port provides a mock function with given fields:
func (_m *AppRelease) Port() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// AppRelease_Port_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Port'
type AppRelease_Port_Call struct {
	*mock.Call
}

// Port is a helper method to define mock.On call
func (_e *AppRelease_Expecter) Port() *AppRelease_Port_Call {
	return &AppRelease_Port_Call{Call: _e.mock.On("Port")}
}

func (_c *AppRelease_Port_Call) Run(run func()) *AppRelease_Port_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_Port_Call) Return(_a0 int) *AppRelease_Port_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_Port_Call) RunAndReturn(run func() int) *AppRelease_Port_Call {
	_c.Call.Return(run)
	return _c
}

// Protocol provides a mock function with given fields:
func (_m *AppRelease) Protocol() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_Protocol_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Protocol'
type AppRelease_Protocol_Call struct {
	*mock.Call
}

// Protocol is a helper method to define mock.On call
func (_e *AppRelease_Expecter) Protocol() *AppRelease_Protocol_Call {
	return &AppRelease_Protocol_Call{Call: _e.mock.On("Protocol")}
}

func (_c *AppRelease_Protocol_Call) Run(run func()) *AppRelease_Protocol_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_Protocol_Call) Return(_a0 string) *AppRelease_Protocol_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_Protocol_Call) RunAndReturn(run func() string) *AppRelease_Protocol_Call {
	_c.Call.Return(run)
	return _c
}

// Repo provides a mock function with given fields:
func (_m *AppRelease) Repo() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_Repo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Repo'
type AppRelease_Repo_Call struct {
	*mock.Call
}

// Repo is a helper method to define mock.On call
func (_e *AppRelease_Expecter) Repo() *AppRelease_Repo_Call {
	return &AppRelease_Repo_Call{Call: _e.mock.On("Repo")}
}

func (_c *AppRelease_Repo_Call) Run(run func()) *AppRelease_Repo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_Repo_Call) Return(_a0 string) *AppRelease_Repo_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_Repo_Call) RunAndReturn(run func() string) *AppRelease_Repo_Call {
	_c.Call.Return(run)
	return _c
}

// Subdomain provides a mock function with given fields:
func (_m *AppRelease) Subdomain() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_Subdomain_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Subdomain'
type AppRelease_Subdomain_Call struct {
	*mock.Call
}

// Subdomain is a helper method to define mock.On call
func (_e *AppRelease_Expecter) Subdomain() *AppRelease_Subdomain_Call {
	return &AppRelease_Subdomain_Call{Call: _e.mock.On("Subdomain")}
}

func (_c *AppRelease_Subdomain_Call) Run(run func()) *AppRelease_Subdomain_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_Subdomain_Call) Return(_a0 string) *AppRelease_Subdomain_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_Subdomain_Call) RunAndReturn(run func() string) *AppRelease_Subdomain_Call {
	_c.Call.Return(run)
	return _c
}

// TerraHelmfileRef provides a mock function with given fields:
func (_m *AppRelease) TerraHelmfileRef() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_TerraHelmfileRef_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TerraHelmfileRef'
type AppRelease_TerraHelmfileRef_Call struct {
	*mock.Call
}

// TerraHelmfileRef is a helper method to define mock.On call
func (_e *AppRelease_Expecter) TerraHelmfileRef() *AppRelease_TerraHelmfileRef_Call {
	return &AppRelease_TerraHelmfileRef_Call{Call: _e.mock.On("TerraHelmfileRef")}
}

func (_c *AppRelease_TerraHelmfileRef_Call) Run(run func()) *AppRelease_TerraHelmfileRef_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_TerraHelmfileRef_Call) Return(_a0 string) *AppRelease_TerraHelmfileRef_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_TerraHelmfileRef_Call) RunAndReturn(run func() string) *AppRelease_TerraHelmfileRef_Call {
	_c.Call.Return(run)
	return _c
}

// Type provides a mock function with given fields:
func (_m *AppRelease) Type() terra.ReleaseType {
	ret := _m.Called()

	var r0 terra.ReleaseType
	if rf, ok := ret.Get(0).(func() terra.ReleaseType); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(terra.ReleaseType)
	}

	return r0
}

// AppRelease_Type_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Type'
type AppRelease_Type_Call struct {
	*mock.Call
}

// Type is a helper method to define mock.On call
func (_e *AppRelease_Expecter) Type() *AppRelease_Type_Call {
	return &AppRelease_Type_Call{Call: _e.mock.On("Type")}
}

func (_c *AppRelease_Type_Call) Run(run func()) *AppRelease_Type_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_Type_Call) Return(_a0 terra.ReleaseType) *AppRelease_Type_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_Type_Call) RunAndReturn(run func() terra.ReleaseType) *AppRelease_Type_Call {
	_c.Call.Return(run)
	return _c
}

// URL provides a mock function with given fields:
func (_m *AppRelease) URL() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppRelease_URL_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'URL'
type AppRelease_URL_Call struct {
	*mock.Call
}

// URL is a helper method to define mock.On call
func (_e *AppRelease_Expecter) URL() *AppRelease_URL_Call {
	return &AppRelease_URL_Call{Call: _e.mock.On("URL")}
}

func (_c *AppRelease_URL_Call) Run(run func()) *AppRelease_URL_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AppRelease_URL_Call) Return(_a0 string) *AppRelease_URL_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AppRelease_URL_Call) RunAndReturn(run func() string) *AppRelease_URL_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewAppRelease interface {
	mock.TestingT
	Cleanup(func())
}

// NewAppRelease creates a new instance of AppRelease. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAppRelease(t mockConstructorTestingTNewAppRelease) *AppRelease {
	mock := &AppRelease{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
