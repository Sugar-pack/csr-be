package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/subcategories"
)

func TestSetSubcategoryHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:subcategoryhandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	SetSubcategoryHandler(logger, api)
	require.NotEmpty(t, api.SubcategoriesCreateNewSubcategoryHandler)
	require.NotEmpty(t, api.SubcategoriesGetSubcategoryByIDHandler)
	require.NotEmpty(t, api.SubcategoriesDeleteSubcategoryHandler)
	require.NotEmpty(t, api.SubcategoriesListSubcategoriesByCategoryIDHandler)
	require.NotEmpty(t, api.SubcategoriesUpdateSubcategoryHandler)
}

type SubcategoryTestSuite struct {
	suite.Suite
	logger     *zap.Logger
	repository *mocks.SubcategoryRepository
	handler    *Subcategory
}

func TestSubcategorySuite(t *testing.T) {
	suite.Run(t, new(SubcategoryTestSuite))
}

func (s *SubcategoryTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.repository = &mocks.SubcategoryRepository{}
	s.handler = NewSubcategory(s.logger)
}

func (s *SubcategoryTestSuite) TestSubcategory_CreateSubcategory_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	subcategoryName := "test subcategory"
	maxReservationTime := int64(100)
	maxReservationUnit := int64(1)
	categoryID := int64(10)
	newSubcategory := models.NewSubcategory{
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnit,
		Name:                &subcategoryName,
	}
	data := subcategories.CreateNewSubcategoryParams{
		HTTPRequest:    &request,
		NewSubcategory: &newSubcategory,
		CategoryID:     categoryID,
	}
	err := errors.New("category not exists")
	s.repository.On("CreateSubcategory", ctx, int(categoryID), newSubcategory).
		Return(nil, err)

	handlerFunc := s.handler.CreateNewSubcategoryFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *SubcategoryTestSuite) TestSubcategory_CreateSubcategory_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	subcategoryName := "test subcategory"
	maxReservationTime := int64(10)
	maxReservationUnit := int64(5)
	categoryID := int64(1)
	newSubcategory := models.NewSubcategory{
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnit,
		Name:                &subcategoryName,
	}
	data := subcategories.CreateNewSubcategoryParams{
		HTTPRequest:    &request,
		NewSubcategory: &newSubcategory,
		CategoryID:     categoryID,
	}
	subcategoryToReturn := validSubcategory(t, 1)
	s.repository.On("CreateSubcategory", ctx, int(categoryID), newSubcategory).
		Return(subcategoryToReturn, nil)

	handlerFunc := s.handler.CreateNewSubcategoryFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusCreated, responseRecorder.Code)

	returnedSubcategory := models.SubcategoryResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedSubcategory)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, subcategoryToReturn.ID, int(*returnedSubcategory.Data.ID))
	require.Equal(t, subcategoryToReturn.Name, *returnedSubcategory.Data.Name)
	require.Equal(t, subcategoryToReturn.MaxReservationTime, *returnedSubcategory.Data.MaxReservationTime)
	require.Equal(t, subcategoryToReturn.MaxReservationUnits, *returnedSubcategory.Data.MaxReservationUnits)

	s.repository.AssertExpectations(t)
}

func (s *SubcategoryTestSuite) TestSubcategory_GetAllSubcategories_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	categoryID := int64(10)
	data := subcategories.ListSubcategoriesByCategoryIDParams{
		HTTPRequest: &request,
		CategoryID:  categoryID,
	}

	err := errors.New("test")
	s.repository.On("ListSubcategories", ctx, int(categoryID)).
		Return(nil, err)

	handlerFunc := s.handler.ListSubcategoriesFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	returnedSubcategory := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(returnedSubcategory, producer)
	require.Equal(t, http.StatusInternalServerError, returnedSubcategory.Code)

	s.repository.AssertExpectations(t)
}

func (s *SubcategoryTestSuite) TestSubcategory_GetAllSubcategories_NotFound() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	categoryID := int64(3)
	data := subcategories.ListSubcategoriesByCategoryIDParams{
		HTTPRequest: &request,
		CategoryID:  categoryID,
	}

	s.repository.On("ListSubcategories", ctx, int(categoryID)).
		Return([]*ent.Subcategory{}, nil)

	handlerFunc := s.handler.ListSubcategoriesFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	returnedSubcategory := models.ListOfSubcategories{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedSubcategory)
	if err != nil {
		t.Fatal(err)
	}
	require.Empty(t, returnedSubcategory)

	s.repository.AssertExpectations(t)
}

