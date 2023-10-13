// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	time "time"

	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	mock "github.com/stretchr/testify/mock"
)

// Environment is an autogenerated mock type for the Environment type
type Environment struct {
	mock.Mock
}

type Environment_Expecter struct {
	mock *mock.Mock
}

func (_m *Environment) EXPECT() *Environment_Expecter {
	return &Environment_Expecter{mock: &_m.Mock}
}

// AutoDelete provides a mock function with given fields:
func (_m *Environment) AutoDelete() terra.AutoDelete {
	ret := _m.Called()

	var r0 terra.AutoDelete
	if rf, ok := ret.Get(0).(func() terra.AutoDelete); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(terra.AutoDelete)
		}
	}

	return r0
}

// Environment_AutoDelete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AutoDelete'
type Environment_AutoDelete_Call struct {
	*mock.Call
}

// AutoDelete is a helper method to define mock.On call
func (_e *Environment_Expecter) AutoDelete() *Environment_AutoDelete_Call {
	return &Environment_AutoDelete_Call{Call: _e.mock.On("AutoDelete")}
}

func (_c *Environment_AutoDelete_Call) Run(run func()) *Environment_AutoDelete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_AutoDelete_Call) Return(_a0 terra.AutoDelete) *Environment_AutoDelete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_AutoDelete_Call) RunAndReturn(run func() terra.AutoDelete) *Environment_AutoDelete_Call {
	_c.Call.Return(run)
	return _c
}

// Base provides a mock function with given fields:
func (_m *Environment) Base() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Environment_Base_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Base'
type Environment_Base_Call struct {
	*mock.Call
}

// Base is a helper method to define mock.On call
func (_e *Environment_Expecter) Base() *Environment_Base_Call {
	return &Environment_Base_Call{Call: _e.mock.On("Base")}
}

func (_c *Environment_Base_Call) Run(run func()) *Environment_Base_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_Base_Call) Return(_a0 string) *Environment_Base_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_Base_Call) RunAndReturn(run func() string) *Environment_Base_Call {
	_c.Call.Return(run)
	return _c
}

// BaseDomain provides a mock function with given fields:
func (_m *Environment) BaseDomain() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Environment_BaseDomain_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BaseDomain'
type Environment_BaseDomain_Call struct {
	*mock.Call
}

// BaseDomain is a helper method to define mock.On call
func (_e *Environment_Expecter) BaseDomain() *Environment_BaseDomain_Call {
	return &Environment_BaseDomain_Call{Call: _e.mock.On("BaseDomain")}
}

func (_c *Environment_BaseDomain_Call) Run(run func()) *Environment_BaseDomain_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_BaseDomain_Call) Return(_a0 string) *Environment_BaseDomain_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_BaseDomain_Call) RunAndReturn(run func() string) *Environment_BaseDomain_Call {
	_c.Call.Return(run)
	return _c
}

// CreatedAt provides a mock function with given fields:
func (_m *Environment) CreatedAt() time.Time {
	ret := _m.Called()

	var r0 time.Time
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}

// Environment_CreatedAt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreatedAt'
type Environment_CreatedAt_Call struct {
	*mock.Call
}

// CreatedAt is a helper method to define mock.On call
func (_e *Environment_Expecter) CreatedAt() *Environment_CreatedAt_Call {
	return &Environment_CreatedAt_Call{Call: _e.mock.On("CreatedAt")}
}

func (_c *Environment_CreatedAt_Call) Run(run func()) *Environment_CreatedAt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_CreatedAt_Call) Return(_a0 time.Time) *Environment_CreatedAt_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_CreatedAt_Call) RunAndReturn(run func() time.Time) *Environment_CreatedAt_Call {
	_c.Call.Return(run)
	return _c
}

