// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// PasswordGenerator is an autogenerated mock type for the PasswordGenerator type
type PasswordGenerator struct {
	mock.Mock
}

// NewPassword provides a mock function with given fields:
func (_m *PasswordGenerator) NewPassword() (string, error) {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewPasswordGenerator interface {
	mock.TestingT
	Cleanup(func())
}

// NewPasswordGenerator creates a new instance of PasswordGenerator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewPasswordGenerator(t mockConstructorTestingTNewPasswordGenerator) *PasswordGenerator {
	mock := &PasswordGenerator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
