// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	sherlock "github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	mock "github.com/stretchr/testify/mock"
)

// StateReadWriter is an autogenerated mock type for the StateReadWriter type
type StateReadWriter struct {
	mock.Mock
}

type StateReadWriter_Expecter struct {
	mock *mock.Mock
}

func (_m *StateReadWriter) EXPECT() *StateReadWriter_Expecter {
	return &StateReadWriter_Expecter{mock: &_m.Mock}
}

// Clusters provides a mock function with given fields:
func (_m *StateReadWriter) Clusters() (sherlock.Clusters, error) {
	ret := _m.Called()

	var r0 sherlock.Clusters
	if rf, ok := ret.Get(0).(func() sherlock.Clusters); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sherlock.Clusters)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StateReadWriter_Clusters_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Clusters'
type StateReadWriter_Clusters_Call struct {
	*mock.Call
}

// Clusters is a helper method to define mock.On call
func (_e *StateReadWriter_Expecter) Clusters() *StateReadWriter_Clusters_Call {
	return &StateReadWriter_Clusters_Call{Call: _e.mock.On("Clusters")}
}

func (_c *StateReadWriter_Clusters_Call) Run(run func()) *StateReadWriter_Clusters_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *StateReadWriter_Clusters_Call) Return(_a0 sherlock.Clusters, _a1 error) *StateReadWriter_Clusters_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// CreateEnvironmentFromTemplate provides a mock function with given fields: templateName, desiredNamePrefix, desiredName, desiredOwnerEmail
func (_m *StateReadWriter) CreateEnvironmentFromTemplate(templateName string, desiredNamePrefix string, desiredName string, desiredOwnerEmail string) (string, error) {
	ret := _m.Called(templateName, desiredNamePrefix, desiredName, desiredOwnerEmail)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string, string, string) string); ok {
		r0 = rf(templateName, desiredNamePrefix, desiredName, desiredOwnerEmail)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string) error); ok {
		r1 = rf(templateName, desiredNamePrefix, desiredName, desiredOwnerEmail)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StateReadWriter_CreateEnvironmentFromTemplate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateEnvironmentFromTemplate'
type StateReadWriter_CreateEnvironmentFromTemplate_Call struct {
	*mock.Call
}

// CreateEnvironmentFromTemplate is a helper method to define mock.On call
//   - templateName string
//   - desiredNamePrefix string
//   - desiredName string
//   - desiredOwnerEmail string
func (_e *StateReadWriter_Expecter) CreateEnvironmentFromTemplate(templateName interface{}, desiredNamePrefix interface{}, desiredName interface{}, desiredOwnerEmail interface{}) *StateReadWriter_CreateEnvironmentFromTemplate_Call {
	return &StateReadWriter_CreateEnvironmentFromTemplate_Call{Call: _e.mock.On("CreateEnvironmentFromTemplate", templateName, desiredNamePrefix, desiredName, desiredOwnerEmail)}
}

func (_c *StateReadWriter_CreateEnvironmentFromTemplate_Call) Run(run func(templateName string, desiredNamePrefix string, desiredName string, desiredOwnerEmail string)) *StateReadWriter_CreateEnvironmentFromTemplate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *StateReadWriter_CreateEnvironmentFromTemplate_Call) Return(_a0 string, _a1 error) *StateReadWriter_CreateEnvironmentFromTemplate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// DeleteEnvironments provides a mock function with given fields: _a0
func (_m *StateReadWriter) DeleteEnvironments(_a0 []terra.Environment) ([]string, error) {
	ret := _m.Called(_a0)

	var r0 []string
	if rf, ok := ret.Get(0).(func([]terra.Environment) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]terra.Environment) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StateReadWriter_DeleteEnvironments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteEnvironments'
type StateReadWriter_DeleteEnvironments_Call struct {
	*mock.Call
}

