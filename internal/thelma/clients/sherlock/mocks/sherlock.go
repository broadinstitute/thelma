// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	sherlock "github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	mock "github.com/stretchr/testify/mock"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

type Client_Expecter struct {
	mock *mock.Mock
}

func (_m *Client) EXPECT() *Client_Expecter {
	return &Client_Expecter{mock: &_m.Mock}
}

// Clusters provides a mock function with given fields:
func (_m *Client) Clusters() (sherlock.Clusters, error) {
	ret := _m.Called()

	var r0 sherlock.Clusters
	var r1 error
	if rf, ok := ret.Get(0).(func() (sherlock.Clusters, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() sherlock.Clusters); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sherlock.Clusters)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_Clusters_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Clusters'
type Client_Clusters_Call struct {
	*mock.Call
}

// Clusters is a helper method to define mock.On call
func (_e *Client_Expecter) Clusters() *Client_Clusters_Call {
	return &Client_Clusters_Call{Call: _e.mock.On("Clusters")}
}

func (_c *Client_Clusters_Call) Run(run func()) *Client_Clusters_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Client_Clusters_Call) Return(_a0 sherlock.Clusters, _a1 error) *Client_Clusters_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Client_Clusters_Call) RunAndReturn(run func() (sherlock.Clusters, error)) *Client_Clusters_Call {
	_c.Call.Return(run)
	return _c
}

// CreateEnvironmentFromTemplate provides a mock function with given fields: templateName, options
func (_m *Client) CreateEnvironmentFromTemplate(templateName string, options terra.CreateOptions) (string, error) {
	ret := _m.Called(templateName, options)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, terra.CreateOptions) (string, error)); ok {
		return rf(templateName, options)
	}
	if rf, ok := ret.Get(0).(func(string, terra.CreateOptions) string); ok {
		r0 = rf(templateName, options)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, terra.CreateOptions) error); ok {
		r1 = rf(templateName, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_CreateEnvironmentFromTemplate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateEnvironmentFromTemplate'
type Client_CreateEnvironmentFromTemplate_Call struct {
	*mock.Call
}

// CreateEnvironmentFromTemplate is a helper method to define mock.On call
//   - templateName string
//   - options terra.CreateOptions
func (_e *Client_Expecter) CreateEnvironmentFromTemplate(templateName interface{}, options interface{}) *Client_CreateEnvironmentFromTemplate_Call {
	return &Client_CreateEnvironmentFromTemplate_Call{Call: _e.mock.On("CreateEnvironmentFromTemplate", templateName, options)}
}

func (_c *Client_CreateEnvironmentFromTemplate_Call) Run(run func(templateName string, options terra.CreateOptions)) *Client_CreateEnvironmentFromTemplate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(terra.CreateOptions))
	})
	return _c
}

func (_c *Client_CreateEnvironmentFromTemplate_Call) Return(_a0 string, _a1 error) *Client_CreateEnvironmentFromTemplate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Client_CreateEnvironmentFromTemplate_Call) RunAndReturn(run func(string, terra.CreateOptions) (string, error)) *Client_CreateEnvironmentFromTemplate_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteEnvironments provides a mock function with given fields: _a0
func (_m *Client) DeleteEnvironments(_a0 []terra.Environment) ([]string, error) {
	ret := _m.Called(_a0)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func([]terra.Environment) ([]string, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func([]terra.Environment) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func([]terra.Environment) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_DeleteEnvironments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteEnvironments'
type Client_DeleteEnvironments_Call struct {
	*mock.Call
}

// DeleteEnvironments is a helper method to define mock.On call
//   - _a0 []terra.Environment
func (_e *Client_Expecter) DeleteEnvironments(_a0 interface{}) *Client_DeleteEnvironments_Call {
	return &Client_DeleteEnvironments_Call{Call: _e.mock.On("DeleteEnvironments", _a0)}
}

func (_c *Client_DeleteEnvironments_Call) Run(run func(_a0 []terra.Environment)) *Client_DeleteEnvironments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]terra.Environment))
	})
	return _c
}

func (_c *Client_DeleteEnvironments_Call) Return(_a0 []string, _a1 error) *Client_DeleteEnvironments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Client_DeleteEnvironments_Call) RunAndReturn(run func([]terra.Environment) ([]string, error)) *Client_DeleteEnvironments_Call {
	_c.Call.Return(run)
	return _c
}

