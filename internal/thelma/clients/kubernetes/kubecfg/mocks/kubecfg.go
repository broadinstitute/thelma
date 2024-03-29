// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	kubecfg "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/kubecfg"
	mock "github.com/stretchr/testify/mock"

	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

// Kubeconfig is an autogenerated mock type for the Kubeconfig type
type Kubeconfig struct {
	mock.Mock
}

type Kubeconfig_Expecter struct {
	mock *mock.Mock
}

func (_m *Kubeconfig) EXPECT() *Kubeconfig_Expecter {
	return &Kubeconfig_Expecter{mock: &_m.Mock}
}

// ConfigFile provides a mock function with given fields:
func (_m *Kubeconfig) ConfigFile() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Kubeconfig_ConfigFile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ConfigFile'
type Kubeconfig_ConfigFile_Call struct {
	*mock.Call
}

// ConfigFile is a helper method to define mock.On call
func (_e *Kubeconfig_Expecter) ConfigFile() *Kubeconfig_ConfigFile_Call {
	return &Kubeconfig_ConfigFile_Call{Call: _e.mock.On("ConfigFile")}
}

func (_c *Kubeconfig_ConfigFile_Call) Run(run func()) *Kubeconfig_ConfigFile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Kubeconfig_ConfigFile_Call) Return(_a0 string) *Kubeconfig_ConfigFile_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Kubeconfig_ConfigFile_Call) RunAndReturn(run func() string) *Kubeconfig_ConfigFile_Call {
	_c.Call.Return(run)
	return _c
}

// ForCluster provides a mock function with given fields: cluster
func (_m *Kubeconfig) ForCluster(cluster terra.Cluster) (kubecfg.Kubectx, error) {
	ret := _m.Called(cluster)

	var r0 kubecfg.Kubectx
	var r1 error
	if rf, ok := ret.Get(0).(func(terra.Cluster) (kubecfg.Kubectx, error)); ok {
		return rf(cluster)
	}
	if rf, ok := ret.Get(0).(func(terra.Cluster) kubecfg.Kubectx); ok {
		r0 = rf(cluster)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(kubecfg.Kubectx)
		}
	}

	if rf, ok := ret.Get(1).(func(terra.Cluster) error); ok {
		r1 = rf(cluster)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Kubeconfig_ForCluster_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ForCluster'
type Kubeconfig_ForCluster_Call struct {
	*mock.Call
}

// ForCluster is a helper method to define mock.On call
//   - cluster terra.Cluster
func (_e *Kubeconfig_Expecter) ForCluster(cluster interface{}) *Kubeconfig_ForCluster_Call {
	return &Kubeconfig_ForCluster_Call{Call: _e.mock.On("ForCluster", cluster)}
}

func (_c *Kubeconfig_ForCluster_Call) Run(run func(cluster terra.Cluster)) *Kubeconfig_ForCluster_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(terra.Cluster))
	})
	return _c
}

func (_c *Kubeconfig_ForCluster_Call) Return(_a0 kubecfg.Kubectx, _a1 error) *Kubeconfig_ForCluster_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Kubeconfig_ForCluster_Call) RunAndReturn(run func(terra.Cluster) (kubecfg.Kubectx, error)) *Kubeconfig_ForCluster_Call {
	_c.Call.Return(run)
	return _c
}

// ForEnvironment provides a mock function with given fields: env
func (_m *Kubeconfig) ForEnvironment(env terra.Environment) ([]kubecfg.Kubectx, error) {
	ret := _m.Called(env)

	var r0 []kubecfg.Kubectx
	var r1 error
	if rf, ok := ret.Get(0).(func(terra.Environment) ([]kubecfg.Kubectx, error)); ok {
		return rf(env)
	}
	if rf, ok := ret.Get(0).(func(terra.Environment) []kubecfg.Kubectx); ok {
		r0 = rf(env)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]kubecfg.Kubectx)
		}
	}

	if rf, ok := ret.Get(1).(func(terra.Environment) error); ok {
		r1 = rf(env)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Kubeconfig_ForEnvironment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ForEnvironment'
type Kubeconfig_ForEnvironment_Call struct {
	*mock.Call
}

// ForEnvironment is a helper method to define mock.On call
//   - env terra.Environment
func (_e *Kubeconfig_Expecter) ForEnvironment(env interface{}) *Kubeconfig_ForEnvironment_Call {
	return &Kubeconfig_ForEnvironment_Call{Call: _e.mock.On("ForEnvironment", env)}
}

func (_c *Kubeconfig_ForEnvironment_Call) Run(run func(env terra.Environment)) *Kubeconfig_ForEnvironment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(terra.Environment))
	})
	return _c
}

