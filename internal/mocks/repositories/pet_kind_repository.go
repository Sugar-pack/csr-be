// Code generated by mockery v2.13.1. DO NOT EDIT.

package repositories

import (
	context "context"

	ent "git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	mock "github.com/stretchr/testify/mock"

	models "git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

// PetKindRepository is an autogenerated mock type for the PetKindRepository type
type PetKindRepository struct {
	mock.Mock
}

// AllPetKinds provides a mock function with given fields: ctx
func (_m *PetKindRepository) AllPetKinds(ctx context.Context) ([]*ent.PetKind, error) {
	ret := _m.Called(ctx)

	var r0 []*ent.PetKind
	if rf, ok := ret.Get(0).(func(context.Context) []*ent.PetKind); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*ent.PetKind)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreatePetKind provides a mock function with given fields: ctx, ps
func (_m *PetKindRepository) CreatePetKind(ctx context.Context, ps models.PetKind) (*ent.PetKind, error) {
	ret := _m.Called(ctx, ps)

	var r0 *ent.PetKind
	if rf, ok := ret.Get(0).(func(context.Context, models.PetKind) *ent.PetKind); ok {
		r0 = rf(ctx, ps)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ent.PetKind)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, models.PetKind) error); ok {
		r1 = rf(ctx, ps)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeletePetKindByID provides a mock function with given fields: ctx, id
func (_m *PetKindRepository) DeletePetKindByID(ctx context.Context, id int) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PetKindByID provides a mock function with given fields: ctx, id
func (_m *PetKindRepository) PetKindByID(ctx context.Context, id int) (*ent.PetKind, error) {
	ret := _m.Called(ctx, id)

	var r0 *ent.PetKind
	if rf, ok := ret.Get(0).(func(context.Context, int) *ent.PetKind); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ent.PetKind)
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

// UpdatePetKindByID provides a mock function with given fields: ctx, id, newPetKind
func (_m *PetKindRepository) UpdatePetKindByID(ctx context.Context, id int, newPetKind *models.PetKind) (*ent.PetKind, error) {
	ret := _m.Called(ctx, id, newPetKind)

	var r0 *ent.PetKind
	if rf, ok := ret.Get(0).(func(context.Context, int, *models.PetKind) *ent.PetKind); ok {
		r0 = rf(ctx, id, newPetKind)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ent.PetKind)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int, *models.PetKind) error); ok {
		r1 = rf(ctx, id, newPetKind)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewPetKindRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewPetKindRepository creates a new instance of PetKindRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewPetKindRepository(t mockConstructorTestingTNewPetKindRepository) *PetKindRepository {
	mock := &PetKindRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