func (s *SubcategoryTestSuite) TestSubcategory_GetAllSubcategories_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	categoryID := int64(1)
	data := subcategories.ListSubcategoriesByCategoryIDParams{
		HTTPRequest: &request,
		CategoryID:  categoryID,
	}

	subcategoriesToReturn := []*ent.Subcategory{
		validSubcategory(t, 1),
		validSubcategory(t, 2),
		validSubcategory(t, 3),
	}
	s.repository.On("ListSubcategories", ctx, int(categoryID)).
		Return(subcategoriesToReturn, nil)

	handlerFunc := s.handler.ListSubcategoriesFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	returnedSubcategory := models.ListOfSubcategories{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedSubcategory)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(subcategoriesToReturn), len(returnedSubcategory))

	s.repository.AssertExpectations(t)
}

func (s *SubcategoryTestSuite) TestSubcategory_GetSubcategoryByID_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := subcategories.GetSubcategoryByIDParams{
		HTTPRequest:   &request,
		SubcategoryID: 1,
	}

	err := errors.New("test")
	s.repository.On("SubcategoryByID", ctx, int(data.SubcategoryID)).
		Return(nil, err)

	handlerFunc := s.handler.GetSubcategoryByIDFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *SubcategoryTestSuite) TestSubcategory_GetSubcategoryByID_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := subcategories.GetSubcategoryByIDParams{
		HTTPRequest:   &request,
		SubcategoryID: 1,
	}

	subcatToReturn := validSubcategory(t, 1)
	s.repository.On("SubcategoryByID", ctx, int(data.SubcategoryID)).
		Return(subcatToReturn, nil)

	handlerFunc := s.handler.GetSubcategoryByIDFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	returnedSubcategory := models.SubcategoryResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedSubcategory)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, subcatToReturn.ID, int(*returnedSubcategory.Data.ID))
	require.Equal(t, subcatToReturn.Name, *returnedSubcategory.Data.Name)
	require.Equal(t, subcatToReturn.MaxReservationTime, *returnedSubcategory.Data.MaxReservationTime)
	require.Equal(t, subcatToReturn.MaxReservationUnits, *returnedSubcategory.Data.MaxReservationUnits)
	require.Equal(t, subcatToReturn.Edges.Category.ID, int(*returnedSubcategory.Data.Category))
	s.repository.AssertExpectations(t)
}

func (s *SubcategoryTestSuite) TestSubcategory_DeleteSubcategory_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := subcategories.DeleteSubcategoryParams{
		HTTPRequest:   &request,
		SubcategoryID: 1,
	}

	err := errors.New("test")
	s.repository.On("DeleteSubcategoryByID", ctx, int(data.SubcategoryID)).Return(err)

	handlerFunc := s.handler.DeleteSubcategoryFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *SubcategoryTestSuite) TestSubcategory_DeleteSubcategory_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := subcategories.DeleteSubcategoryParams{
		HTTPRequest:   &request,
		SubcategoryID: 1,
	}

	s.repository.On("DeleteSubcategoryByID", ctx, int(data.SubcategoryID)).Return(nil)

	handlerFunc := s.handler.DeleteSubcategoryFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *SubcategoryTestSuite) TestSubcategory_UpdateSubcategory_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	name := "test"
	update := models.NewSubcategory{
		Name: &name,
	}
	data := subcategories.UpdateSubcategoryParams{
		HTTPRequest:       &request,
		SubcategoryID:     1,
		UpdateSubcategory: &update,
	}

	err := errors.New("test")
	s.repository.On("UpdateSubcategory", ctx, int(data.SubcategoryID), update).
		Return(nil, err)

	handlerFunc := s.handler.UpdateSubcategoryFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *SubcategoryTestSuite) TestSubcategory_UpdateSubcategory_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	name := "test"
	update := models.NewSubcategory{
		Name: &name,
	}
	data := subcategories.UpdateSubcategoryParams{
		HTTPRequest:       &request,
		SubcategoryID:     1,
		UpdateSubcategory: &update,
	}

	updatedSubcategory := validSubcategory(t, 1)
	updatedSubcategory.Name = name
	s.repository.On("UpdateSubcategory", ctx, int(data.SubcategoryID), update).
		Return(updatedSubcategory, nil)

	handlerFunc := s.handler.UpdateSubcategoryFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	returnedSubcategory := models.SubcategoryResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedSubcategory)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, updatedSubcategory.ID, int(*returnedSubcategory.Data.ID))
	require.Equal(t, updatedSubcategory.MaxReservationUnits, *returnedSubcategory.Data.MaxReservationUnits)
	require.Equal(t, updatedSubcategory.MaxReservationTime, *returnedSubcategory.Data.MaxReservationTime)
	require.Equal(t, updatedSubcategory.Edges.Category.ID, int(*returnedSubcategory.Data.Category))

	s.repository.AssertExpectations(t)
}

func validSubcategory(t *testing.T, id int) *ent.Subcategory {
	t.Helper()
	return &ent.Subcategory{
		ID:                  id,
		Name:                fmt.Sprintf("subcategory %d", id),
		MaxReservationTime:  10,
		MaxReservationUnits: 5,
		Edges: ent.SubcategoryEdges{
			Category: validCategory(t, 1),
		},
	}
}
