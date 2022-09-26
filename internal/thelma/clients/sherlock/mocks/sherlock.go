// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	sherlock "github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	mock "github.com/stretchr/testify/mock"
)

// StateLoader is an autogenerated mock type for the StateLoader type
type StateLoader struct {
	mock.Mock
}

// ClusterReleases provides a mock function with given fields: clusterName
func (_m *StateLoader) ClusterReleases(clusterName string) (sherlock.Releases, error) {
	ret := _m.Called(clusterName)

	var r0 sherlock.Releases
	if rf, ok := ret.Get(0).(func(string) sherlock.Releases); ok {
		r0 = rf(clusterName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sherlock.Releases)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(clusterName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Clusters provides a mock function with given fields:
func (_m *StateLoader) Clusters() (sherlock.Clusters, error) {
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

// EnvironmentReleases provides a mock function with given fields: environmentName
func (_m *StateLoader) EnvironmentReleases(environmentName string) (sherlock.Releases, error) {
	ret := _m.Called(environmentName)

	var r0 sherlock.Releases
	if rf, ok := ret.Get(0).(func(string) sherlock.Releases); ok {
		r0 = rf(environmentName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sherlock.Releases)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(environmentName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Environments provides a mock function with given fields:
func (_m *StateLoader) Environments() (sherlock.Environments, error) {
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

type mockConstructorTestingTNewStateLoader interface {
	mock.TestingT
	Cleanup(func())
}

// NewStateLoader creates a new instance of StateLoader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewStateLoader(t mockConstructorTestingTNewStateLoader) *StateLoader {
	mock := &StateLoader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}