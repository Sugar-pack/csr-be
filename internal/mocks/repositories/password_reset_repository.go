// Code generated by mockery v2.13.1. DO NOT EDIT.

package repositories

import (
	context "context"

	ent "git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// PasswordResetRepository is an autogenerated mock type for the PasswordResetRepository type
type PasswordResetRepository struct {
	mock.Mock
}

// CreateToken provides a mock function with given fields: ctx, token, ttl, userID
func (_m *PasswordResetRepository) CreateToken(ctx context.Context, token string, ttl time.Time, userID int) error {
	ret := _m.Called(ctx, token, ttl, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time, int) error); ok {
		r0 = rf(ctx, token, ttl, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteToken provides a mock function with given fields: ctx, token
func (_m *PasswordResetRepository) DeleteToken(ctx context.Context, token string) error {
	ret := _m.Called(ctx, token)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, token)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetToken provides a mock function with given fields: ctx, token
func (_m *PasswordResetRepository) GetToken(ctx context.Context, token string) (*ent.PasswordReset, error) {
	ret := _m.Called(ctx, token)

	var r0 *ent.PasswordReset
	if rf, ok := ret.Get(0).(func(context.Context, string) *ent.PasswordReset); ok {
		r0 = rf(ctx, token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ent.PasswordReset)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewPasswordResetRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewPasswordResetRepository creates a new instance of PasswordResetRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewPasswordResetRepository(t mockConstructorTestingTNewPasswordResetRepository) *PasswordResetRepository {
	mock := &PasswordResetRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