// DisableRelease provides a mock function with given fields: _a0, _a1
func (_m *Client) DisableRelease(_a0 string, _a1 string) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_DisableRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DisableRelease'
type Client_DisableRelease_Call struct {
	*mock.Call
}

// DisableRelease is a helper method to define mock.On call
//   - _a0 string
//   - _a1 string
func (_e *Client_Expecter) DisableRelease(_a0 interface{}, _a1 interface{}) *Client_DisableRelease_Call {
	return &Client_DisableRelease_Call{Call: _e.mock.On("DisableRelease", _a0, _a1)}
}

func (_c *Client_DisableRelease_Call) Run(run func(_a0 string, _a1 string)) *Client_DisableRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Client_DisableRelease_Call) Return(_a0 error) *Client_DisableRelease_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_DisableRelease_Call) RunAndReturn(run func(string, string) error) *Client_DisableRelease_Call {
	_c.Call.Return(run)
	return _c
}

// EnableRelease provides a mock function with given fields: _a0, _a1
func (_m *Client) EnableRelease(_a0 terra.Environment, _a1 string) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(terra.Environment, string) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_EnableRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EnableRelease'
type Client_EnableRelease_Call struct {
	*mock.Call
}

// EnableRelease is a helper method to define mock.On call
//   - _a0 terra.Environment
//   - _a1 string
func (_e *Client_Expecter) EnableRelease(_a0 interface{}, _a1 interface{}) *Client_EnableRelease_Call {
	return &Client_EnableRelease_Call{Call: _e.mock.On("EnableRelease", _a0, _a1)}
}

func (_c *Client_EnableRelease_Call) Run(run func(_a0 terra.Environment, _a1 string)) *Client_EnableRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(terra.Environment), args[1].(string))
	})
	return _c
}

func (_c *Client_EnableRelease_Call) Return(_a0 error) *Client_EnableRelease_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_EnableRelease_Call) RunAndReturn(run func(terra.Environment, string) error) *Client_EnableRelease_Call {
	_c.Call.Return(run)
	return _c
}

// Environments provides a mock function with given fields:
func (_m *Client) Environments() (sherlock.Environments, error) {
	ret := _m.Called()

	var r0 sherlock.Environments
	var r1 error
	if rf, ok := ret.Get(0).(func() (sherlock.Environments, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() sherlock.Environments); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sherlock.Environments)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_Environments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Environments'
type Client_Environments_Call struct {
	*mock.Call
}

// Environments is a helper method to define mock.On call
func (_e *Client_Expecter) Environments() *Client_Environments_Call {
	return &Client_Environments_Call{Call: _e.mock.On("Environments")}
}

func (_c *Client_Environments_Call) Run(run func()) *Client_Environments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Client_Environments_Call) Return(_a0 sherlock.Environments, _a1 error) *Client_Environments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Client_Environments_Call) RunAndReturn(run func() (sherlock.Environments, error)) *Client_Environments_Call {
	_c.Call.Return(run)
	return _c
}

// GetStatus provides a mock function with given fields:
func (_m *Client) GetStatus() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_GetStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetStatus'
type Client_GetStatus_Call struct {
	*mock.Call
}

// GetStatus is a helper method to define mock.On call
func (_e *Client_Expecter) GetStatus() *Client_GetStatus_Call {
	return &Client_GetStatus_Call{Call: _e.mock.On("GetStatus")}
}

func (_c *Client_GetStatus_Call) Run(run func()) *Client_GetStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Client_GetStatus_Call) Return(_a0 error) *Client_GetStatus_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_GetStatus_Call) RunAndReturn(run func() error) *Client_GetStatus_Call {
	_c.Call.Return(run)
	return _c
}

// PinEnvironmentVersions provides a mock function with given fields: environmentName, versions
func (_m *Client) PinEnvironmentVersions(environmentName string, versions map[string]terra.VersionOverride) error {
	ret := _m.Called(environmentName, versions)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, map[string]terra.VersionOverride) error); ok {
		r0 = rf(environmentName, versions)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_PinEnvironmentVersions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PinEnvironmentVersions'
type Client_PinEnvironmentVersions_Call struct {
	*mock.Call
}

// PinEnvironmentVersions is a helper method to define mock.On call
//   - environmentName string
//   - versions map[string]terra.VersionOverride
func (_e *Client_Expecter) PinEnvironmentVersions(environmentName interface{}, versions interface{}) *Client_PinEnvironmentVersions_Call {
	return &Client_PinEnvironmentVersions_Call{Call: _e.mock.On("PinEnvironmentVersions", environmentName, versions)}
}

func (_c *Client_PinEnvironmentVersions_Call) Run(run func(environmentName string, versions map[string]terra.VersionOverride)) *Client_PinEnvironmentVersions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(map[string]terra.VersionOverride))
	})
	return _c
}

