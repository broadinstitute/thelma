// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	terra "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	mock "github.com/stretchr/testify/mock"
)

// AppRelease is an autogenerated mock type for the AppRelease type
type AppRelease struct {
	mock.Mock
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
