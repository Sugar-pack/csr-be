// Code generated by mockery v2.13.1. DO NOT EDIT.

package repositories

import (
	context "context"

	ent "git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	mock "github.com/stretchr/testify/mock"

	models "git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

// SubcategoryRepository is an autogenerated mock type for the SubcategoryRepository type
type SubcategoryRepository struct {
	mock.Mock
}

// CreateSubcategory provides a mock function with given fields: ctx, categoryID, newSubcategory
func (_m *SubcategoryRepository) CreateSubcategory(ctx context.Context, categoryID int, newSubcategory models.NewSubcategory) (*ent.Subcategory, error) {
	ret := _m.Called(ctx, categoryID, newSubcategory)

	var r0 *ent.Subcategory
	if rf, ok := ret.Get(0).(func(context.Context, int, models.NewSubcategory) *ent.Subcategory); ok {
		r0 = rf(ctx, categoryID, newSubcategory)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ent.Subcategory)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int, models.NewSubcategory) error); ok {
		r1 = rf(ctx, categoryID, newSubcategory)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteSubcategoryByID provides a mock function with given fields: ctx, id
func (_m *SubcategoryRepository) DeleteSubcategoryByID(ctx context.Context, id int) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListSubcategories provides a mock function with given fields: ctx, categoryID
func (_m *SubcategoryRepository) ListSubcategories(ctx context.Context, categoryID int) ([]*ent.Subcategory, error) {
	ret := _m.Called(ctx, categoryID)

	var r0 []*ent.Subcategory
	if rf, ok := ret.Get(0).(func(context.Context, int) []*ent.Subcategory); ok {
		r0 = rf(ctx, categoryID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*ent.Subcategory)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(ctx, categoryID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SubcategoryByID provides a mock function with given fields: ctx, id
func (_m *SubcategoryRepository) SubcategoryByID(ctx context.Context, id int) (*ent.Subcategory, error) {
	ret := _m.Called(ctx, id)

	var r0 *ent.Subcategory
	if rf, ok := ret.Get(0).(func(context.Context, int) *ent.Subcategory); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ent.Subcategory)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateSubcategory provides a mock function with given fields: ctx, id, update
func (_m *SubcategoryRepository) UpdateSubcategory(ctx context.Context, id int, update models.NewSubcategory) (*ent.Subcategory, error) {
	ret := _m.Called(ctx, id, update)

	var r0 *ent.Subcategory
	if rf, ok := ret.Get(0).(func(context.Context, int, models.NewSubcategory) *ent.Subcategory); ok {
		r0 = rf(ctx, id, update)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ent.Subcategory)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int, models.NewSubcategory) error); ok {
		r1 = rf(ctx, id, update)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewSubcategoryRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewSubcategoryRepository creates a new instance of SubcategoryRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSubcategoryRepository(t mockConstructorTestingTNewSubcategoryRepository) *SubcategoryRepository {
	mock := &SubcategoryRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