// DefaultCluster provides a mock function with given fields:
func (_m *Environment) DefaultCluster() terra.Cluster {
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

// Environment_DefaultCluster_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DefaultCluster'
type Environment_DefaultCluster_Call struct {
	*mock.Call
}

// DefaultCluster is a helper method to define mock.On call
func (_e *Environment_Expecter) DefaultCluster() *Environment_DefaultCluster_Call {
	return &Environment_DefaultCluster_Call{Call: _e.mock.On("DefaultCluster")}
}

func (_c *Environment_DefaultCluster_Call) Run(run func()) *Environment_DefaultCluster_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_DefaultCluster_Call) Return(_a0 terra.Cluster) *Environment_DefaultCluster_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_DefaultCluster_Call) RunAndReturn(run func() terra.Cluster) *Environment_DefaultCluster_Call {
	_c.Call.Return(run)
	return _c
}

// IsCluster provides a mock function with given fields:
func (_m *Environment) IsCluster() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Environment_IsCluster_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsCluster'
type Environment_IsCluster_Call struct {
	*mock.Call
}

// IsCluster is a helper method to define mock.On call
func (_e *Environment_Expecter) IsCluster() *Environment_IsCluster_Call {
	return &Environment_IsCluster_Call{Call: _e.mock.On("IsCluster")}
}

func (_c *Environment_IsCluster_Call) Run(run func()) *Environment_IsCluster_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_IsCluster_Call) Return(_a0 bool) *Environment_IsCluster_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_IsCluster_Call) RunAndReturn(run func() bool) *Environment_IsCluster_Call {
	_c.Call.Return(run)
	return _c
}

// IsEnvironment provides a mock function with given fields:
func (_m *Environment) IsEnvironment() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Environment_IsEnvironment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsEnvironment'
type Environment_IsEnvironment_Call struct {
	*mock.Call
}

// IsEnvironment is a helper method to define mock.On call
func (_e *Environment_Expecter) IsEnvironment() *Environment_IsEnvironment_Call {
	return &Environment_IsEnvironment_Call{Call: _e.mock.On("IsEnvironment")}
}

func (_c *Environment_IsEnvironment_Call) Run(run func()) *Environment_IsEnvironment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_IsEnvironment_Call) Return(_a0 bool) *Environment_IsEnvironment_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_IsEnvironment_Call) RunAndReturn(run func() bool) *Environment_IsEnvironment_Call {
	_c.Call.Return(run)
	return _c
}

// Lifecycle provides a mock function with given fields:
func (_m *Environment) Lifecycle() terra.Lifecycle {
	ret := _m.Called()

	var r0 terra.Lifecycle
	if rf, ok := ret.Get(0).(func() terra.Lifecycle); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(terra.Lifecycle)
	}

	return r0
}

// Environment_Lifecycle_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Lifecycle'
type Environment_Lifecycle_Call struct {
	*mock.Call
}

// Lifecycle is a helper method to define mock.On call
func (_e *Environment_Expecter) Lifecycle() *Environment_Lifecycle_Call {
	return &Environment_Lifecycle_Call{Call: _e.mock.On("Lifecycle")}
}

func (_c *Environment_Lifecycle_Call) Run(run func()) *Environment_Lifecycle_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_Lifecycle_Call) Return(_a0 terra.Lifecycle) *Environment_Lifecycle_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_Lifecycle_Call) RunAndReturn(run func() terra.Lifecycle) *Environment_Lifecycle_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with given fields:
func (_m *Environment) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Environment_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type Environment_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *Environment_Expecter) Name() *Environment_Name_Call {
	return &Environment_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *Environment_Name_Call) Run(run func()) *Environment_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_Name_Call) Return(_a0 string) *Environment_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_Name_Call) RunAndReturn(run func() string) *Environment_Name_Call {
	_c.Call.Return(run)
	return _c
}

// NamePrefixesDomain provides a mock function with given fields:
func (_m *Environment) NamePrefixesDomain() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Environment_NamePrefixesDomain_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NamePrefixesDomain'
type Environment_NamePrefixesDomain_Call struct {
	*mock.Call
}

