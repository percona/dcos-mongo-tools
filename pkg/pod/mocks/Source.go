// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import mock "github.com/stretchr/testify/mock"
import pod "github.com/percona/mongodb-orchestration-tools/pkg/pod"

// Source is an autogenerated mock type for the Source type
type Source struct {
	mock.Mock
}

// GetPodTasks provides a mock function with given fields: podName
func (_m *Source) GetPodTasks(podName string) ([]pod.Task, error) {
	ret := _m.Called(podName)

	var r0 []pod.Task
	if rf, ok := ret.Get(0).(func(string) []pod.Task); ok {
		r0 = rf(podName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]pod.Task)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(podName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPodURL provides a mock function with given fields:
func (_m *Source) GetPodURL() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetPods provides a mock function with given fields:
func (_m *Source) GetPods() (*pod.Pods, error) {
	ret := _m.Called()

	var r0 *pod.Pods
	if rf, ok := ret.Get(0).(func() *pod.Pods); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pod.Pods)
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

// Name provides a mock function with given fields:
func (_m *Source) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