// DeleteEnvironments is a helper method to define mock.On call
//   - _a0 []terra.Environment
func (_e *StateReadWriter_Expecter) DeleteEnvironments(_a0 interface{}) *StateReadWriter_DeleteEnvironments_Call {
	return &StateReadWriter_DeleteEnvironments_Call{Call: _e.mock.On("DeleteEnvironments", _a0)}
}

func (_c *StateReadWriter_DeleteEnvironments_Call) Run(run func(_a0 []terra.Environment)) *StateReadWriter_DeleteEnvironments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]terra.Environment))
	})
	return _c
}

func (_c *StateReadWriter_DeleteEnvironments_Call) Return(_a0 []string, _a1 error) *StateReadWriter_DeleteEnvironments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// DisableRelease provides a mock function with given fields: _a0, _a1
func (_m *StateReadWriter) DisableRelease(_a0 string, _a1 string) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StateReadWriter_DisableRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DisableRelease'
type StateReadWriter_DisableRelease_Call struct {
	*mock.Call
}

// DisableRelease is a helper method to define mock.On call
//   - _a0 string
//   - _a1 string
func (_e *StateReadWriter_Expecter) DisableRelease(_a0 interface{}, _a1 interface{}) *StateReadWriter_DisableRelease_Call {
	return &StateReadWriter_DisableRelease_Call{Call: _e.mock.On("DisableRelease", _a0, _a1)}
}

func (_c *StateReadWriter_DisableRelease_Call) Run(run func(_a0 string, _a1 string)) *StateReadWriter_DisableRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *StateReadWriter_DisableRelease_Call) Return(_a0 error) *StateReadWriter_DisableRelease_Call {
	_c.Call.Return(_a0)
	return _c
}

// EnableRelease provides a mock function with given fields: _a0, _a1
func (_m *StateReadWriter) EnableRelease(_a0 terra.Environment, _a1 string) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(terra.Environment, string) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StateReadWriter_EnableRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EnableRelease'
type StateReadWriter_EnableRelease_Call struct {
	*mock.Call
}

// EnableRelease is a helper method to define mock.On call
//   - _a0 terra.Environment
//   - _a1 string
func (_e *StateReadWriter_Expecter) EnableRelease(_a0 interface{}, _a1 interface{}) *StateReadWriter_EnableRelease_Call {
	return &StateReadWriter_EnableRelease_Call{Call: _e.mock.On("EnableRelease", _a0, _a1)}
}

func (_c *StateReadWriter_EnableRelease_Call) Run(run func(_a0 terra.Environment, _a1 string)) *StateReadWriter_EnableRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(terra.Environment), args[1].(string))
	})
	return _c
}

func (_c *StateReadWriter_EnableRelease_Call) Return(_a0 error) *StateReadWriter_EnableRelease_Call {
	_c.Call.Return(_a0)
	return _c
}

// Environments provides a mock function with given fields:
func (_m *StateReadWriter) Environments() (sherlock.Environments, error) {
	ret := _m.Called()

	var r0 sherlock.Environments
	if rf, ok := ret.Get(0).(func() sherlock.Environments); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sherlock.Environments)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StateReadWriter_Environments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Environments'
type StateReadWriter_Environments_Call struct {
	*mock.Call
}

// Environments is a helper method to define mock.On call
func (_e *StateReadWriter_Expecter) Environments() *StateReadWriter_Environments_Call {
	return &StateReadWriter_Environments_Call{Call: _e.mock.On("Environments")}
}