// NamePrefixesDomain is a helper method to define mock.On call
func (_e *Environment_Expecter) NamePrefixesDomain() *Environment_NamePrefixesDomain_Call {
	return &Environment_NamePrefixesDomain_Call{Call: _e.mock.On("NamePrefixesDomain")}
}

func (_c *Environment_NamePrefixesDomain_Call) Run(run func()) *Environment_NamePrefixesDomain_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_NamePrefixesDomain_Call) Return(_a0 bool) *Environment_NamePrefixesDomain_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_NamePrefixesDomain_Call) RunAndReturn(run func() bool) *Environment_NamePrefixesDomain_Call {
	_c.Call.Return(run)
	return _c
}

// Namespace provides a mock function with given fields:
func (_m *Environment) Namespace() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Environment_Namespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Namespace'
type Environment_Namespace_Call struct {
	*mock.Call
}

// Namespace is a helper method to define mock.On call
func (_e *Environment_Expecter) Namespace() *Environment_Namespace_Call {
	return &Environment_Namespace_Call{Call: _e.mock.On("Namespace")}
}

func (_c *Environment_Namespace_Call) Run(run func()) *Environment_Namespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_Namespace_Call) Return(_a0 string) *Environment_Namespace_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_Namespace_Call) RunAndReturn(run func() string) *Environment_Namespace_Call {
	_c.Call.Return(run)
	return _c
}

// Offline provides a mock function with given fields:
func (_m *Environment) Offline() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Environment_Offline_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Offline'
type Environment_Offline_Call struct {
	*mock.Call
}

// Offline is a helper method to define mock.On call
func (_e *Environment_Expecter) Offline() *Environment_Offline_Call {
	return &Environment_Offline_Call{Call: _e.mock.On("Offline")}
}

func (_c *Environment_Offline_Call) Run(run func()) *Environment_Offline_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_Offline_Call) Return(_a0 bool) *Environment_Offline_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_Offline_Call) RunAndReturn(run func() bool) *Environment_Offline_Call {
	_c.Call.Return(run)
	return _c
}

// OfflineScheduleBeginEnabled provides a mock function with given fields:
func (_m *Environment) OfflineScheduleBeginEnabled() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Environment_OfflineScheduleBeginEnabled_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OfflineScheduleBeginEnabled'
type Environment_OfflineScheduleBeginEnabled_Call struct {
	*mock.Call
}

// OfflineScheduleBeginEnabled is a helper method to define mock.On call
func (_e *Environment_Expecter) OfflineScheduleBeginEnabled() *Environment_OfflineScheduleBeginEnabled_Call {
	return &Environment_OfflineScheduleBeginEnabled_Call{Call: _e.mock.On("OfflineScheduleBeginEnabled")}
}

func (_c *Environment_OfflineScheduleBeginEnabled_Call) Run(run func()) *Environment_OfflineScheduleBeginEnabled_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_OfflineScheduleBeginEnabled_Call) Return(_a0 bool) *Environment_OfflineScheduleBeginEnabled_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_OfflineScheduleBeginEnabled_Call) RunAndReturn(run func() bool) *Environment_OfflineScheduleBeginEnabled_Call {
	_c.Call.Return(run)
	return _c
}

// OfflineScheduleBeginTime provides a mock function with given fields:
func (_m *Environment) OfflineScheduleBeginTime() time.Time {
	ret := _m.Called()

	var r0 time.Time
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}

// Environment_OfflineScheduleBeginTime_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OfflineScheduleBeginTime'
type Environment_OfflineScheduleBeginTime_Call struct {
	*mock.Call
}

// OfflineScheduleBeginTime is a helper method to define mock.On call
func (_e *Environment_Expecter) OfflineScheduleBeginTime() *Environment_OfflineScheduleBeginTime_Call {
	return &Environment_OfflineScheduleBeginTime_Call{Call: _e.mock.On("OfflineScheduleBeginTime")}
}

