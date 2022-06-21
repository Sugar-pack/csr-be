// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// TokenManager is an autogenerated mock type for the TokenManager type
type TokenManager struct {
	mock.Mock
}

// GenerateAccessToken provides a mock function with given fields: ctx, login, password
func (_m *TokenManager) GenerateAccessToken(ctx context.Context, login string, password string) (string, bool, error) {
	ret := _m.Called(ctx, login, password)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = rf(ctx, login, password)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(context.Context, string, string) bool); ok {
		r1 = rf(ctx, login, password)
	} else {
		r1 = ret.Get(1).(bool)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, string, string) error); ok {
		r2 = rf(ctx, login, password)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// RefreshToken provides a mock function with given fields: ctx, token
func (_m *TokenManager) RefreshToken(ctx context.Context, token string) (string, bool, error) {
	ret := _m.Called(ctx, token)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, token)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(context.Context, string) bool); ok {
		r1 = rf(ctx, token)
	} else {
		r1 = ret.Get(1).(bool)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, string) error); ok {
		r2 = rf(ctx, token)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

type mockConstructorTestingTNewTokenManager interface {
	mock.TestingT
	Cleanup(func())
}

// NewTokenManager creates a new instance of TokenManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTokenManager(t mockConstructorTestingTNewTokenManager) *TokenManager {
	mock := &TokenManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
