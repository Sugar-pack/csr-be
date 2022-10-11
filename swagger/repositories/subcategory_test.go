package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type subcategoryRepositorySuite struct {
	suite.Suite
	ctx           context.Context
	client        *ent.Client
	repository    SubcategoryRepository
	subcategories []*ent.Subcategory
	category      *ent.Category
}

func TestSubcategorySuite(t *testing.T) {
	suite.Run(t, new(subcategoryRepositorySuite))
}

func (s *subcategoryRepositorySuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:subcategory?mode=memory&cache=shared&_fk=1")
	s.client = client
	s.repository = NewSubcategoryRepository()

	categoryName := "category"
	_, err := s.client.Category.Delete().Exec(s.ctx) // clean up
	if err != nil {
		t.Fatal(err)
	}
	category, err := s.client.Category.Create().SetName(categoryName).Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	s.category = category

	s.subcategories = []*ent.Subcategory{
		{
			Name:                "subcategory 1",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
		{
			Name:                "subcategory 2",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
		{
			Name:                "subcategory 3",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
		{
			Name:                "subcategory 4",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
	}
	_, err = s.client.Subcategory.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range s.subcategories {
		subcat, err := s.client.Subcategory.Create().
			SetName(value.Name).SetMaxReservationTime(value.MaxReservationTime).
			SetMaxReservationUnits(value.MaxReservationUnits).
			SetCategory(s.category).
			Save(s.ctx)
		if err != nil {
			t.Fatal(err)
		}
		s.subcategories[i] = subcat
		s.subcategories[i].Edges.Category = category
	}
}

func (s *subcategoryRepositorySuite) TearDownSuite() {
	s.client.Close()
}

func (s *subcategoryRepositorySuite) TestSubcategoryRepository_CreateSubcategory_CategoryNotExists() {
	t := s.T()
	name := "test subcategory"
	maxReservationTime := int64(10)
	maxReservationUnits := int64(1)
	categoryID := int64(s.category.ID + 10)
	newSubcategory := models.NewSubcategory{
		Category:            &categoryID,
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnits,
		Name:                &name,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	createdSubcategory, err := s.repository.CreateSubcategory(ctx, int(categoryID), newSubcategory)
	assert.Error(t, err)
	assert.Nil(t, createdSubcategory)
	assert.NoError(t, tx.Rollback())
}

func (s *subcategoryRepositorySuite) TestSubcategoryRepository_CreateSubcategory_OK() {
	t := s.T()
	name := "test subcategory"
	maxReservationTime := int64(10)
	maxReservationUnits := int64(1)
	categoryID := int64(s.category.ID)
	newSubcategory := models.NewSubcategory{
		Category:            &categoryID,
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnits,
		Name:                &name,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	createdSubcategory, err := s.repository.CreateSubcategory(ctx, int(categoryID), newSubcategory)
	assert.NoError(t, err)
	assert.Equal(t, name, createdSubcategory.Name)
	assert.Equal(t, maxReservationTime, createdSubcategory.MaxReservationTime)
	assert.Equal(t, maxReservationUnits, createdSubcategory.MaxReservationUnits)
	assert.NoError(t, tx.Rollback())
}

func (s *subcategoryRepositorySuite) TestSubcategoryRepository_ListSubcategories_CategoryNotExists() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.ListSubcategories(ctx, s.category.ID+10)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, 0, len(categories))
}

func (s *subcategoryRepositorySuite) TestSubcategoryRepository_ListSubcategories_OK() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.ListSubcategories(ctx, s.category.ID)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.subcategories), len(categories))
	for _, value := range categories {
		assert.True(t, containsSubcategory(t, value, s.subcategories))
	}
}

func (s *subcategoryRepositorySuite) TestSubcategoryRepository_SubcategoryByID() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	subcat, err := s.repository.SubcategoryByID(ctx, s.subcategories[0].ID)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, s.subcategories[0].Name, subcat.Name)
	assert.Equal(t, s.subcategories[0].MaxReservationTime, subcat.MaxReservationTime)
	assert.Equal(t, s.subcategories[0].MaxReservationUnits, subcat.MaxReservationUnits)
	assert.Equal(t, s.subcategories[0].Edges.Category.ID, subcat.Edges.Category.ID)
}

func (s *subcategoryRepositorySuite) TestSubcategoryRepository_DeleteSubcategoryByID() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = s.repository.DeleteSubcategoryByID(ctx, s.subcategories[0].ID)
	assert.NoError(t, err)
	assert.NoError(t, tx.Rollback())
}

func (s *subcategoryRepositorySuite) TestSubcategoryRepository_UpdateSubcategory() {
	t := s.T()
	name := "new subcategory name"
	update := models.NewSubcategory{
		Name: &name,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	subcat, err := s.repository.UpdateSubcategory(ctx, s.subcategories[0].ID, update)
	assert.NoError(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Equal(t, name, subcat.Name)
	assert.Equal(t, s.subcategories[0].MaxReservationTime, subcat.MaxReservationTime)
	assert.Equal(t, s.subcategories[0].MaxReservationUnits, subcat.MaxReservationUnits)
}

func containsSubcategory(t *testing.T, eq *ent.Subcategory, list []*ent.Subcategory) bool {
	t.Helper()
	for _, v := range list {
		if eq.Name == v.Name && eq.ID == v.ID && eq.MaxReservationUnits == v.MaxReservationUnits &&
			eq.MaxReservationTime == v.MaxReservationTime {
			return true
		}
	}
	return false
}