func (_c *StateReadWriter_Environments_Call) Run(run func()) *StateReadWriter_Environments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *StateReadWriter_Environments_Call) Return(_a0 sherlock.Environments, _a1 error) *StateReadWriter_Environments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// PinEnvironmentVersions provides a mock function with given fields: environmentName, versions
func (_m *StateReadWriter) PinEnvironmentVersions(environmentName string, versions map[string]terra.VersionOverride) error {
	ret := _m.Called(environmentName, versions)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, map[string]terra.VersionOverride) error); ok {
		r0 = rf(environmentName, versions)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StateReadWriter_PinEnvironmentVersions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PinEnvironmentVersions'
type StateReadWriter_PinEnvironmentVersions_Call struct {
	*mock.Call
}

// PinEnvironmentVersions is a helper method to define mock.On call
//   - environmentName string
//   - versions map[string]terra.VersionOverride
func (_e *StateReadWriter_Expecter) PinEnvironmentVersions(environmentName interface{}, versions interface{}) *StateReadWriter_PinEnvironmentVersions_Call {
	return &StateReadWriter_PinEnvironmentVersions_Call{Call: _e.mock.On("PinEnvironmentVersions", environmentName, versions)}
}

func (_c *StateReadWriter_PinEnvironmentVersions_Call) Run(run func(environmentName string, versions map[string]terra.VersionOverride)) *StateReadWriter_PinEnvironmentVersions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(map[string]terra.VersionOverride))
	})
	return _c
}

func (_c *StateReadWriter_PinEnvironmentVersions_Call) Return(_a0 error) *StateReadWriter_PinEnvironmentVersions_Call {
	_c.Call.Return(_a0)
	return _c
}

// Releases provides a mock function with given fields:
func (_m *StateReadWriter) Releases() (sherlock.Releases, error) {
	ret := _m.Called()

	var r0 sherlock.Releases
	if rf, ok := ret.Get(0).(func() sherlock.Releases); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sherlock.Releases)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StateReadWriter_Releases_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Releases'
type StateReadWriter_Releases_Call struct {
	*mock.Call
}

// Releases is a helper method to define mock.On call
func (_e *StateReadWriter_Expecter) Releases() *StateReadWriter_Releases_Call {
	return &StateReadWriter_Releases_Call{Call: _e.mock.On("Releases")}
}

func (_c *StateReadWriter_Releases_Call) Run(run func()) *StateReadWriter_Releases_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *StateReadWriter_Releases_Call) Return(_a0 sherlock.Releases, _a1 error) *StateReadWriter_Releases_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// ResetEnvironmentAndPinToDev provides a mock function with given fields: environment
func (_m *StateReadWriter) ResetEnvironmentAndPinToDev(environment terra.Environment) error {
	ret := _m.Called(environment)

	var r0 error
	if rf, ok := ret.Get(0).(func(terra.Environment) error); ok {
		r0 = rf(environment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StateReadWriter_ResetEnvironmentAndPinToDev_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResetEnvironmentAndPinToDev'
type StateReadWriter_ResetEnvironmentAndPinToDev_Call struct {
	*mock.Call
}

// ResetEnvironmentAndPinToDev is a helper method to define mock.On call
//   - environment terra.Environment
func (_e *StateReadWriter_Expecter) ResetEnvironmentAndPinToDev(environment interface{}) *StateReadWriter_ResetEnvironmentAndPinToDev_Call {
	return &StateReadWriter_ResetEnvironmentAndPinToDev_Call{Call: _e.mock.On("ResetEnvironmentAndPinToDev", environment)}
}

func (_c *StateReadWriter_ResetEnvironmentAndPinToDev_Call) Run(run func(environment terra.Environment)) *StateReadWriter_ResetEnvironmentAndPinToDev_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(terra.Environment))
	})
	return _c
}

func (_c *StateReadWriter_ResetEnvironmentAndPinToDev_Call) Return(_a0 error) *StateReadWriter_ResetEnvironmentAndPinToDev_Call {
	_c.Call.Return(_a0)
	return _c
}