func (_c *Environment_OfflineScheduleBeginTime_Call) Run(run func()) *Environment_OfflineScheduleBeginTime_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_OfflineScheduleBeginTime_Call) Return(_a0 time.Time) *Environment_OfflineScheduleBeginTime_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_OfflineScheduleBeginTime_Call) RunAndReturn(run func() time.Time) *Environment_OfflineScheduleBeginTime_Call {
	_c.Call.Return(run)
	return _c
}

// OfflineScheduleEndEnabled provides a mock function with given fields:
func (_m *Environment) OfflineScheduleEndEnabled() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Environment_OfflineScheduleEndEnabled_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OfflineScheduleEndEnabled'
type Environment_OfflineScheduleEndEnabled_Call struct {
	*mock.Call
}

// OfflineScheduleEndEnabled is a helper method to define mock.On call
func (_e *Environment_Expecter) OfflineScheduleEndEnabled() *Environment_OfflineScheduleEndEnabled_Call {
	return &Environment_OfflineScheduleEndEnabled_Call{Call: _e.mock.On("OfflineScheduleEndEnabled")}
}

func (_c *Environment_OfflineScheduleEndEnabled_Call) Run(run func()) *Environment_OfflineScheduleEndEnabled_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_OfflineScheduleEndEnabled_Call) Return(_a0 bool) *Environment_OfflineScheduleEndEnabled_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_OfflineScheduleEndEnabled_Call) RunAndReturn(run func() bool) *Environment_OfflineScheduleEndEnabled_Call {
	_c.Call.Return(run)
	return _c
}

// OfflineScheduleEndTime provides a mock function with given fields:
func (_m *Environment) OfflineScheduleEndTime() time.Time {
	ret := _m.Called()

	var r0 time.Time
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}

// Environment_OfflineScheduleEndTime_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OfflineScheduleEndTime'
type Environment_OfflineScheduleEndTime_Call struct {
	*mock.Call
}

// OfflineScheduleEndTime is a helper method to define mock.On call
func (_e *Environment_Expecter) OfflineScheduleEndTime() *Environment_OfflineScheduleEndTime_Call {
	return &Environment_OfflineScheduleEndTime_Call{Call: _e.mock.On("OfflineScheduleEndTime")}
}

func (_c *Environment_OfflineScheduleEndTime_Call) Run(run func()) *Environment_OfflineScheduleEndTime_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_OfflineScheduleEndTime_Call) Return(_a0 time.Time) *Environment_OfflineScheduleEndTime_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_OfflineScheduleEndTime_Call) RunAndReturn(run func() time.Time) *Environment_OfflineScheduleEndTime_Call {
	_c.Call.Return(run)
	return _c
}

// OfflineScheduleEndWeekends provides a mock function with given fields:
func (_m *Environment) OfflineScheduleEndWeekends() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Environment_OfflineScheduleEndWeekends_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OfflineScheduleEndWeekends'
type Environment_OfflineScheduleEndWeekends_Call struct {
	*mock.Call
}

// OfflineScheduleEndWeekends is a helper method to define mock.On call
func (_e *Environment_Expecter) OfflineScheduleEndWeekends() *Environment_OfflineScheduleEndWeekends_Call {
	return &Environment_OfflineScheduleEndWeekends_Call{Call: _e.mock.On("OfflineScheduleEndWeekends")}
}

func (_c *Environment_OfflineScheduleEndWeekends_Call) Run(run func()) *Environment_OfflineScheduleEndWeekends_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_OfflineScheduleEndWeekends_Call) Return(_a0 bool) *Environment_OfflineScheduleEndWeekends_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_OfflineScheduleEndWeekends_Call) RunAndReturn(run func() bool) *Environment_OfflineScheduleEndWeekends_Call {
	_c.Call.Return(run)
	return _c
}