func (_c *Client_PinEnvironmentVersions_Call) Return(_a0 error) *Client_PinEnvironmentVersions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_PinEnvironmentVersions_Call) RunAndReturn(run func(string, map[string]terra.VersionOverride) error) *Client_PinEnvironmentVersions_Call {
	_c.Call.Return(run)
	return _c
}

// Releases provides a mock function with given fields:
func (_m *Client) Releases() (sherlock.Releases, error) {
	ret := _m.Called()

	var r0 sherlock.Releases
	var r1 error
	if rf, ok := ret.Get(0).(func() (sherlock.Releases, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() sherlock.Releases); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sherlock.Releases)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_Releases_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Releases'
type Client_Releases_Call struct {
	*mock.Call
}

// Releases is a helper method to define mock.On call
func (_e *Client_Expecter) Releases() *Client_Releases_Call {
	return &Client_Releases_Call{Call: _e.mock.On("Releases")}
}

func (_c *Client_Releases_Call) Run(run func()) *Client_Releases_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Client_Releases_Call) Return(_a0 sherlock.Releases, _a1 error) *Client_Releases_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Client_Releases_Call) RunAndReturn(run func() (sherlock.Releases, error)) *Client_Releases_Call {
	_c.Call.Return(run)
	return _c
}

// ReportNewChartVersion provides a mock function with given fields: chartName, newVersion, lastVersion, description
func (_m *Client) ReportNewChartVersion(chartName string, newVersion string, lastVersion string, description string) error {
	ret := _m.Called(chartName, newVersion, lastVersion, description)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string) error); ok {
		r0 = rf(chartName, newVersion, lastVersion, description)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_ReportNewChartVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReportNewChartVersion'
type Client_ReportNewChartVersion_Call struct {
	*mock.Call
}

// ReportNewChartVersion is a helper method to define mock.On call
//   - chartName string
//   - newVersion string
//   - lastVersion string
//   - description string
func (_e *Client_Expecter) ReportNewChartVersion(chartName interface{}, newVersion interface{}, lastVersion interface{}, description interface{}) *Client_ReportNewChartVersion_Call {
	return &Client_ReportNewChartVersion_Call{Call: _e.mock.On("ReportNewChartVersion", chartName, newVersion, lastVersion, description)}
}

func (_c *Client_ReportNewChartVersion_Call) Run(run func(chartName string, newVersion string, lastVersion string, description string)) *Client_ReportNewChartVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *Client_ReportNewChartVersion_Call) Return(_a0 error) *Client_ReportNewChartVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_ReportNewChartVersion_Call) RunAndReturn(run func(string, string, string, string) error) *Client_ReportNewChartVersion_Call {
	_c.Call.Return(run)
	return _c
}

// ResetEnvironmentAndPinToDev provides a mock function with given fields: environment
func (_m *Client) ResetEnvironmentAndPinToDev(environment terra.Environment) error {
	ret := _m.Called(environment)

	var r0 error
	if rf, ok := ret.Get(0).(func(terra.Environment) error); ok {
		r0 = rf(environment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_ResetEnvironmentAndPinToDev_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResetEnvironmentAndPinToDev'
type Client_ResetEnvironmentAndPinToDev_Call struct {
	*mock.Call
}

// ResetEnvironmentAndPinToDev is a helper method to define mock.On call
//   - environment terra.Environment
func (_e *Client_Expecter) ResetEnvironmentAndPinToDev(environment interface{}) *Client_ResetEnvironmentAndPinToDev_Call {
	return &Client_ResetEnvironmentAndPinToDev_Call{Call: _e.mock.On("ResetEnvironmentAndPinToDev", environment)}
}

func (_c *Client_ResetEnvironmentAndPinToDev_Call) Run(run func(environment terra.Environment)) *Client_ResetEnvironmentAndPinToDev_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(terra.Environment))
	})
	return _c
}

func (_c *Client_ResetEnvironmentAndPinToDev_Call) Return(_a0 error) *Client_ResetEnvironmentAndPinToDev_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_ResetEnvironmentAndPinToDev_Call) RunAndReturn(run func(terra.Environment) error) *Client_ResetEnvironmentAndPinToDev_Call {
	_c.Call.Return(run)
	return _c
}