// SetTerraHelmfileRefForEntireEnvironment provides a mock function with given fields: environment, terraHelmfileRef
func (_m *StateReadWriter) SetTerraHelmfileRefForEntireEnvironment(environment terra.Environment, terraHelmfileRef string) error {
	ret := _m.Called(environment, terraHelmfileRef)

	var r0 error
	if rf, ok := ret.Get(0).(func(terra.Environment, string) error); ok {
		r0 = rf(environment, terraHelmfileRef)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StateReadWriter_SetTerraHelmfileRefForEntireEnvironment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetTerraHelmfileRefForEntireEnvironment'
type StateReadWriter_SetTerraHelmfileRefForEntireEnvironment_Call struct {
	*mock.Call
}

// SetTerraHelmfileRefForEntireEnvironment is a helper method to define mock.On call
//   - environment terra.Environment
//   - terraHelmfileRef string
func (_e *StateReadWriter_Expecter) SetTerraHelmfileRefForEntireEnvironment(environment interface{}, terraHelmfileRef interface{}) *StateReadWriter_SetTerraHelmfileRefForEntireEnvironment_Call {
	return &StateReadWriter_SetTerraHelmfileRefForEntireEnvironment_Call{Call: _e.mock.On("SetTerraHelmfileRefForEntireEnvironment", environment, terraHelmfileRef)}
}

func (_c *StateReadWriter_SetTerraHelmfileRefForEntireEnvironment_Call) Run(run func(environment terra.Environment, terraHelmfileRef string)) *StateReadWriter_SetTerraHelmfileRefForEntireEnvironment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(terra.Environment), args[1].(string))
	})
	return _c
}

func (_c *StateReadWriter_SetTerraHelmfileRefForEntireEnvironment_Call) Return(_a0 error) *StateReadWriter_SetTerraHelmfileRefForEntireEnvironment_Call {
	_c.Call.Return(_a0)
	return _c
}

// WriteClusters provides a mock function with given fields: _a0
func (_m *StateReadWriter) WriteClusters(_a0 []terra.Cluster) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func([]terra.Cluster) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StateReadWriter_WriteClusters_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteClusters'
type StateReadWriter_WriteClusters_Call struct {
	*mock.Call
}

// WriteClusters is a helper method to define mock.On call
//   - _a0 []terra.Cluster
func (_e *StateReadWriter_Expecter) WriteClusters(_a0 interface{}) *StateReadWriter_WriteClusters_Call {
	return &StateReadWriter_WriteClusters_Call{Call: _e.mock.On("WriteClusters", _a0)}
}

func (_c *StateReadWriter_WriteClusters_Call) Run(run func(_a0 []terra.Cluster)) *StateReadWriter_WriteClusters_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]terra.Cluster))
	})
	return _c
}

func (_c *StateReadWriter_WriteClusters_Call) Return(_a0 error) *StateReadWriter_WriteClusters_Call {
	_c.Call.Return(_a0)
	return _c
}

// WriteEnvironments provides a mock function with given fields: _a0
func (_m *StateReadWriter) WriteEnvironments(_a0 []terra.Environment) ([]string, error) {
	ret := _m.Called(_a0)

	var r0 []string
	if rf, ok := ret.Get(0).(func([]terra.Environment) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]terra.Environment) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StateReadWriter_WriteEnvironments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteEnvironments'
type StateReadWriter_WriteEnvironments_Call struct {
	*mock.Call
}

// WriteEnvironments is a helper method to define mock.On call
//   - _a0 []terra.Environment
func (_e *StateReadWriter_Expecter) WriteEnvironments(_a0 interface{}) *StateReadWriter_WriteEnvironments_Call {
	return &StateReadWriter_WriteEnvironments_Call{Call: _e.mock.On("WriteEnvironments", _a0)}
}

func (_c *StateReadWriter_WriteEnvironments_Call) Run(run func(_a0 []terra.Environment)) *StateReadWriter_WriteEnvironments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]terra.Environment))
	})
	return _c
}

func (_c *StateReadWriter_WriteEnvironments_Call) Return(_a0 []string, _a1 error) *StateReadWriter_WriteEnvironments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewStateReadWriter interface {
	mock.TestingT
	Cleanup(func())
}

// NewStateReadWriter creates a new instance of StateReadWriter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewStateReadWriter(t mockConstructorTestingTNewStateReadWriter) *StateReadWriter {
	mock := &StateReadWriter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
