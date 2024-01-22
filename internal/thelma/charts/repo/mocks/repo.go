// Code generated by mockery v2.40.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Repo is an autogenerated mock type for the Repo type
type Repo struct {
	mock.Mock
}

type Repo_Expecter struct {
	mock *mock.Mock
}

func (_m *Repo) EXPECT() *Repo_Expecter {
	return &Repo_Expecter{mock: &_m.Mock}
}

// DownloadIndex provides a mock function with given fields: destPath
func (_m *Repo) DownloadIndex(destPath string) error {
	ret := _m.Called(destPath)

	if len(ret) == 0 {
		panic("no return value specified for DownloadIndex")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(destPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repo_DownloadIndex_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DownloadIndex'
type Repo_DownloadIndex_Call struct {
	*mock.Call
}

// DownloadIndex is a helper method to define mock.On call
//   - destPath string
func (_e *Repo_Expecter) DownloadIndex(destPath interface{}) *Repo_DownloadIndex_Call {
	return &Repo_DownloadIndex_Call{Call: _e.mock.On("DownloadIndex", destPath)}
}

func (_c *Repo_DownloadIndex_Call) Run(run func(destPath string)) *Repo_DownloadIndex_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Repo_DownloadIndex_Call) Return(_a0 error) *Repo_DownloadIndex_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repo_DownloadIndex_Call) RunAndReturn(run func(string) error) *Repo_DownloadIndex_Call {
	_c.Call.Return(run)
	return _c
}

// HasIndex provides a mock function with given fields:
func (_m *Repo) HasIndex() (bool, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for HasIndex")
	}

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

// Repo_HasIndex_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasIndex'
type Repo_HasIndex_Call struct {
	*mock.Call
}

// HasIndex is a helper method to define mock.On call
func (_e *Repo_Expecter) HasIndex() *Repo_HasIndex_Call {
	return &Repo_HasIndex_Call{Call: _e.mock.On("HasIndex")}
}

func (_c *Repo_HasIndex_Call) Run(run func()) *Repo_HasIndex_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Repo_HasIndex_Call) Return(_a0 bool, _a1 error) *Repo_HasIndex_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Repo_HasIndex_Call) RunAndReturn(run func() (bool, error)) *Repo_HasIndex_Call {
	_c.Call.Return(run)
	return _c
}

// IsLocked provides a mock function with given fields:
func (_m *Repo) IsLocked() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsLocked")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Repo_IsLocked_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsLocked'
type Repo_IsLocked_Call struct {
	*mock.Call
}

// IsLocked is a helper method to define mock.On call
func (_e *Repo_Expecter) IsLocked() *Repo_IsLocked_Call {
	return &Repo_IsLocked_Call{Call: _e.mock.On("IsLocked")}
}

func (_c *Repo_IsLocked_Call) Run(run func()) *Repo_IsLocked_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Repo_IsLocked_Call) Return(_a0 bool) *Repo_IsLocked_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repo_IsLocked_Call) RunAndReturn(run func() bool) *Repo_IsLocked_Call {
	_c.Call.Return(run)
	return _c
}

// Lock provides a mock function with given fields:
func (_m *Repo) Lock() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Lock")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repo_Lock_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Lock'
type Repo_Lock_Call struct {
	*mock.Call
}

// Lock is a helper method to define mock.On call
func (_e *Repo_Expecter) Lock() *Repo_Lock_Call {
	return &Repo_Lock_Call{Call: _e.mock.On("Lock")}
}

func (_c *Repo_Lock_Call) Run(run func()) *Repo_Lock_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Repo_Lock_Call) Return(_a0 error) *Repo_Lock_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repo_Lock_Call) RunAndReturn(run func() error) *Repo_Lock_Call {
	_c.Call.Return(run)
	return _c
}

// RepoURL provides a mock function with given fields:
func (_m *Repo) RepoURL() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RepoURL")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Repo_RepoURL_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RepoURL'
type Repo_RepoURL_Call struct {
	*mock.Call
}

// RepoURL is a helper method to define mock.On call
func (_e *Repo_Expecter) RepoURL() *Repo_RepoURL_Call {
	return &Repo_RepoURL_Call{Call: _e.mock.On("RepoURL")}
}

func (_c *Repo_RepoURL_Call) Run(run func()) *Repo_RepoURL_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Repo_RepoURL_Call) Return(_a0 string) *Repo_RepoURL_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repo_RepoURL_Call) RunAndReturn(run func() string) *Repo_RepoURL_Call {
	_c.Call.Return(run)
	return _c
}

// Unlock provides a mock function with given fields:
func (_m *Repo) Unlock() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Unlock")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repo_Unlock_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Unlock'
type Repo_Unlock_Call struct {
	*mock.Call
}

// Unlock is a helper method to define mock.On call
func (_e *Repo_Expecter) Unlock() *Repo_Unlock_Call {
	return &Repo_Unlock_Call{Call: _e.mock.On("Unlock")}
}

func (_c *Repo_Unlock_Call) Run(run func()) *Repo_Unlock_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Repo_Unlock_Call) Return(_a0 error) *Repo_Unlock_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repo_Unlock_Call) RunAndReturn(run func() error) *Repo_Unlock_Call {
	_c.Call.Return(run)
	return _c
}

// UploadChart provides a mock function with given fields: fromPath
func (_m *Repo) UploadChart(fromPath string) error {
	ret := _m.Called(fromPath)

	if len(ret) == 0 {
		panic("no return value specified for UploadChart")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(fromPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repo_UploadChart_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UploadChart'
type Repo_UploadChart_Call struct {
	*mock.Call
}

// UploadChart is a helper method to define mock.On call
//   - fromPath string
func (_e *Repo_Expecter) UploadChart(fromPath interface{}) *Repo_UploadChart_Call {
	return &Repo_UploadChart_Call{Call: _e.mock.On("UploadChart", fromPath)}
}

func (_c *Repo_UploadChart_Call) Run(run func(fromPath string)) *Repo_UploadChart_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Repo_UploadChart_Call) Return(_a0 error) *Repo_UploadChart_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repo_UploadChart_Call) RunAndReturn(run func(string) error) *Repo_UploadChart_Call {
	_c.Call.Return(run)
	return _c
}

// UploadIndex provides a mock function with given fields: fromPath
func (_m *Repo) UploadIndex(fromPath string) error {
	ret := _m.Called(fromPath)

	if len(ret) == 0 {
		panic("no return value specified for UploadIndex")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(fromPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repo_UploadIndex_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UploadIndex'
type Repo_UploadIndex_Call struct {
	*mock.Call
}

// UploadIndex is a helper method to define mock.On call
//   - fromPath string
func (_e *Repo_Expecter) UploadIndex(fromPath interface{}) *Repo_UploadIndex_Call {
	return &Repo_UploadIndex_Call{Call: _e.mock.On("UploadIndex", fromPath)}
}

func (_c *Repo_UploadIndex_Call) Run(run func(fromPath string)) *Repo_UploadIndex_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Repo_UploadIndex_Call) Return(_a0 error) *Repo_UploadIndex_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repo_UploadIndex_Call) RunAndReturn(run func(string) error) *Repo_UploadIndex_Call {
	_c.Call.Return(run)
	return _c
}

// NewRepo creates a new instance of Repo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *Repo {
	mock := &Repo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
