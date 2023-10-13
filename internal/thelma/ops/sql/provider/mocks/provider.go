// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	api "github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	dbms "github.com/broadinstitute/thelma/internal/thelma/ops/sql/dbms"

	mock "github.com/stretchr/testify/mock"

	podrun "github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"

	provider "github.com/broadinstitute/thelma/internal/thelma/ops/sql/provider"
)

// Provider is an autogenerated mock type for the Provider type
type Provider struct {
	mock.Mock
}

type Provider_Expecter struct {
	mock *mock.Mock
}

func (_m *Provider) EXPECT() *Provider_Expecter {
	return &Provider_Expecter{mock: &_m.Mock}
}

// ClientSettings provides a mock function with given fields: _a0
func (_m *Provider) ClientSettings(_a0 ...provider.ConnectionOverride) (dbms.ClientSettings, error) {
	_va := make([]interface{}, len(_a0))
	for _i := range _a0 {
		_va[_i] = _a0[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 dbms.ClientSettings
	var r1 error
	if rf, ok := ret.Get(0).(func(...provider.ConnectionOverride) (dbms.ClientSettings, error)); ok {
		return rf(_a0...)
	}
	if rf, ok := ret.Get(0).(func(...provider.ConnectionOverride) dbms.ClientSettings); ok {
		r0 = rf(_a0...)
	} else {
		r0 = ret.Get(0).(dbms.ClientSettings)
	}

	if rf, ok := ret.Get(1).(func(...provider.ConnectionOverride) error); ok {
		r1 = rf(_a0...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Provider_ClientSettings_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClientSettings'
type Provider_ClientSettings_Call struct {
	*mock.Call
}

// ClientSettings is a helper method to define mock.On call
//   - _a0 ...provider.ConnectionOverride
func (_e *Provider_Expecter) ClientSettings(_a0 ...interface{}) *Provider_ClientSettings_Call {
	return &Provider_ClientSettings_Call{Call: _e.mock.On("ClientSettings",
		append([]interface{}{}, _a0...)...)}
}

func (_c *Provider_ClientSettings_Call) Run(run func(_a0 ...provider.ConnectionOverride)) *Provider_ClientSettings_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]provider.ConnectionOverride, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(provider.ConnectionOverride)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *Provider_ClientSettings_Call) Return(_a0 dbms.ClientSettings, _a1 error) *Provider_ClientSettings_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Provider_ClientSettings_Call) RunAndReturn(run func(...provider.ConnectionOverride) (dbms.ClientSettings, error)) *Provider_ClientSettings_Call {
	_c.Call.Return(run)
	return _c
}

// DetectDBMS provides a mock function with given fields:
func (_m *Provider) DetectDBMS() (api.DBMS, error) {
	ret := _m.Called()

	var r0 api.DBMS
	var r1 error
	if rf, ok := ret.Get(0).(func() (api.DBMS, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() api.DBMS); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(api.DBMS)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Provider_DetectDBMS_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DetectDBMS'
type Provider_DetectDBMS_Call struct {
	*mock.Call
}

// DetectDBMS is a helper method to define mock.On call
func (_e *Provider_Expecter) DetectDBMS() *Provider_DetectDBMS_Call {
	return &Provider_DetectDBMS_Call{Call: _e.mock.On("DetectDBMS")}
}

func (_c *Provider_DetectDBMS_Call) Run(run func()) *Provider_DetectDBMS_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Provider_DetectDBMS_Call) Return(_a0 api.DBMS, _a1 error) *Provider_DetectDBMS_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Provider_DetectDBMS_Call) RunAndReturn(run func() (api.DBMS, error)) *Provider_DetectDBMS_Call {
	_c.Call.Return(run)
	return _c
}

// Initialize provides a mock function with given fields:
func (_m *Provider) Initialize() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Provider_Initialize_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Initialize'
type Provider_Initialize_Call struct {
	*mock.Call
}

// Initialize is a helper method to define mock.On call
func (_e *Provider_Expecter) Initialize() *Provider_Initialize_Call {
	return &Provider_Initialize_Call{Call: _e.mock.On("Initialize")}
}

func (_c *Provider_Initialize_Call) Run(run func()) *Provider_Initialize_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Provider_Initialize_Call) Return(_a0 error) *Provider_Initialize_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Provider_Initialize_Call) RunAndReturn(run func() error) *Provider_Initialize_Call {
	_c.Call.Return(run)
	return _c
}

// Initialized provides a mock function with given fields:
func (_m *Provider) Initialized() (bool, error) {
	ret := _m.Called()

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func() (bool, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Provider_Initialized_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Initialized'
type Provider_Initialized_Call struct {
	*mock.Call
}

// Initialized is a helper method to define mock.On call
func (_e *Provider_Expecter) Initialized() *Provider_Initialized_Call {
	return &Provider_Initialized_Call{Call: _e.mock.On("Initialized")}
}

func (_c *Provider_Initialized_Call) Run(run func()) *Provider_Initialized_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Provider_Initialized_Call) Return(_a0 bool, _a1 error) *Provider_Initialized_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Provider_Initialized_Call) RunAndReturn(run func() (bool, error)) *Provider_Initialized_Call {
	_c.Call.Return(run)
	return _c
}

// PodSpec provides a mock function with given fields: _a0
func (_m *Provider) PodSpec(_a0 ...provider.ConnectionOverride) (podrun.ProviderSpec, error) {
	_va := make([]interface{}, len(_a0))
	for _i := range _a0 {
		_va[_i] = _a0[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 podrun.ProviderSpec
	var r1 error
	if rf, ok := ret.Get(0).(func(...provider.ConnectionOverride) (podrun.ProviderSpec, error)); ok {
		return rf(_a0...)
	}
	if rf, ok := ret.Get(0).(func(...provider.ConnectionOverride) podrun.ProviderSpec); ok {
		r0 = rf(_a0...)
	} else {
		r0 = ret.Get(0).(podrun.ProviderSpec)
	}

	if rf, ok := ret.Get(1).(func(...provider.ConnectionOverride) error); ok {
		r1 = rf(_a0...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Provider_PodSpec_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PodSpec'
type Provider_PodSpec_Call struct {
	*mock.Call
}

// PodSpec is a helper method to define mock.On call
//   - _a0 ...provider.ConnectionOverride
func (_e *Provider_Expecter) PodSpec(_a0 ...interface{}) *Provider_PodSpec_Call {
	return &Provider_PodSpec_Call{Call: _e.mock.On("PodSpec",
		append([]interface{}{}, _a0...)...)}
}

func (_c *Provider_PodSpec_Call) Run(run func(_a0 ...provider.ConnectionOverride)) *Provider_PodSpec_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]provider.ConnectionOverride, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(provider.ConnectionOverride)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *Provider_PodSpec_Call) Return(_a0 podrun.ProviderSpec, _a1 error) *Provider_PodSpec_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Provider_PodSpec_Call) RunAndReturn(run func(...provider.ConnectionOverride) (podrun.ProviderSpec, error)) *Provider_PodSpec_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewProvider interface {
	mock.TestingT
	Cleanup(func())
}

// NewProvider creates a new instance of Provider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewProvider(t mockConstructorTestingTNewProvider) *Provider {
	mock := &Provider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
