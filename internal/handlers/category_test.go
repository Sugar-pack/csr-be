package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/category"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/categories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func TestSetCategoryHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:categoryhandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	SetCategoryHandler(logger, api)
	require.NotEmpty(t, api.CategoriesCreateNewCategoryHandler)
	require.NotEmpty(t, api.CategoriesGetCategoryByIDHandler)
	require.NotEmpty(t, api.CategoriesDeleteCategoryHandler)
	require.NotEmpty(t, api.CategoriesGetAllCategoriesHandler)
	require.NotEmpty(t, api.CategoriesUpdateCategoryHandler)
}

type CategoryTestSuite struct {
	suite.Suite
	logger     *zap.Logger
	repository *mocks.CategoryRepository
	handler    *Category
}

func TestCategorySuite(t *testing.T) {
	suite.Run(t, new(CategoryTestSuite))
}

func (s *CategoryTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.repository = &mocks.CategoryRepository{}
	s.handler = NewCategory(s.logger)
}

func (s *CategoryTestSuite) TestCategory_CreateCategory_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	maxReservationTime := int64(100)
	maxReservationUnit := int64(1)
	categoryName := "test"
	newCategory := models.CreateNewCategory{
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnit,
		Name:                &categoryName,
	}
	data := categories.CreateNewCategoryParams{
		HTTPRequest: &request,
		NewCategory: &newCategory,
	}
	err := errors.New("test")
	s.repository.On("CreateCategory", ctx, newCategory).Return(nil, err)

	handlerFunc := s.handler.CreateNewCategoryFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_CreateCategory_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	maxReservationTime := int64(100)
	maxReservationUnit := int64(1)
	categoryName := "test"
	hasSubcategories := true
	newCategory := models.CreateNewCategory{
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnit,
		Name:                &categoryName,
		HasSubcategory:      &hasSubcategories,
	}
	data := categories.CreateNewCategoryParams{
		HTTPRequest: &request,
		NewCategory: &newCategory,
	}

	categoryToReturn := &ent.Category{
		ID:                  1,
		Name:                categoryName,
		MaxReservationTime:  maxReservationTime,
		MaxReservationUnits: maxReservationUnit,
		HasSubcategory:      hasSubcategories,
	}
	s.repository.On("CreateCategory", ctx, newCategory).Return(categoryToReturn, nil)

	handlerFunc := s.handler.CreateNewCategoryFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusCreated, responseRecorder.Code)

	returnedCategory := models.CreateNewCategoryResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedCategory)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, categoryToReturn.ID, int(*returnedCategory.Data.ID))
	require.Equal(t, categoryToReturn.Name, *returnedCategory.Data.Name)
	require.Equal(t, categoryToReturn.MaxReservationTime, *returnedCategory.Data.MaxReservationTime)
	require.Equal(t, categoryToReturn.MaxReservationUnits, *returnedCategory.Data.MaxReservationUnits)
	require.Equal(t, categoryToReturn.HasSubcategory, *returnedCategory.Data.HasSubcategory)

	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_GetAllCategories_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := categories.GetAllCategoriesParams{
		HTTPRequest: &request,
	}

	err := errors.New("test")
	s.repository.On("AllCategoriesTotal", ctx).Return(0, err)

	handlerFunc := s.handler.GetAllCategoriesFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_GetAllCategories_NotFound() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := categories.GetAllCategoriesParams{
		HTTPRequest: &request,
	}

	s.repository.On("AllCategoriesTotal", ctx).Return(0, nil)

	handlerFunc := s.handler.GetAllCategoriesFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var returnedCategories models.ListOfCategories
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedCategories)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 0, int(*returnedCategories.Total))
	require.Equal(t, 0, len(returnedCategories.Items))

	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_GetAllCategories_EmptyParams() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			OrderBy:     utils.AscOrder,
			OrderColumn: category.FieldID,
		},
	}
	data := categories.GetAllCategoriesParams{
		HTTPRequest: &request,
	}

	categoriesToReturn := []*ent.Category{
		validCategory(t, 1),
		validCategory(t, 2),
		validCategory(t, 3),
	}

	s.repository.On("AllCategoriesTotal", ctx).Return(len(categoriesToReturn), nil)
	s.repository.On("AllCategories", ctx, filter).Return(categoriesToReturn, nil)

	handlerFunc := s.handler.GetAllCategoriesFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var returnedCategories models.ListOfCategories
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedCategories)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(categoriesToReturn), int(*returnedCategories.Total))
	require.Equal(t, len(categoriesToReturn), len(returnedCategories.Items))

	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_GetAllCategories_LimitGreaterThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	limit := int64(10)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := category.FieldID

	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       10,
			OrderBy:     utils.AscOrder,
			OrderColumn: category.FieldID,
		},
	}
	data := categories.GetAllCategoriesParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	categoriesToReturn := []*ent.Category{
		validCategory(t, 1),
		validCategory(t, 2),
		validCategory(t, 3),
	}

	s.repository.On("AllCategoriesTotal", ctx).Return(len(categoriesToReturn), nil)
	s.repository.On("AllCategories", ctx, filter).Return(categoriesToReturn, nil)

	handlerFunc := s.handler.GetAllCategoriesFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var returnedCategories models.ListOfCategories
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedCategories)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(categoriesToReturn), int(*returnedCategories.Total))
	require.Equal(t, len(categoriesToReturn), len(returnedCategories.Items))
	require.GreaterOrEqual(t, int(limit), len(returnedCategories.Items))

	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_GetAllCategories_LimitLessThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	limit := int64(2)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := category.FieldID
	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       2,
			OrderBy:     utils.AscOrder,
			OrderColumn: category.FieldID,
		},
	}

	data := categories.GetAllCategoriesParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	categoriesToReturn := []*ent.Category{
		validCategory(t, 1),
		validCategory(t, 2),
		validCategory(t, 3),
	}

	s.repository.On("AllCategoriesTotal", ctx).Return(len(categoriesToReturn), nil)
	s.repository.On("AllCategories", ctx, filter).Return(categoriesToReturn[:limit], nil)

	handlerFunc := s.handler.GetAllCategoriesFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var returnedCategories models.ListOfCategories
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedCategories)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(categoriesToReturn), int(*returnedCategories.Total))
	require.Greater(t, len(categoriesToReturn), len(returnedCategories.Items))
	require.GreaterOrEqual(t, int(limit), len(returnedCategories.Items))

	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_GetAllCategories_SecondPage() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	limit := int64(2)
	offset := int64(2)
	orderBy := utils.AscOrder
	orderColumn := category.FieldID
	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       2,
			Offset:      2,
			OrderBy:     utils.AscOrder,
			OrderColumn: category.FieldID,
		},
	}

	data := categories.GetAllCategoriesParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	categoriesToReturn := []*ent.Category{
		validCategory(t, 1),
		validCategory(t, 2),
		validCategory(t, 3),
	}

	s.repository.On("AllCategoriesTotal", ctx).Return(len(categoriesToReturn), nil)
	s.repository.On("AllCategories", ctx, filter).Return(categoriesToReturn[offset:], nil)

	handlerFunc := s.handler.GetAllCategoriesFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var returnedCategories models.ListOfCategories
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedCategories)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(categoriesToReturn), int(*returnedCategories.Total))
	require.Greater(t, len(categoriesToReturn), len(returnedCategories.Items))
	require.GreaterOrEqual(t, int(limit), len(returnedCategories.Items))
	require.Equal(t, len(categoriesToReturn)-int(offset), len(returnedCategories.Items))

	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_GetAllCategories_SeveralPages() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	limit := int64(3)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := category.FieldID
	filter := domain.CategoryFilter{
		Filter: domain.Filter{
			Limit:       3,
			OrderBy:     utils.AscOrder,
			OrderColumn: category.FieldID,
		},
	}
	data := categories.GetAllCategoriesParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	categoriesToReturn := []*ent.Category{
		validCategory(t, 1),
		validCategory(t, 2),
		validCategory(t, 3),
		validCategory(t, 4),
		validCategory(t, 5),
	}

	s.repository.On("AllCategoriesTotal", ctx).Return(len(categoriesToReturn), nil)
	s.repository.On("AllCategories", ctx, filter).Return(categoriesToReturn[:limit], nil)

	handlerFunc := s.handler.GetAllCategoriesFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var firstPage models.ListOfCategories
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &firstPage)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(categoriesToReturn), int(*firstPage.Total))
	require.Greater(t, len(categoriesToReturn), len(firstPage.Items))
	require.Equal(t, int(limit), len(firstPage.Items))

	offset = limit
	filter.Offset = filter.Limit
	data = categories.GetAllCategoriesParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}
	s.repository.On("AllCategoriesTotal", ctx).Return(len(categoriesToReturn), nil)
	s.repository.On("AllCategories", ctx, filter).Return(categoriesToReturn[offset:], nil)

	resp = handlerFunc.Handle(data, nil)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var secondPage models.ListOfCategories
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &secondPage)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(categoriesToReturn), int(*secondPage.Total))
	require.Greater(t, len(categoriesToReturn), len(secondPage.Items))
	require.GreaterOrEqual(t, int(limit), len(secondPage.Items))
	require.Equal(t, len(categoriesToReturn)-int(offset), len(secondPage.Items))

	require.False(t, categoriesDuplicated(t, firstPage.Items, secondPage.Items))
	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_GetCategoryByID_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := categories.GetCategoryByIDParams{
		HTTPRequest: &request,
		CategoryID:  1,
	}

	err := errors.New("test")
	s.repository.On("CategoryByID", ctx, int(data.CategoryID)).Return(nil, err)

	handlerFunc := s.handler.GetCategoryByIDFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_GetCategoryByID_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := categories.GetCategoryByIDParams{
		HTTPRequest: &request,
		CategoryID:  1,
	}

	categoryToReturn := &ent.Category{
		ID:                  1,
		Name:                "test",
		MaxReservationTime:  100,
		MaxReservationUnits: 1,
		HasSubcategory:      true,
	}
	s.repository.On("CategoryByID", ctx, int(data.CategoryID)).Return(categoryToReturn, nil)

	handlerFunc := s.handler.GetCategoryByIDFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	returnedCategory := models.GetCategoryByIDResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedCategory)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, categoryToReturn.ID, int(*returnedCategory.Data.ID))
	require.Equal(t, categoryToReturn.Name, *returnedCategory.Data.Name)
	require.Equal(t, categoryToReturn.MaxReservationTime, *returnedCategory.Data.MaxReservationTime)
	require.Equal(t, categoryToReturn.MaxReservationUnits, *returnedCategory.Data.MaxReservationUnits)
	require.Equal(t, categoryToReturn.HasSubcategory, *returnedCategory.Data.HasSubcategory)

	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_DeleteCategory_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := categories.DeleteCategoryParams{
		HTTPRequest: &request,
		CategoryID:  1,
	}

	err := errors.New("test")
	s.repository.On("DeleteCategoryByID", ctx, int(data.CategoryID)).Return(err)

	handlerFunc := s.handler.DeleteCategoryFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_DeleteCategory_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := categories.DeleteCategoryParams{
		HTTPRequest: &request,
		CategoryID:  1,
	}

	s.repository.On("DeleteCategoryByID", ctx, int(data.CategoryID)).Return(nil)

	handlerFunc := s.handler.DeleteCategoryFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_UpdateCategory_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	name := "test"
	update := models.UpdateCategoryRequest{
		Name: &name,
	}
	data := categories.UpdateCategoryParams{
		HTTPRequest:    &request,
		CategoryID:     1,
		UpdateCategory: &update,
	}

	err := errors.New("test")
	s.repository.On("UpdateCategory", ctx, int(data.CategoryID), update).Return(nil, err)

	handlerFunc := s.handler.UpdateCategoryFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *CategoryTestSuite) TestCategory_UpdateCategory_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	name := "test"
	hasSubcategories := true
	maxReservationTime := int64(100)
	maxReservationUnits := int64(1)
	update := models.UpdateCategoryRequest{
		HasSubcategory:      &hasSubcategories,
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnits,
		Name:                &name,
	}
	data := categories.UpdateCategoryParams{
		HTTPRequest:    &request,
		CategoryID:     1,
		UpdateCategory: &update,
	}

	updatedCategory := &ent.Category{
		ID:                  1,
		Name:                *update.Name,
		MaxReservationTime:  100,
		MaxReservationUnits: 1,
	}
	s.repository.On("UpdateCategory", ctx, int(data.CategoryID), update).Return(updatedCategory, nil)

	handlerFunc := s.handler.UpdateCategoryFunc(s.repository)

	resp := handlerFunc.Handle(data, nil)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	returnedCategory := models.UpdateCategoryResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedCategory)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, updatedCategory.ID, int(*returnedCategory.Data.ID))
	require.Equal(t, updatedCategory.Name, *returnedCategory.Data.Name)
	require.Equal(t, updatedCategory.MaxReservationTime, *returnedCategory.Data.MaxReservationTime)
	require.Equal(t, updatedCategory.MaxReservationUnits, *returnedCategory.Data.MaxReservationUnits)
	require.Equal(t, updatedCategory.HasSubcategory, *returnedCategory.Data.HasSubcategory)

	s.repository.AssertExpectations(t)
}

func validCategory(t *testing.T, id int) *ent.Category {
	t.Helper()
	return &ent.Category{
		ID:                  id,
		Name:                fmt.Sprintf("category %d", id),
		MaxReservationTime:  10,
		MaxReservationUnits: 5,
		HasSubcategory:      true,
	}
}
func categoriesDuplicated(t *testing.T, array1, array2 []*models.Category) bool {
	t.Helper()
	diff := make(map[int64]int, len(array1))
	for _, v := range array1 {
		diff[*v.ID] = 1
	}
	for _, v := range array2 {
		if _, ok := diff[*v.ID]; ok {
			return true
		}
	}
	return false
}
