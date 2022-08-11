package handlers

import (
	"encoding/json"
	"errors"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/activearea"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/active_areas"
)

func TestSetActiveAreaHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:activeareashandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	SetActiveAreaHandler(client, logger, api)
	assert.NotEmpty(t, api.ActiveAreasGetAllActiveAreasHandler)
}

type ActiveAreaTestSuite struct {
	suite.Suite
	logger     *zap.Logger
	repository *repomock.ActiveAreaRepository
	handler    *ActiveArea
	areas      []*ent.ActiveArea
}

func TestActiveAreaSuite(t *testing.T) {
	suite.Run(t, new(ActiveAreaTestSuite))
}

func (s *ActiveAreaTestSuite) SetupTest() {
	s.logger, _ = zap.NewDevelopment()
	s.repository = &repomock.ActiveAreaRepository{}
	s.handler = NewActiveArea(s.logger)
	s.areas = []*ent.ActiveArea{
		{
			ID:   1,
			Name: "test1",
		},
		{
			ID:   2,
			Name: "test2",
		},
		{
			ID:   3,
			Name: "test3",
		},
		{
			ID:   4,
			Name: "test4",
		},
		{
			ID:   5,
			Name: "test5",
		},
	}
}

func (s *ActiveAreaTestSuite) TestActiveArea_GetActiveAreasFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.handler.GetActiveAreasFunc(s.repository)
	data := active_areas.GetAllActiveAreasParams{
		HTTPRequest: &request,
	}

	err := errors.New("some error")
	s.repository.On("TotalActiveAreas", ctx).Return(0, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *ActiveAreaTestSuite) TestActiveArea_GetActiveAreasFunc_LimitGreaterThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := int64(10)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := activearea.FieldID

	handlerFunc := s.handler.GetActiveAreasFunc(s.repository)
	data := active_areas.GetAllActiveAreasParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	s.repository.On("TotalActiveAreas", ctx).Return(5, nil)
	s.repository.On("AllActiveAreas", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(s.areas, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var responseAreas models.ListOfActiveAreas
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseAreas)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(s.areas), len(responseAreas.Items))
	assert.Equal(t, len(s.areas), int(*responseAreas.Total))
	s.repository.AssertExpectations(t)
}

func (s *ActiveAreaTestSuite) TestActiveArea_GetActiveAreasFunc_LimitLessThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := int64(3)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := activearea.FieldID

	handlerFunc := s.handler.GetActiveAreasFunc(s.repository)
	data := active_areas.GetAllActiveAreasParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	s.repository.On("TotalActiveAreas", ctx).Return(5, nil)
	s.repository.On("AllActiveAreas", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(s.areas[:limit], nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var responseAreas models.ListOfActiveAreas
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseAreas)
	if err != nil {
		t.Fatal(err)
	}
	assert.Greater(t, len(s.areas), len(responseAreas.Items))
	assert.Equal(t, len(s.areas), int(*responseAreas.Total))
	assert.Equal(t, int(limit), len(responseAreas.Items))
	s.repository.AssertExpectations(t)
}

func (s *ActiveAreaTestSuite) TestActiveArea_GetActiveAreasFunc_SecondPage() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := int64(3)
	offset := int64(3)
	orderBy := utils.AscOrder
	orderColumn := activearea.FieldID

	handlerFunc := s.handler.GetActiveAreasFunc(s.repository)
	data := active_areas.GetAllActiveAreasParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	s.repository.On("TotalActiveAreas", ctx).Return(5, nil)
	s.repository.On("AllActiveAreas", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(s.areas[offset:], nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var responseAreas models.ListOfActiveAreas
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseAreas)
	if err != nil {
		t.Fatal(err)
	}
	assert.Greater(t, len(s.areas), len(responseAreas.Items))
	assert.Equal(t, len(s.areas), int(*responseAreas.Total))
	assert.GreaterOrEqual(t, int(limit), len(responseAreas.Items))
	s.repository.AssertExpectations(t)
}

func (s *ActiveAreaTestSuite) TestActiveArea_GetActiveAreasFunc_EmptyPaginationParams() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	var limit = math.MaxInt
	var offset = 0
	orderBy := utils.AscOrder
	orderColumn := activearea.FieldID

	handlerFunc := s.handler.GetActiveAreasFunc(s.repository)
	data := active_areas.GetAllActiveAreasParams{
		HTTPRequest: &request,
	}

	s.repository.On("TotalActiveAreas", ctx).Return(5, nil)
	s.repository.On("AllActiveAreas", ctx, limit, offset, orderBy, orderColumn).Return(s.areas, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var responseAreas models.ListOfActiveAreas
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseAreas)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(s.areas), len(responseAreas.Items))
	assert.Equal(t, len(s.areas), int(*responseAreas.Total))
	s.repository.AssertExpectations(t)
}

func (s *ActiveAreaTestSuite) TestActiveArea_GetActiveAreasFunc_SeveralPages() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := int64(3)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := activearea.FieldID

	handlerFunc := s.handler.GetActiveAreasFunc(s.repository)
	data := active_areas.GetAllActiveAreasParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	s.repository.On("TotalActiveAreas", ctx).Return(5, nil)
	s.repository.On("AllActiveAreas", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(s.areas[:limit], nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var responseAreasFirstPage models.ListOfActiveAreas
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseAreasFirstPage)
	if err != nil {
		t.Fatal(err)
	}
	assert.Greater(t, len(s.areas), len(responseAreasFirstPage.Items))
	assert.Equal(t, len(s.areas), int(*responseAreasFirstPage.Total))
	assert.GreaterOrEqual(t, int(limit), len(responseAreasFirstPage.Items))

	offset = limit
	s.repository.On("TotalActiveAreas", ctx).Return(5, nil)
	s.repository.On("AllActiveAreas", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(s.areas[offset:], nil)

	resp = handlerFunc(data, access)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var responseAreasSecondPage models.ListOfActiveAreas
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &responseAreasSecondPage)
	if err != nil {
		t.Fatal(err)
	}
	assert.Greater(t, len(s.areas), len(responseAreasSecondPage.Items))
	assert.Equal(t, len(s.areas), int(*responseAreasSecondPage.Total))
	assert.GreaterOrEqual(t, int(limit), len(responseAreasFirstPage.Items))

	assert.Equal(t, len(s.areas), len(responseAreasFirstPage.Items)+len(responseAreasSecondPage.Items))
	assert.False(t, areasDuplicated(t, responseAreasFirstPage.Items, responseAreasSecondPage.Items))
	s.repository.AssertExpectations(t)
}

func areasDuplicated(t *testing.T, array1, array2 []*models.ActiveArea) bool {
	t.Helper()
	diff := make(map[string]int, len(array1))
	for _, v := range array1 {
		diff[*v.Name] = 1
	}
	for _, v := range array2 {
		if _, ok := diff[*v.Name]; ok {
			return true
		}
	}
	return false
}