func (_c *Kubeconfig_ForEnvironment_Call) Return(_a0 []kubecfg.Kubectx, _a1 error) *Kubeconfig_ForEnvironment_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Kubeconfig_ForEnvironment_Call) RunAndReturn(run func(terra.Environment) ([]kubecfg.Kubectx, error)) *Kubeconfig_ForEnvironment_Call {
	_c.Call.Return(run)
	return _c
}

// ForRelease provides a mock function with given fields: _a0
func (_m *Kubeconfig) ForRelease(_a0 terra.Release) (kubecfg.Kubectx, error) {
	ret := _m.Called(_a0)

	var r0 kubecfg.Kubectx
	var r1 error
	if rf, ok := ret.Get(0).(func(terra.Release) (kubecfg.Kubectx, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(terra.Release) kubecfg.Kubectx); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(kubecfg.Kubectx)
		}
	}

	if rf, ok := ret.Get(1).(func(terra.Release) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Kubeconfig_ForRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ForRelease'
type Kubeconfig_ForRelease_Call struct {
	*mock.Call
}

// ForRelease is a helper method to define mock.On call
//   - _a0 terra.Release
func (_e *Kubeconfig_Expecter) ForRelease(_a0 interface{}) *Kubeconfig_ForRelease_Call {
	return &Kubeconfig_ForRelease_Call{Call: _e.mock.On("ForRelease", _a0)}
}

func (_c *Kubeconfig_ForRelease_Call) Run(run func(_a0 terra.Release)) *Kubeconfig_ForRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(terra.Release))
	})
	return _c
}

func (_c *Kubeconfig_ForRelease_Call) Return(_a0 kubecfg.Kubectx, _a1 error) *Kubeconfig_ForRelease_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Kubeconfig_ForRelease_Call) RunAndReturn(run func(terra.Release) (kubecfg.Kubectx, error)) *Kubeconfig_ForRelease_Call {
	_c.Call.Return(run)
	return _c
}

// ForReleases provides a mock function with given fields: releases
func (_m *Kubeconfig) ForReleases(releases ...terra.Release) ([]kubecfg.ReleaseKtx, error) {
	_va := make([]interface{}, len(releases))
	for _i := range releases {
		_va[_i] = releases[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []kubecfg.ReleaseKtx
	var r1 error
	if rf, ok := ret.Get(0).(func(...terra.Release) ([]kubecfg.ReleaseKtx, error)); ok {
		return rf(releases...)
	}
	if rf, ok := ret.Get(0).(func(...terra.Release) []kubecfg.ReleaseKtx); ok {
		r0 = rf(releases...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]kubecfg.ReleaseKtx)
		}
	}

	if rf, ok := ret.Get(1).(func(...terra.Release) error); ok {
		r1 = rf(releases...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Kubeconfig_ForReleases_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ForReleases'
type Kubeconfig_ForReleases_Call struct {
	*mock.Call
}

// ForReleases is a helper method to define mock.On call
//   - releases ...terra.Release
func (_e *Kubeconfig_Expecter) ForReleases(releases ...interface{}) *Kubeconfig_ForReleases_Call {
	return &Kubeconfig_ForReleases_Call{Call: _e.mock.On("ForReleases",
		append([]interface{}{}, releases...)...)}
}

func (_c *Kubeconfig_ForReleases_Call) Run(run func(releases ...terra.Release)) *Kubeconfig_ForReleases_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]terra.Release, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(terra.Release)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *Kubeconfig_ForReleases_Call) Return(_a0 []kubecfg.ReleaseKtx, _a1 error) *Kubeconfig_ForReleases_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Kubeconfig_ForReleases_Call) RunAndReturn(run func(...terra.Release) ([]kubecfg.ReleaseKtx, error)) *Kubeconfig_ForReleases_Call {
	_c.Call.Return(run)
	return _c
}

// NewKubeconfig creates a new instance of Kubeconfig. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewKubeconfig(t interface {
	mock.TestingT
	Cleanup(func())
}) *Kubeconfig {
	mock := &Kubeconfig{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