// Owner provides a mock function with given fields:
func (_m *Environment) Owner() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Environment_Owner_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Owner'
type Environment_Owner_Call struct {
	*mock.Call
}

// Owner is a helper method to define mock.On call
func (_e *Environment_Expecter) Owner() *Environment_Owner_Call {
	return &Environment_Owner_Call{Call: _e.mock.On("Owner")}
}

func (_c *Environment_Owner_Call) Run(run func()) *Environment_Owner_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_Owner_Call) Return(_a0 string) *Environment_Owner_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_Owner_Call) RunAndReturn(run func() string) *Environment_Owner_Call {
	_c.Call.Return(run)
	return _c
}

// PreventDeletion provides a mock function with given fields:
func (_m *Environment) PreventDeletion() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Environment_PreventDeletion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PreventDeletion'
type Environment_PreventDeletion_Call struct {
	*mock.Call
}

// PreventDeletion is a helper method to define mock.On call
func (_e *Environment_Expecter) PreventDeletion() *Environment_PreventDeletion_Call {
	return &Environment_PreventDeletion_Call{Call: _e.mock.On("PreventDeletion")}
}

func (_c *Environment_PreventDeletion_Call) Run(run func()) *Environment_PreventDeletion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_PreventDeletion_Call) Return(_a0 bool) *Environment_PreventDeletion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_PreventDeletion_Call) RunAndReturn(run func() bool) *Environment_PreventDeletion_Call {
	_c.Call.Return(run)
	return _c
}

// ReleaseType provides a mock function with given fields:
func (_m *Environment) ReleaseType() terra.ReleaseType {
	ret := _m.Called()

	var r0 terra.ReleaseType
	if rf, ok := ret.Get(0).(func() terra.ReleaseType); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(terra.ReleaseType)
	}

	return r0
}

// Environment_ReleaseType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReleaseType'
type Environment_ReleaseType_Call struct {
	*mock.Call
}

// ReleaseType is a helper method to define mock.On call
func (_e *Environment_Expecter) ReleaseType() *Environment_ReleaseType_Call {
	return &Environment_ReleaseType_Call{Call: _e.mock.On("ReleaseType")}
}

func (_c *Environment_ReleaseType_Call) Run(run func()) *Environment_ReleaseType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_ReleaseType_Call) Return(_a0 terra.ReleaseType) *Environment_ReleaseType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_ReleaseType_Call) RunAndReturn(run func() terra.ReleaseType) *Environment_ReleaseType_Call {
	_c.Call.Return(run)
	return _c
}

// Releases provides a mock function with given fields:
func (_m *Environment) Releases() []terra.Release {
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

// Environment_Releases_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Releases'
type Environment_Releases_Call struct {
	*mock.Call
}

// Releases is a helper method to define mock.On call
func (_e *Environment_Expecter) Releases() *Environment_Releases_Call {
	return &Environment_Releases_Call{Call: _e.mock.On("Releases")}
}

func (_c *Environment_Releases_Call) Run(run func()) *Environment_Releases_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_Releases_Call) Return(_a0 []terra.Release) *Environment_Releases_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_Releases_Call) RunAndReturn(run func() []terra.Release) *Environment_Releases_Call {
	_c.Call.Return(run)
	return _c
}

// RequireSuitable provides a mock function with given fields:
func (_m *Environment) RequireSuitable() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Environment_RequireSuitable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RequireSuitable'
type Environment_RequireSuitable_Call struct {
	*mock.Call
}

// RequireSuitable is a helper method to define mock.On call
func (_e *Environment_Expecter) RequireSuitable() *Environment_RequireSuitable_Call {
	return &Environment_RequireSuitable_Call{Call: _e.mock.On("RequireSuitable")}
}