// SetEnvironmentOffline provides a mock function with given fields: environmentName, offline
func (_m *Client) SetEnvironmentOffline(environmentName string, offline bool) error {
	ret := _m.Called(environmentName, offline)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, bool) error); ok {
		r0 = rf(environmentName, offline)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_SetEnvironmentOffline_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetEnvironmentOffline'
type Client_SetEnvironmentOffline_Call struct {
	*mock.Call
}

// SetEnvironmentOffline is a helper method to define mock.On call
//   - environmentName string
//   - offline bool
func (_e *Client_Expecter) SetEnvironmentOffline(environmentName interface{}, offline interface{}) *Client_SetEnvironmentOffline_Call {
	return &Client_SetEnvironmentOffline_Call{Call: _e.mock.On("SetEnvironmentOffline", environmentName, offline)}
}

func (_c *Client_SetEnvironmentOffline_Call) Run(run func(environmentName string, offline bool)) *Client_SetEnvironmentOffline_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(bool))
	})
	return _c
}

func (_c *Client_SetEnvironmentOffline_Call) Return(_a0 error) *Client_SetEnvironmentOffline_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_SetEnvironmentOffline_Call) RunAndReturn(run func(string, bool) error) *Client_SetEnvironmentOffline_Call {
	_c.Call.Return(run)
	return _c
}

// SetTerraHelmfileRefForEntireEnvironment provides a mock function with given fields: environment, terraHelmfileRef
func (_m *Client) SetTerraHelmfileRefForEntireEnvironment(environment terra.Environment, terraHelmfileRef string) error {
	ret := _m.Called(environment, terraHelmfileRef)

	var r0 error
	if rf, ok := ret.Get(0).(func(terra.Environment, string) error); ok {
		r0 = rf(environment, terraHelmfileRef)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_SetTerraHelmfileRefForEntireEnvironment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetTerraHelmfileRefForEntireEnvironment'
type Client_SetTerraHelmfileRefForEntireEnvironment_Call struct {
	*mock.Call
}

// SetTerraHelmfileRefForEntireEnvironment is a helper method to define mock.On call
//   - environment terra.Environment
//   - terraHelmfileRef string
func (_e *Client_Expecter) SetTerraHelmfileRefForEntireEnvironment(environment interface{}, terraHelmfileRef interface{}) *Client_SetTerraHelmfileRefForEntireEnvironment_Call {
	return &Client_SetTerraHelmfileRefForEntireEnvironment_Call{Call: _e.mock.On("SetTerraHelmfileRefForEntireEnvironment", environment, terraHelmfileRef)}
}

func (_c *Client_SetTerraHelmfileRefForEntireEnvironment_Call) Run(run func(environment terra.Environment, terraHelmfileRef string)) *Client_SetTerraHelmfileRefForEntireEnvironment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(terra.Environment), args[1].(string))
	})
	return _c
}

func (_c *Client_SetTerraHelmfileRefForEntireEnvironment_Call) Return(_a0 error) *Client_SetTerraHelmfileRefForEntireEnvironment_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_SetTerraHelmfileRefForEntireEnvironment_Call) RunAndReturn(run func(terra.Environment, string) error) *Client_SetTerraHelmfileRefForEntireEnvironment_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateChartReleaseStatuses provides a mock function with given fields: chartReleaseStatuses
func (_m *Client) UpdateChartReleaseStatuses(chartReleaseStatuses map[string]string) error {
	ret := _m.Called(chartReleaseStatuses)

	var r0 error
	if rf, ok := ret.Get(0).(func(map[string]string) error); ok {
		r0 = rf(chartReleaseStatuses)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_UpdateChartReleaseStatuses_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateChartReleaseStatuses'
type Client_UpdateChartReleaseStatuses_Call struct {
	*mock.Call
}

// UpdateChartReleaseStatuses is a helper method to define mock.On call
//   - chartReleaseStatuses map[string]string
func (_e *Client_Expecter) UpdateChartReleaseStatuses(chartReleaseStatuses interface{}) *Client_UpdateChartReleaseStatuses_Call {
	return &Client_UpdateChartReleaseStatuses_Call{Call: _e.mock.On("UpdateChartReleaseStatuses", chartReleaseStatuses)}
}

func (_c *Client_UpdateChartReleaseStatuses_Call) Run(run func(chartReleaseStatuses map[string]string)) *Client_UpdateChartReleaseStatuses_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(map[string]string))
	})
	return _c
}

