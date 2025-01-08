package repositories

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/category"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type categoryRepositorySuite struct {
	suite.Suite
	ctx        context.Context
	client     *ent.Client
	repository domain.CategoryRepository
	categories []*ent.Category
}

func TestCategorySuite(t *testing.T) {
	suite.Run(t, new(categoryRepositorySuite))
}

func (s *categoryRepositorySuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:category?mode=memory&cache=shared&_fk=1")
	s.client = client

	_, err := s.client.Equipment.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	eq, err := s.client.Equipment.Create().
		SetName("equipment").
		Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	s.repository = NewCategoryRepository()

	s.categories = []*ent.Category{
		{
			Name:                "category 1",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
		{
			Name:                "category 2",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
		{
			Name:                "category 3",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
		{
			Name:                "category 4",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
		{
			Name:                "category 5",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
	}
	if _, err = s.client.Category.Delete().Exec(s.ctx); err != nil {
		t.Fatal(err)
	}

	for i, value := range s.categories {
		q := s.client.Category.Create().
			SetName(value.Name).SetMaxReservationTime(value.MaxReservationTime).
			SetMaxReservationUnits(value.MaxReservationUnits)
		if i >= len(s.categories)-1 {
			q.AddEquipments(eq)
		}

		c, errCreate := q.Save(s.ctx)
		if errCreate != nil {
			t.Fatal(errCreate)
		}
		s.categories[i].ID = c.ID
	}
}

func (s *categoryRepositorySuite) TearDownSuite() {
	s.client.Close()
}

func (s *categoryRepositorySuite) TestCategoryRepository_CreateCategory() {
	t := s.T()
	name := "category"
	maxReservationTime := int64(10)
	maxReservationUnits := int64(1)
	hasSubcat := true
	newCategory := models.CreateNewCategory{
		Name:                &name,
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnits,
		HasSubcategory:      &hasSubcat,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	createdCategory, err := s.repository.CreateCategory(ctx, newCategory)
	require.NoError(t, err)
	require.Equal(t, name, createdCategory.Name)
	require.Equal(t, maxReservationTime, createdCategory.MaxReservationTime)
	require.Equal(t, maxReservationUnits, createdCategory.MaxReservationUnits)
	require.NoError(t, tx.Rollback())
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategoriesTotal() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	total, err := s.repository.AllCategoriesTotal(ctx)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.categories), total)
}
func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_EmptyOrderBy() {
	t := s.T()

	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			OrderColumn: category.FieldID,
		},
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, filter)
	require.Error(t, err)
	require.NoError(t, tx.Commit())
	require.Nil(t, categories)
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_EmptyOrderColumn() {
	t := s.T()
	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:   math.MaxInt,
			OrderBy: utils.AscOrder,
		},
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, filter)
	require.Error(t, err)
	require.NoError(t, tx.Commit())
	require.Nil(t, categories)
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_WrongOrderColumn() {
	t := s.T()

	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			OrderBy:     utils.AscOrder,
			OrderColumn: category.FieldMaxReservationTime,
		},
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, filter)
	require.Error(t, err)
	require.NoError(t, tx.Commit())
	require.Nil(t, categories)
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_OrderByIDDesc() {
	t := s.T()

	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			OrderBy:     utils.DescOrder,
			OrderColumn: category.FieldID,
		},
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, filter)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.categories), len(categories))
	prevCategoryID := math.MaxInt
	for _, value := range categories {
		require.True(t, containsCategory(t, value, s.categories))
		require.LessOrEqual(t, value.ID, prevCategoryID)
		prevCategoryID = value.ID
	}
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_OrderByNameDesc() {
	t := s.T()
	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			OrderBy:     utils.DescOrder,
			OrderColumn: category.FieldName,
		},
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, filter)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.categories), len(categories))
	prevCategoryName := "zzzzzzzzzzzzzzzzzzzzz"
	for _, value := range categories {
		require.True(t, containsCategory(t, value, s.categories))
		require.LessOrEqual(t, value.Name, prevCategoryName)
		prevCategoryName = value.Name
	}
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_OrderByIDAsc() {
	t := s.T()

	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			OrderBy:     utils.AscOrder,
			OrderColumn: category.FieldID,
		},
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, filter)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.categories), len(categories))
	prevCategoryID := 0
	for _, value := range categories {
		require.True(t, containsCategory(t, value, s.categories))
		require.GreaterOrEqual(t, value.ID, prevCategoryID)
		prevCategoryID = value.ID
	}
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_OrderByNameAsc() {
	t := s.T()

	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			OrderBy:     utils.AscOrder,
			OrderColumn: category.FieldName,
		},
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, filter)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.categories), len(categories))
	prevCategoryName := ""
	for _, value := range categories {
		require.True(t, containsCategory(t, value, s.categories))
		require.GreaterOrEqual(t, value.Name, prevCategoryName)
		prevCategoryName = value.Name
	}
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_Limit() {
	t := s.T()

	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       5,
			OrderBy:     utils.AscOrder,
			OrderColumn: category.FieldID,
		},
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, filter.Limit, len(categories))
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_Offset() {
	t := s.T()

	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       5,
			OrderBy:     utils.AscOrder,
			OrderColumn: category.FieldID,
		},
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.categories)-filter.Offset, len(categories))
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_HasEquipment() {
	t := s.T()

	filter := domain.CategoryFilter{
		HasEquipments: true,
		Filter: domain.Filter{
			OrderBy:     utils.AscOrder,
			OrderColumn: category.FieldID,
		},
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, 1, len(categories))
}

func (s *categoryRepositorySuite) TestCategoryRepository_CategoryByID() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	category, err := s.repository.CategoryByID(ctx, s.categories[0].ID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, s.categories[0].Name, category.Name)
	require.Equal(t, s.categories[0].MaxReservationTime, category.MaxReservationTime)
	require.Equal(t, s.categories[0].MaxReservationUnits, category.MaxReservationUnits)
}

func (s *categoryRepositorySuite) TestCategoryRepository_DeleteCategoryByID() {
	t := s.T()
	name := "category"
	maxReservationTime := int64(10)
	maxReservationUnits := int64(1)
	createdCategory, err := s.client.Category.Create().SetName(name).
		SetMaxReservationTime(maxReservationTime).
		SetMaxReservationUnits(maxReservationUnits).
		Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = s.repository.DeleteCategoryByID(ctx, createdCategory.ID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
}

func (s *categoryRepositorySuite) TestCategoryRepository_UpdateCategory() {
	t := s.T()
	name := "category 0"
	update := models.UpdateCategoryRequest{
		Name: &name,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	category, err := s.repository.UpdateCategory(ctx, s.categories[0].ID, update)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, name, category.Name)
	require.Equal(t, s.categories[0].MaxReservationTime, category.MaxReservationTime)
	require.Equal(t, s.categories[0].MaxReservationUnits, category.MaxReservationUnits)
}

func containsCategory(t *testing.T, eq *ent.Category, list []*ent.Category) bool {
	t.Helper()
	for _, v := range list {
		if eq.Name == v.Name && eq.ID == v.ID && eq.MaxReservationUnits == v.MaxReservationUnits &&
			eq.MaxReservationTime == v.MaxReservationTime {
			return true
		}
	}
	return false
}