func (_c *Environment_RequireSuitable_Call) Run(run func()) *Environment_RequireSuitable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_RequireSuitable_Call) Return(_a0 bool) *Environment_RequireSuitable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_RequireSuitable_Call) RunAndReturn(run func() bool) *Environment_RequireSuitable_Call {
	_c.Call.Return(run)
	return _c
}

// Template provides a mock function with given fields:
func (_m *Environment) Template() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Environment_Template_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Template'
type Environment_Template_Call struct {
	*mock.Call
}

// Template is a helper method to define mock.On call
func (_e *Environment_Expecter) Template() *Environment_Template_Call {
	return &Environment_Template_Call{Call: _e.mock.On("Template")}
}

func (_c *Environment_Template_Call) Run(run func()) *Environment_Template_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_Template_Call) Return(_a0 string) *Environment_Template_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_Template_Call) RunAndReturn(run func() string) *Environment_Template_Call {
	_c.Call.Return(run)
	return _c
}

// TerraHelmfileRef provides a mock function with given fields:
func (_m *Environment) TerraHelmfileRef() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Environment_TerraHelmfileRef_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TerraHelmfileRef'
type Environment_TerraHelmfileRef_Call struct {
	*mock.Call
}

// TerraHelmfileRef is a helper method to define mock.On call
func (_e *Environment_Expecter) TerraHelmfileRef() *Environment_TerraHelmfileRef_Call {
	return &Environment_TerraHelmfileRef_Call{Call: _e.mock.On("TerraHelmfileRef")}
}

func (_c *Environment_TerraHelmfileRef_Call) Run(run func()) *Environment_TerraHelmfileRef_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_TerraHelmfileRef_Call) Return(_a0 string) *Environment_TerraHelmfileRef_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_TerraHelmfileRef_Call) RunAndReturn(run func() string) *Environment_TerraHelmfileRef_Call {
	_c.Call.Return(run)
	return _c
}

// Type provides a mock function with given fields:
func (_m *Environment) Type() terra.DestinationType {
	ret := _m.Called()

	var r0 terra.DestinationType
	if rf, ok := ret.Get(0).(func() terra.DestinationType); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(terra.DestinationType)
	}

	return r0
}

// Environment_Type_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Type'
type Environment_Type_Call struct {
	*mock.Call
}

// Type is a helper method to define mock.On call
func (_e *Environment_Expecter) Type() *Environment_Type_Call {
	return &Environment_Type_Call{Call: _e.mock.On("Type")}
}

func (_c *Environment_Type_Call) Run(run func()) *Environment_Type_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_Type_Call) Return(_a0 terra.DestinationType) *Environment_Type_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_Type_Call) RunAndReturn(run func() terra.DestinationType) *Environment_Type_Call {
	_c.Call.Return(run)
	return _c
}

// UniqueResourcePrefix provides a mock function with given fields:
func (_m *Environment) UniqueResourcePrefix() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Environment_UniqueResourcePrefix_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UniqueResourcePrefix'
type Environment_UniqueResourcePrefix_Call struct {
	*mock.Call
}

// UniqueResourcePrefix is a helper method to define mock.On call
func (_e *Environment_Expecter) UniqueResourcePrefix() *Environment_UniqueResourcePrefix_Call {
	return &Environment_UniqueResourcePrefix_Call{Call: _e.mock.On("UniqueResourcePrefix")}
}

func (_c *Environment_UniqueResourcePrefix_Call) Run(run func()) *Environment_UniqueResourcePrefix_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Environment_UniqueResourcePrefix_Call) Return(_a0 string) *Environment_UniqueResourcePrefix_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Environment_UniqueResourcePrefix_Call) RunAndReturn(run func() string) *Environment_UniqueResourcePrefix_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewEnvironment interface {
	mock.TestingT
	Cleanup(func())
}

// NewEnvironment creates a new instance of Environment. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewEnvironment(t mockConstructorTestingTNewEnvironment) *Environment {
	mock := &Environment{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
