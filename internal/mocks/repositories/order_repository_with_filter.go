// Code generated by mockery v2.13.1. DO NOT EDIT.

package repositories

import (
	context "context"

	ent "git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// OrderRepositoryWithFilter is an autogenerated mock type for the OrderRepositoryWithFilter type
type OrderRepositoryWithFilter struct {
	mock.Mock
}

// OrdersByPeriodAndStatus provides a mock function with given fields: ctx, from, to, status, limit, offset, orderBy, orderColumn
func (_m *OrderRepositoryWithFilter) OrdersByPeriodAndStatus(ctx context.Context, from time.Time, to time.Time, status string, limit int, offset int, orderBy string, orderColumn string) ([]*ent.Order, error) {
	ret := _m.Called(ctx, from, to, status, limit, offset, orderBy, orderColumn)

	var r0 []*ent.Order
	if rf, ok := ret.Get(0).(func(context.Context, time.Time, time.Time, string, int, int, string, string) []*ent.Order); ok {
		r0 = rf(ctx, from, to, status, limit, offset, orderBy, orderColumn)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*ent.Order)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, time.Time, time.Time, string, int, int, string, string) error); ok {
		r1 = rf(ctx, from, to, status, limit, offset, orderBy, orderColumn)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OrdersByPeriodAndStatusTotal provides a mock function with given fields: ctx, from, to, status
func (_m *OrderRepositoryWithFilter) OrdersByPeriodAndStatusTotal(ctx context.Context, from time.Time, to time.Time, status string) (int, error) {
	ret := _m.Called(ctx, from, to, status)

	var r0 int
	if rf, ok := ret.Get(0).(func(context.Context, time.Time, time.Time, string) int); ok {
		r0 = rf(ctx, from, to, status)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, time.Time, time.Time, string) error); ok {
		r1 = rf(ctx, from, to, status)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OrdersByStatus provides a mock function with given fields: ctx, status, limit, offset, orderBy, orderColumn
func (_m *OrderRepositoryWithFilter) OrdersByStatus(ctx context.Context, status string, limit int, offset int, orderBy string, orderColumn string) ([]*ent.Order, error) {
	ret := _m.Called(ctx, status, limit, offset, orderBy, orderColumn)

	var r0 []*ent.Order
	if rf, ok := ret.Get(0).(func(context.Context, string, int, int, string, string) []*ent.Order); ok {
		r0 = rf(ctx, status, limit, offset, orderBy, orderColumn)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*ent.Order)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, int, int, string, string) error); ok {
		r1 = rf(ctx, status, limit, offset, orderBy, orderColumn)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OrdersByStatusTotal provides a mock function with given fields: ctx, status
func (_m *OrderRepositoryWithFilter) OrdersByStatusTotal(ctx context.Context, status string) (int, error) {
	ret := _m.Called(ctx, status)

	var r0 int
	if rf, ok := ret.Get(0).(func(context.Context, string) int); ok {
		r0 = rf(ctx, status)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, status)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewOrderRepositoryWithFilter interface {
	mock.TestingT
	Cleanup(func())
}

// NewOrderRepositoryWithFilter creates a new instance of OrderRepositoryWithFilter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewOrderRepositoryWithFilter(t mockConstructorTestingTNewOrderRepositoryWithFilter) *OrderRepositoryWithFilter {
	mock := &OrderRepositoryWithFilter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
