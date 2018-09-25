// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import mock "github.com/stretchr/testify/mock"
import replset "github.com/percona/dcos-mongo-tools/watchdog/replset"
import watcher "github.com/percona/dcos-mongo-tools/watchdog/watcher"

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Manager) Close() {
	_m.Called()
}

// Get provides a mock function with given fields: rsName
func (_m *Manager) Get(rsName string) *watcher.Watcher {
	ret := _m.Called(rsName)

	var r0 *watcher.Watcher
	if rf, ok := ret.Get(0).(func(string) *watcher.Watcher); ok {
		r0 = rf(rsName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*watcher.Watcher)
		}
	}

	return r0
}

// HasWatcher provides a mock function with given fields: rsName
func (_m *Manager) HasWatcher(rsName string) bool {
	ret := _m.Called(rsName)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(rsName)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Stop provides a mock function with given fields: rsName
func (_m *Manager) Stop(rsName string) {
	_m.Called(rsName)
}

// Watch provides a mock function with given fields: rs
func (_m *Manager) Watch(rs *replset.Replset) {
	_m.Called(rs)
}
