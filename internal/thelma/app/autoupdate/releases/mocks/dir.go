// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Dir is an autogenerated mock type for the Dir type
type Dir struct {
	mock.Mock
}

type Dir_Expecter struct {
	mock *mock.Mock
}

func (_m *Dir) EXPECT() *Dir_Expecter {
	return &Dir_Expecter{mock: &_m.Mock}
}

// CleanupOldReleases provides a mock function with given fields: keepReleases
func (_m *Dir) CleanupOldReleases(keepReleases int) error {
	ret := _m.Called(keepReleases)

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(keepReleases)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Dir_CleanupOldReleases_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CleanupOldReleases'
type Dir_CleanupOldReleases_Call struct {
	*mock.Call
}

// CleanupOldReleases is a helper method to define mock.On call
//   - keepReleases int
func (_e *Dir_Expecter) CleanupOldReleases(keepReleases interface{}) *Dir_CleanupOldReleases_Call {
	return &Dir_CleanupOldReleases_Call{Call: _e.mock.On("CleanupOldReleases", keepReleases)}
}

func (_c *Dir_CleanupOldReleases_Call) Run(run func(keepReleases int)) *Dir_CleanupOldReleases_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *Dir_CleanupOldReleases_Call) Return(_a0 error) *Dir_CleanupOldReleases_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Dir_CleanupOldReleases_Call) RunAndReturn(run func(int) error) *Dir_CleanupOldReleases_Call {
	_c.Call.Return(run)
	return _c
}

// CopyUnpackedArchive provides a mock function with given fields: unpackDir
func (_m *Dir) CopyUnpackedArchive(unpackDir string) error {
	ret := _m.Called(unpackDir)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(unpackDir)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Dir_CopyUnpackedArchive_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CopyUnpackedArchive'
type Dir_CopyUnpackedArchive_Call struct {
	*mock.Call
}

// CopyUnpackedArchive is a helper method to define mock.On call
//   - unpackDir string
func (_e *Dir_Expecter) CopyUnpackedArchive(unpackDir interface{}) *Dir_CopyUnpackedArchive_Call {
	return &Dir_CopyUnpackedArchive_Call{Call: _e.mock.On("CopyUnpackedArchive", unpackDir)}
}

func (_c *Dir_CopyUnpackedArchive_Call) Run(run func(unpackDir string)) *Dir_CopyUnpackedArchive_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Dir_CopyUnpackedArchive_Call) Return(_a0 error) *Dir_CopyUnpackedArchive_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Dir_CopyUnpackedArchive_Call) RunAndReturn(run func(string) error) *Dir_CopyUnpackedArchive_Call {
	_c.Call.Return(run)
	return _c
}

// CurrentVersion provides a mock function with given fields:
func (_m *Dir) CurrentVersion() (string, error) {
	ret := _m.Called()

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func() (string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Dir_CurrentVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CurrentVersion'
type Dir_CurrentVersion_Call struct {
	*mock.Call
}

// CurrentVersion is a helper method to define mock.On call
func (_e *Dir_Expecter) CurrentVersion() *Dir_CurrentVersion_Call {
	return &Dir_CurrentVersion_Call{Call: _e.mock.On("CurrentVersion")}
}

func (_c *Dir_CurrentVersion_Call) Run(run func()) *Dir_CurrentVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Dir_CurrentVersion_Call) Return(_a0 string, _a1 error) *Dir_CurrentVersion_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Dir_CurrentVersion_Call) RunAndReturn(run func() (string, error)) *Dir_CurrentVersion_Call {
	_c.Call.Return(run)
	return _c
}

// CurrentVersionMatches provides a mock function with given fields: version
func (_m *Dir) CurrentVersionMatches(version string) bool {
	ret := _m.Called(version)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(version)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Dir_CurrentVersionMatches_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CurrentVersionMatches'
type Dir_CurrentVersionMatches_Call struct {
	*mock.Call
}

// CurrentVersionMatches is a helper method to define mock.On call
//   - version string
func (_e *Dir_Expecter) CurrentVersionMatches(version interface{}) *Dir_CurrentVersionMatches_Call {
	return &Dir_CurrentVersionMatches_Call{Call: _e.mock.On("CurrentVersionMatches", version)}
}

func (_c *Dir_CurrentVersionMatches_Call) Run(run func(version string)) *Dir_CurrentVersionMatches_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Dir_CurrentVersionMatches_Call) Return(_a0 bool) *Dir_CurrentVersionMatches_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Dir_CurrentVersionMatches_Call) RunAndReturn(run func(string) bool) *Dir_CurrentVersionMatches_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateCurrentReleaseSymlink provides a mock function with given fields: version
func (_m *Dir) UpdateCurrentReleaseSymlink(version string) error {
	ret := _m.Called(version)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(version)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Dir_UpdateCurrentReleaseSymlink_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateCurrentReleaseSymlink'
type Dir_UpdateCurrentReleaseSymlink_Call struct {
	*mock.Call
}

// UpdateCurrentReleaseSymlink is a helper method to define mock.On call
//   - version string
func (_e *Dir_Expecter) UpdateCurrentReleaseSymlink(version interface{}) *Dir_UpdateCurrentReleaseSymlink_Call {
	return &Dir_UpdateCurrentReleaseSymlink_Call{Call: _e.mock.On("UpdateCurrentReleaseSymlink", version)}
}

func (_c *Dir_UpdateCurrentReleaseSymlink_Call) Run(run func(version string)) *Dir_UpdateCurrentReleaseSymlink_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Dir_UpdateCurrentReleaseSymlink_Call) Return(_a0 error) *Dir_UpdateCurrentReleaseSymlink_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Dir_UpdateCurrentReleaseSymlink_Call) RunAndReturn(run func(string) error) *Dir_UpdateCurrentReleaseSymlink_Call {
	_c.Call.Return(run)
	return _c
}

// WithInstallerLock provides a mock function with given fields: fn
func (_m *Dir) WithInstallerLock(fn func() error) error {
	ret := _m.Called(fn)

	var r0 error
	if rf, ok := ret.Get(0).(func(func() error) error); ok {
		r0 = rf(fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Dir_WithInstallerLock_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithInstallerLock'
type Dir_WithInstallerLock_Call struct {
	*mock.Call
}

// WithInstallerLock is a helper method to define mock.On call
//   - fn func() error
func (_e *Dir_Expecter) WithInstallerLock(fn interface{}) *Dir_WithInstallerLock_Call {
	return &Dir_WithInstallerLock_Call{Call: _e.mock.On("WithInstallerLock", fn)}
}

func (_c *Dir_WithInstallerLock_Call) Run(run func(fn func() error)) *Dir_WithInstallerLock_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(func() error))
	})
	return _c
}

func (_c *Dir_WithInstallerLock_Call) Return(_a0 error) *Dir_WithInstallerLock_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Dir_WithInstallerLock_Call) RunAndReturn(run func(func() error) error) *Dir_WithInstallerLock_Call {
	_c.Call.Return(run)
	return _c
}

// NewDir creates a new instance of Dir. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDir(t interface {
	mock.TestingT
	Cleanup(func())
}) *Dir {
	mock := &Dir{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
