package repositories

import (
	"context"
	"math"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/category"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type categoryRepositorySuite struct {
	suite.Suite
	ctx        context.Context
	client     *ent.Client
	repository CategoryRepository
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
	_, err := s.client.Category.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range s.categories {
		category, errCreate := s.client.Category.Create().
			SetName(value.Name).SetMaxReservationTime(value.MaxReservationTime).
			SetMaxReservationUnits(value.MaxReservationUnits).
			Save(s.ctx)
		if errCreate != nil {
			t.Fatal(errCreate)
		}
		s.categories[i].ID = category.ID
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
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	createdCategory, err := s.repository.CreateCategory(ctx, newCategory)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, name, createdCategory.Name)
	assert.Equal(t, maxReservationTime, createdCategory.MaxReservationTime)
	assert.Equal(t, maxReservationUnits, createdCategory.MaxReservationUnits)

	_, err = s.client.Category.Delete().Where(category.IDEQ(createdCategory.ID)).Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategoriesTotal() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	total, err := s.repository.AllCategoriesTotal(ctx)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.categories), total)
}
func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_EmptyOrderBy() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := ""
	orderColumn := category.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Commit())
	assert.Nil(t, categories)
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_EmptyOrderColumn() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := ""
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Commit())
	assert.Nil(t, categories)
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_WrongOrderColumn() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := category.FieldMaxReservationTime
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Commit())
	assert.Nil(t, categories)
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_OrderByIDDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := category.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.categories), len(categories))
	prevCategoryID := math.MaxInt
	for _, value := range categories {
		assert.True(t, containsCategory(t, value, s.categories))
		assert.LessOrEqual(t, value.ID, prevCategoryID)
		prevCategoryID = value.ID
	}
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_OrderByNameDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := category.FieldName
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.categories), len(categories))
	prevCategoryName := "zzzzzzzzzzzzzzzzzzzzz"
	for _, value := range categories {
		assert.True(t, containsCategory(t, value, s.categories))
		assert.LessOrEqual(t, value.Name, prevCategoryName)
		prevCategoryName = value.Name
	}
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_OrderByIDAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := category.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.categories), len(categories))
	prevCategoryID := 0
	for _, value := range categories {
		assert.True(t, containsCategory(t, value, s.categories))
		assert.GreaterOrEqual(t, value.ID, prevCategoryID)
		prevCategoryID = value.ID
	}
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_OrderByNameAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := category.FieldName
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.categories), len(categories))
	prevCategoryName := ""
	for _, value := range categories {
		assert.True(t, containsCategory(t, value, s.categories))
		assert.GreaterOrEqual(t, value.Name, prevCategoryName)
		prevCategoryName = value.Name
	}
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_Limit() {
	t := s.T()
	limit := 5
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := category.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, limit, len(categories))
}

func (s *categoryRepositorySuite) TestCategoryRepository_AllCategory_Offset() {
	t := s.T()
	limit := 0
	offset := 5
	orderBy := utils.AscOrder
	orderColumn := category.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	categories, err := s.repository.AllCategories(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.categories)-offset, len(categories))
}

func (s *categoryRepositorySuite) TestCategoryRepository_CategoryByID() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	category, err := s.repository.CategoryByID(ctx, s.categories[0].ID)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, s.categories[0].Name, category.Name)
	assert.Equal(t, s.categories[0].MaxReservationTime, category.MaxReservationTime)
	assert.Equal(t, s.categories[0].MaxReservationUnits, category.MaxReservationUnits)

	_, err = s.client.Category.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
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
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = s.repository.DeleteCategoryByID(ctx, createdCategory.ID)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
}

func (s *categoryRepositorySuite) TestCategoryRepository_UpdateCategory() {
	t := s.T()
	name := "category 0"
	update := models.UpdateCategoryRequest{
		Name: &name,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	category, err := s.repository.UpdateCategory(ctx, s.categories[0].ID, update)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, name, category.Name)
	assert.Equal(t, s.categories[0].MaxReservationTime, category.MaxReservationTime)
	assert.Equal(t, s.categories[0].MaxReservationUnits, category.MaxReservationUnits)
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