func (_c *Client_UpdateChartReleaseStatuses_Call) Return(_a0 error) *Client_UpdateChartReleaseStatuses_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_UpdateChartReleaseStatuses_Call) RunAndReturn(run func(map[string]string) error) *Client_UpdateChartReleaseStatuses_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateForNewChartVersion provides a mock function with given fields: chartSelector, newVersion, lastVersion, description, chartReleaseSelectors
func (_m *Client) UpdateForNewChartVersion(chartSelector string, newVersion string, lastVersion string, description string, chartReleaseSelectors []string) error {
	ret := _m.Called(chartSelector, newVersion, lastVersion, description, chartReleaseSelectors)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, []string) error); ok {
		r0 = rf(chartSelector, newVersion, lastVersion, description, chartReleaseSelectors)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_UpdateForNewChartVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateForNewChartVersion'
type Client_UpdateForNewChartVersion_Call struct {
	*mock.Call
}

// UpdateForNewChartVersion is a helper method to define mock.On call
//   - chartSelector string
//   - newVersion string
//   - lastVersion string
//   - description string
//   - chartReleaseSelectors []string
func (_e *Client_Expecter) UpdateForNewChartVersion(chartSelector interface{}, newVersion interface{}, lastVersion interface{}, description interface{}, chartReleaseSelectors interface{}) *Client_UpdateForNewChartVersion_Call {
	return &Client_UpdateForNewChartVersion_Call{Call: _e.mock.On("UpdateForNewChartVersion", chartSelector, newVersion, lastVersion, description, chartReleaseSelectors)}
}

func (_c *Client_UpdateForNewChartVersion_Call) Run(run func(chartSelector string, newVersion string, lastVersion string, description string, chartReleaseSelectors []string)) *Client_UpdateForNewChartVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(string), args[3].(string), args[4].([]string))
	})
	return _c
}

func (_c *Client_UpdateForNewChartVersion_Call) Return(_a0 error) *Client_UpdateForNewChartVersion_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_UpdateForNewChartVersion_Call) RunAndReturn(run func(string, string, string, string, []string) error) *Client_UpdateForNewChartVersion_Call {
	_c.Call.Return(run)
	return _c
}

// WriteClusters provides a mock function with given fields: _a0
func (_m *Client) WriteClusters(_a0 []terra.Cluster) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func([]terra.Cluster) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_WriteClusters_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteClusters'
type Client_WriteClusters_Call struct {
	*mock.Call
}

// WriteClusters is a helper method to define mock.On call
//   - _a0 []terra.Cluster
func (_e *Client_Expecter) WriteClusters(_a0 interface{}) *Client_WriteClusters_Call {
	return &Client_WriteClusters_Call{Call: _e.mock.On("WriteClusters", _a0)}
}

func (_c *Client_WriteClusters_Call) Run(run func(_a0 []terra.Cluster)) *Client_WriteClusters_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]terra.Cluster))
	})
	return _c
}

func (_c *Client_WriteClusters_Call) Return(_a0 error) *Client_WriteClusters_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_WriteClusters_Call) RunAndReturn(run func([]terra.Cluster) error) *Client_WriteClusters_Call {
	_c.Call.Return(run)
	return _c
}

// WriteEnvironments provides a mock function with given fields: _a0
func (_m *Client) WriteEnvironments(_a0 []terra.Environment) ([]string, error) {
	ret := _m.Called(_a0)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func([]terra.Environment) ([]string, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func([]terra.Environment) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func([]terra.Environment) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_WriteEnvironments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteEnvironments'
type Client_WriteEnvironments_Call struct {
	*mock.Call
}

// WriteEnvironments is a helper method to define mock.On call
//   - _a0 []terra.Environment
func (_e *Client_Expecter) WriteEnvironments(_a0 interface{}) *Client_WriteEnvironments_Call {
	return &Client_WriteEnvironments_Call{Call: _e.mock.On("WriteEnvironments", _a0)}
}

func (_c *Client_WriteEnvironments_Call) Run(run func(_a0 []terra.Environment)) *Client_WriteEnvironments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]terra.Environment))
	})
	return _c
}

func (_c *Client_WriteEnvironments_Call) Return(_a0 []string, _a1 error) *Client_WriteEnvironments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Client_WriteEnvironments_Call) RunAndReturn(run func([]terra.Environment) ([]string, error)) *Client_WriteEnvironments_Call {
	_c.Call.Return(run)
	return _c
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
