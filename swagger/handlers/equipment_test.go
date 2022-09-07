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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	servicesmock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/services"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/equipment"
)

func TestSetEquipmentHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:equipmenthandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)

	fileManager := &servicesmock.FileManager{}
	SetEquipmentHandler(client, logger, api, fileManager)
	assert.NotEmpty(t, api.EquipmentCreateNewEquipmentHandler)
	assert.NotEmpty(t, api.EquipmentGetEquipmentHandler)
	assert.NotEmpty(t, api.EquipmentEditEquipmentHandler)
	assert.NotEmpty(t, api.EquipmentDeleteEquipmentHandler)
	assert.NotEmpty(t, api.EquipmentGetAllEquipmentHandler)
	assert.NotEmpty(t, api.EquipmentFindEquipmentHandler)
}

type EquipmentTestSuite struct {
	suite.Suite
	logger        *zap.Logger
	equipmentRepo *repomock.EquipmentRepository
	equipment     *Equipment
	fileManager   *servicesmock.FileManager
}

func InvalidEquipment(t *testing.T) *ent.Equipment {
	t.Helper()
	return &ent.Equipment{
		ID:   1,
		Name: "no edges",
	}
}

func ValidEquipment(t *testing.T, id int) *ent.Equipment {
	t.Helper()
	return &ent.Equipment{
		ID:   id,
		Name: fmt.Sprintf("test equipment %d", id),
		Edges: ent.EquipmentEdges{
			Category: &ent.Category{},
			Status:   &ent.Statuses{},
			Photo: &ent.Photo{
				ID:       "photoid",
				URL:      "localhost:8080/api/photoid",
				FileName: "photoid.jpg",
			},
		},
	}
}

func TestEquipmentSuite(t *testing.T) {
	suite.Run(t, new(EquipmentTestSuite))
}

func (s *EquipmentTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.equipmentRepo = &repomock.EquipmentRepository{}
	s.equipment = NewEquipment(s.logger)
	s.fileManager = &servicesmock.FileManager{}
}

func (s *EquipmentTestSuite) TestEquipment_PostEquipmentFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.PostEquipmentFunc(s.equipmentRepo)
	equipmentToAdd := models.Equipment{
		NameSubstring: "test",
	}
	data := equipment.CreateNewEquipmentParams{
		HTTPRequest:  &request,
		NewEquipment: &equipmentToAdd,
	}
	err := errors.New("test error")

	s.equipmentRepo.On("CreateEquipment", ctx, equipmentToAdd).Return(nil, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_PostEquipmentFunc_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.PostEquipmentFunc(s.equipmentRepo)
	equipmentToAdd := models.Equipment{
		NameSubstring: "test",
	}
	data := equipment.CreateNewEquipmentParams{
		HTTPRequest:  &request,
		NewEquipment: &equipmentToAdd,
	}

	equipmentToReturn := InvalidEquipment(t)

	s.equipmentRepo.On("CreateEquipment", ctx, equipmentToAdd).Return(equipmentToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_PostEquipmentFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.PostEquipmentFunc(s.equipmentRepo)
	equipmentToAdd := models.Equipment{
		NameSubstring: "test",
	}
	data := equipment.CreateNewEquipmentParams{
		HTTPRequest:  &request,
		NewEquipment: &equipmentToAdd,
	}

	equipmentToReturn := ValidEquipment(t, 1)

	s.equipmentRepo.On("CreateEquipment", ctx, equipmentToAdd).Return(equipmentToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	actualEquipment := &models.Equipment{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), actualEquipment)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Equal(t, equipmentToReturn.Name, *actualEquipment.Name)
}

func (s *EquipmentTestSuite) TestEquipment_GetEquipmentFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.GetEquipmentFunc(s.equipmentRepo)
	equipmentId := int64(1)
	data := equipment.GetEquipmentParams{
		HTTPRequest: &request,
		EquipmentID: equipmentId,
	}
	err := errors.New("test error")

	s.equipmentRepo.On("EquipmentByID", ctx, int(equipmentId)).Return(nil, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_GetEquipmentFunc_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.GetEquipmentFunc(s.equipmentRepo)
	equipmentId := int64(1)
	data := equipment.GetEquipmentParams{
		HTTPRequest: &request,
		EquipmentID: equipmentId,
	}
	equipmentToReturn := InvalidEquipment(t)

	s.equipmentRepo.On("EquipmentByID", ctx, int(equipmentId)).Return(equipmentToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_GetEquipmentFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.GetEquipmentFunc(s.equipmentRepo)
	equipmentId := int64(1)
	data := equipment.GetEquipmentParams{
		HTTPRequest: &request,
		EquipmentID: equipmentId,
	}
	equipmentToReturn := ValidEquipment(t, 1)

	s.equipmentRepo.On("EquipmentByID", ctx, int(equipmentId)).Return(equipmentToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	actualEquipment := &models.Equipment{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), actualEquipment)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Equal(t, equipmentToReturn.Name, *actualEquipment.Name)
}

func (s *EquipmentTestSuite) TestEquipment_DeleteEquipmentFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.DeleteEquipmentFunc(s.equipmentRepo, s.fileManager)
	equipmentId := int64(1)
	data := equipment.DeleteEquipmentParams{
		HTTPRequest: &request,
		EquipmentID: equipmentId,
	}
	err := errors.New("test error")

	equipmentToReturn := ValidEquipment(t, 1)
	s.equipmentRepo.On("EquipmentByID", ctx, int(equipmentId)).Return(equipmentToReturn, nil)
	s.equipmentRepo.On("DeleteEquipmentByID", ctx, int(equipmentId)).Return(err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_DeleteEquipmentFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.DeleteEquipmentFunc(s.equipmentRepo, s.fileManager)
	equipmentId := int64(1)
	data := equipment.DeleteEquipmentParams{
		HTTPRequest: &request,
		EquipmentID: equipmentId,
	}

	equipmentToReturn := ValidEquipment(t, 1)
	s.equipmentRepo.On("EquipmentByID", ctx, int(equipmentId)).Return(equipmentToReturn, nil)
	s.equipmentRepo.On("DeleteEquipmentByID", ctx, int(equipmentId)).Return(nil)
	s.equipmentRepo.On("DeleteEquipmentPhoto", ctx, equipmentToReturn.Edges.Photo.ID).Return(nil)
	s.fileManager.On("DeleteFile", equipmentToReturn.Edges.Photo.FileName).Return(nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_ListEquipmentFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.ListEquipmentFunc(s.equipmentRepo)
	data := equipment.GetAllEquipmentParams{
		HTTPRequest: &request,
	}
	err := errors.New("test error")
	s.equipmentRepo.On("AllEquipmentsTotal", ctx).Return(0, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_ListEquipmentFunc_NotFound() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.ListEquipmentFunc(s.equipmentRepo)
	data := equipment.GetAllEquipmentParams{
		HTTPRequest: &request,
	}
	s.equipmentRepo.On("AllEquipmentsTotal", ctx).Return(0, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)

	var responseEquipments models.ListEquipment
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseEquipments)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	assert.Equal(t, 0, int(*responseEquipments.Total))
	assert.Equal(t, 0, len(responseEquipments.Items))
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_ListEquipmentFunc_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := math.MaxInt
	offset := 0
	orderBy := "asc"
	orderColumn := "id"

	handlerFunc := s.equipment.ListEquipmentFunc(s.equipmentRepo)
	data := equipment.GetAllEquipmentParams{
		HTTPRequest: &request,
	}
	var equipmentToReturn []*ent.Equipment
	equipmentToReturn = append(equipmentToReturn, InvalidEquipment(t))
	s.equipmentRepo.On("AllEquipmentsTotal", ctx).Return(1, nil)
	s.equipmentRepo.On("AllEquipments", ctx, limit, offset, orderBy, orderColumn).Return(equipmentToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_ListEquipmentFunc_EmptyPaginationParams() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := math.MaxInt
	offset := 0
	orderBy := "asc"
	orderColumn := "id"

	handlerFunc := s.equipment.ListEquipmentFunc(s.equipmentRepo)
	data := equipment.GetAllEquipmentParams{
		HTTPRequest: &request,
	}
	var equipmentToReturn []*ent.Equipment
	equipmentToReturn = append(equipmentToReturn, ValidEquipment(t, 1))
	s.equipmentRepo.On("AllEquipmentsTotal", ctx).Return(1, nil)
	s.equipmentRepo.On("AllEquipments", ctx, limit, offset, orderBy, orderColumn).Return(equipmentToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	var responseEquipments models.ListEquipment
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseEquipments)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Equal(t, 1, int(*responseEquipments.Total))
	assert.Equal(t, equipmentToReturn[0].Name, *responseEquipments.Items[0].Name)
}

func (s *EquipmentTestSuite) TestEquipment_ListEquipmentFunc_LimitGreaterThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	var limit int64 = 10
	var offset int64 = 0
	var orderBy = "asc"
	var orderColumn = "id"

	handlerFunc := s.equipment.ListEquipmentFunc(s.equipmentRepo)
	data := equipment.GetAllEquipmentParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}
	var equipmentToReturn []*ent.Equipment
	equipmentToReturn = append(equipmentToReturn, ValidEquipment(t, 1))
	s.equipmentRepo.On("AllEquipmentsTotal", ctx).Return(1, nil)
	s.equipmentRepo.On("AllEquipments", ctx, int(limit), int(offset), orderBy, orderColumn).Return(equipmentToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	var responseEquipments models.ListEquipment
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseEquipments)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Equal(t, 1, int(*responseEquipments.Total))
	assert.Equal(t, equipmentToReturn[0].Name, *responseEquipments.Items[0].Name)
}

func (s *EquipmentTestSuite) TestEquipment_ListEquipmentFunc_LimitLessThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	var limit int64 = 3
	var offset int64 = 0
	var orderBy = "asc"
	var orderColumn = "id"

	handlerFunc := s.equipment.ListEquipmentFunc(s.equipmentRepo)
	data := equipment.GetAllEquipmentParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}
	equipmentToReturn := []*ent.Equipment{
		ValidEquipment(t, 1),
		ValidEquipment(t, 2),
		ValidEquipment(t, 3),
		ValidEquipment(t, 4),
		ValidEquipment(t, 5),
	}
	s.equipmentRepo.On("AllEquipmentsTotal", ctx).Return(5, nil)
	s.equipmentRepo.On("AllEquipments", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(equipmentToReturn[:limit], nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	var responseEquipments models.ListEquipment
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseEquipments)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Greater(t, len(equipmentToReturn), len(responseEquipments.Items))
	assert.Equal(t, len(equipmentToReturn), int(*responseEquipments.Total))
	assert.Equal(t, int(limit), len(responseEquipments.Items))
	for _, item := range responseEquipments.Items {
		assert.True(t, containsEquipment(t, equipmentToReturn, item))
	}
}

func (s *EquipmentTestSuite) TestEquipment_ListEquipmentFunc_SecondPage() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	var limit int64 = 3
	var offset int64 = 3
	var orderBy = "asc"
	var orderColumn = "id"

	handlerFunc := s.equipment.ListEquipmentFunc(s.equipmentRepo)
	data := equipment.GetAllEquipmentParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}
	equipmentToReturn := []*ent.Equipment{
		ValidEquipment(t, 1),
		ValidEquipment(t, 2),
		ValidEquipment(t, 3),
		ValidEquipment(t, 4),
		ValidEquipment(t, 5),
	}
	s.equipmentRepo.On("AllEquipmentsTotal", ctx).Return(5, nil)
	s.equipmentRepo.On("AllEquipments", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(equipmentToReturn[offset:], nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	var responseEquipments models.ListEquipment
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseEquipments)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Greater(t, len(equipmentToReturn), len(responseEquipments.Items))
	assert.Equal(t, len(equipmentToReturn), int(*responseEquipments.Total))
	assert.Equal(t, 2, len(responseEquipments.Items))
	for _, item := range responseEquipments.Items {
		assert.True(t, containsEquipment(t, equipmentToReturn, item))
	}
}

func (s *EquipmentTestSuite) TestEquipment_ListEquipmentFunc_SeveralPages() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	var limit int64 = 3
	var offset int64 = 0
	var orderBy = "asc"
	var orderColumn = "id"

	handlerFunc := s.equipment.ListEquipmentFunc(s.equipmentRepo)
	data := equipment.GetAllEquipmentParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}
	equipmentToReturn := []*ent.Equipment{
		ValidEquipment(t, 1),
		ValidEquipment(t, 2),
		ValidEquipment(t, 3),
		ValidEquipment(t, 4),
		ValidEquipment(t, 5),
	}
	s.equipmentRepo.On("AllEquipmentsTotal", ctx).Return(5, nil)
	s.equipmentRepo.On("AllEquipments", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(equipmentToReturn[:limit], nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	var responseFirstPage models.ListEquipment
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseFirstPage)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Greater(t, len(equipmentToReturn), len(responseFirstPage.Items))
	assert.Equal(t, len(equipmentToReturn), int(*responseFirstPage.Total))
	assert.Equal(t, 3, len(responseFirstPage.Items))
	for _, item := range responseFirstPage.Items {
		assert.True(t, containsEquipment(t, equipmentToReturn, item))
	}

	offset = limit
	s.equipmentRepo.On("AllEquipmentsTotal", ctx).Return(5, nil)
	s.equipmentRepo.On("AllEquipments", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(equipmentToReturn[offset:], nil)
	resp = handlerFunc(data, access)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	var responseSecondPage models.ListEquipment
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &responseSecondPage)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Greater(t, len(equipmentToReturn), len(responseSecondPage.Items))
	assert.Equal(t, len(equipmentToReturn), int(*responseSecondPage.Total))
	assert.Equal(t, 2, len(responseSecondPage.Items))
	for _, item := range responseSecondPage.Items {
		assert.True(t, containsEquipment(t, equipmentToReturn, item))
	}

	assert.Equal(t, len(equipmentToReturn), len(responseFirstPage.Items)+len(responseSecondPage.Items))
	assert.False(t, equipmentsDuplicated(t, responseFirstPage.Items, responseSecondPage.Items))
}

func (s *EquipmentTestSuite) TestEquipment_FindEquipmentFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.FindEquipmentFunc(s.equipmentRepo)
	equipmentFilter := models.EquipmentFilter{
		NameSubstring: "test",
	}
	data := equipment.FindEquipmentParams{
		HTTPRequest:   &request,
		FindEquipment: &equipmentFilter,
	}
	err := errors.New("test error")

	s.equipmentRepo.On("EquipmentsByFilterTotal", ctx, equipmentFilter).Return(0, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_FindEquipmentFunc_NoResult() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.FindEquipmentFunc(s.equipmentRepo)
	equipmentFilter := models.EquipmentFilter{
		NameSubstring: "test",
	}
	data := equipment.FindEquipmentParams{
		HTTPRequest:   &request,
		FindEquipment: &equipmentFilter,
	}

	s.equipmentRepo.On("EquipmentsByFilterTotal", ctx, equipmentFilter).Return(0, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var responseEquipments models.ListEquipment
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseEquipments)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Equal(t, 0, int(*responseEquipments.Total))
	assert.Equal(t, 0, len(responseEquipments.Items))
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_FindEquipmentFunc_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := math.MaxInt
	offset := 0
	orderBy := "asc"
	orderColumn := "id"

	handlerFunc := s.equipment.FindEquipmentFunc(s.equipmentRepo)
	equipmentFilter := models.EquipmentFilter{
		NameSubstring: "test",
	}
	data := equipment.FindEquipmentParams{
		HTTPRequest:   &request,
		FindEquipment: &equipmentFilter,
	}
	var equipmentToReturn []*ent.Equipment
	equipmentToReturn = append(equipmentToReturn, InvalidEquipment(t))

	s.equipmentRepo.On("EquipmentsByFilterTotal", ctx, equipmentFilter).Return(1, nil)
	s.equipmentRepo.On("EquipmentsByFilter", ctx, equipmentFilter, limit, offset, orderBy, orderColumn).
		Return(equipmentToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_FindEquipmentFunc_EmptyPaginationParams() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := math.MaxInt
	offset := 0
	orderBy := "asc"
	orderColumn := "id"

	handlerFunc := s.equipment.FindEquipmentFunc(s.equipmentRepo)
	equipmentFilter := models.EquipmentFilter{
		NameSubstring: "equipment 1",
	}
	data := equipment.FindEquipmentParams{
		HTTPRequest:   &request,
		FindEquipment: &equipmentFilter,
	}
	equipmentToReturn := []*ent.Equipment{
		ValidEquipment(t, 1),
		ValidEquipment(t, 11),
		ValidEquipment(t, 2),
		ValidEquipment(t, 3),
		ValidEquipment(t, 4),
	}
	s.equipmentRepo.On("EquipmentsByFilterTotal", ctx, equipmentFilter).Return(2, nil)
	s.equipmentRepo.On("EquipmentsByFilter", ctx, equipmentFilter, limit, offset, orderBy, orderColumn).
		Return(equipmentToReturn[:2], nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	var responseEquipments models.ListEquipment
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseEquipments)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Equal(t, 2, int(*responseEquipments.Total))
	assert.Equal(t, 2, len(responseEquipments.Items))
	assert.Equal(t, equipmentToReturn[0].Name, *responseEquipments.Items[0].Name)
	assert.Equal(t, equipmentToReturn[1].Name, *responseEquipments.Items[1].Name)
}

func (s *EquipmentTestSuite) TestEquipment_FindEquipmentFunc_LimitGreaterThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	var limit int64 = 10
	var offset int64 = 0
	orderBy := "asc"
	orderColumn := "id"

	handlerFunc := s.equipment.FindEquipmentFunc(s.equipmentRepo)
	equipmentFilter := models.EquipmentFilter{
		NameSubstring: "equipment 1",
	}
	data := equipment.FindEquipmentParams{
		HTTPRequest:   &request,
		FindEquipment: &equipmentFilter,
		Limit:         &limit,
		Offset:        &offset,
		OrderBy:       &orderBy,
		OrderColumn:   &orderColumn,
	}
	equipmentToReturn := []*ent.Equipment{
		ValidEquipment(t, 1),
		ValidEquipment(t, 11),
		ValidEquipment(t, 2),
		ValidEquipment(t, 3),
		ValidEquipment(t, 4),
	}

	s.equipmentRepo.On("EquipmentsByFilterTotal", ctx, equipmentFilter).Return(2, nil)
	s.equipmentRepo.On("EquipmentsByFilter", ctx, equipmentFilter, int(limit), int(offset), orderBy, orderColumn).
		Return(equipmentToReturn[:2], nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	var actualEquipment models.ListEquipment
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &actualEquipment)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Equal(t, 2, int(*actualEquipment.Total))
	assert.Equal(t, 2, len(actualEquipment.Items))
	assert.Equal(t, equipmentToReturn[0].Name, *actualEquipment.Items[0].Name)
}

func (s *EquipmentTestSuite) TestEquipment_FindEquipmentFunc_LimitLessThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	var limit int64 = 2
	var offset int64 = 0
	orderBy := "asc"
	orderColumn := "id"

	handlerFunc := s.equipment.FindEquipmentFunc(s.equipmentRepo)
	equipmentFilter := models.EquipmentFilter{
		NameSubstring: "1",
	}
	data := equipment.FindEquipmentParams{
		HTTPRequest:   &request,
		FindEquipment: &equipmentFilter,
		Limit:         &limit,
		Offset:        &offset,
		OrderBy:       &orderBy,
		OrderColumn:   &orderColumn,
	}
	equipmentToReturn := []*ent.Equipment{
		ValidEquipment(t, 1),
		ValidEquipment(t, 11),
		ValidEquipment(t, 111),
		ValidEquipment(t, 3),
		ValidEquipment(t, 4),
	}

	s.equipmentRepo.On("EquipmentsByFilterTotal", ctx, equipmentFilter).Return(3, nil)
	s.equipmentRepo.On("EquipmentsByFilter", ctx, equipmentFilter, int(limit), int(offset), orderBy, orderColumn).
		Return(equipmentToReturn[:2], nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	var actualEquipment models.ListEquipment
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &actualEquipment)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Equal(t, 3, int(*actualEquipment.Total))
	assert.Equal(t, 2, len(actualEquipment.Items))
	for _, item := range actualEquipment.Items {
		assert.True(t, containsEquipment(t, equipmentToReturn, item))
	}
}

func (s *EquipmentTestSuite) TestEquipment_FindEquipmentFunc_SecondPage() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	var limit int64 = 3
	var offset int64 = 3
	var orderBy = "asc"
	var orderColumn = "id"

	handlerFunc := s.equipment.FindEquipmentFunc(s.equipmentRepo)
	equipmentFilter := models.EquipmentFilter{
		NameSubstring: "test",
	}
	data := equipment.FindEquipmentParams{
		HTTPRequest:   &request,
		FindEquipment: &equipmentFilter,
		Limit:         &limit,
		Offset:        &offset,
		OrderBy:       &orderBy,
		OrderColumn:   &orderColumn,
	}
	equipmentToReturn := []*ent.Equipment{
		ValidEquipment(t, 1),
		ValidEquipment(t, 2),
		ValidEquipment(t, 3),
		ValidEquipment(t, 4),
		ValidEquipment(t, 5),
	}
	s.equipmentRepo.On("EquipmentsByFilterTotal", ctx, equipmentFilter).Return(5, nil)
	s.equipmentRepo.On("EquipmentsByFilter", ctx, equipmentFilter, int(limit), int(offset), orderBy, orderColumn).
		Return(equipmentToReturn[offset:], nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	var responseEquipments models.ListEquipment
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseEquipments)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Greater(t, len(equipmentToReturn), len(responseEquipments.Items))
	assert.Equal(t, len(equipmentToReturn), int(*responseEquipments.Total))
	assert.Equal(t, 2, len(responseEquipments.Items))
	for _, item := range responseEquipments.Items {
		assert.True(t, containsEquipment(t, equipmentToReturn, item))
	}
}

func (s *EquipmentTestSuite) TestEquipment_EditEquipmentFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.EditEquipmentFunc(s.equipmentRepo)
	equipmentId := int64(1)
	equipmentCategoryUpdate := int64(10)
	equipmentUpdate := &models.Equipment{
		Category: &equipmentCategoryUpdate,
	}
	data := equipment.EditEquipmentParams{
		HTTPRequest:   &request,
		EquipmentID:   equipmentId,
		EditEquipment: equipmentUpdate,
	}
	err := errors.New("test error")

	s.equipmentRepo.On("UpdateEquipmentByID", ctx, int(equipmentId), equipmentUpdate).
		Return(nil, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_EditEquipmentFunc_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.EditEquipmentFunc(s.equipmentRepo)
	equipmentId := int64(1)
	equipmentCategoryUpdate := int64(10)
	equipmentUpdate := &models.Equipment{
		Category: &equipmentCategoryUpdate,
	}
	data := equipment.EditEquipmentParams{
		HTTPRequest:   &request,
		EquipmentID:   equipmentId,
		EditEquipment: equipmentUpdate,
	}
	equipmentToReturn := InvalidEquipment(t)

	s.equipmentRepo.On("UpdateEquipmentByID", ctx, int(equipmentId), equipmentUpdate).
		Return(equipmentToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_EditEquipmentFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.EditEquipmentFunc(s.equipmentRepo)
	equipmentId := int64(1)
	equipmentKindUpdate := int64(10)
	equipmentUpdate := &models.Equipment{
		Category: &equipmentKindUpdate,
	}
	data := equipment.EditEquipmentParams{
		HTTPRequest:   &request,
		EquipmentID:   equipmentId,
		EditEquipment: equipmentUpdate,
	}
	equipmentToReturn := ValidEquipment(t, 1)

	s.equipmentRepo.On("UpdateEquipmentByID", ctx, int(equipmentId), equipmentUpdate).
		Return(equipmentToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var responseEquipment models.Equipment
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseEquipment)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Equal(t, equipmentToReturn.Name, *responseEquipment.Name)

	s.equipmentRepo.AssertExpectations(t)
}

func containsEquipment(t *testing.T, array []*ent.Equipment, item *models.EquipmentResponse) bool {
	t.Helper()
	for _, v := range array {
		if *item.Name == v.Name && int(*item.ID) == v.ID {
			return true
		}
	}
	return false
}

func equipmentsDuplicated(t *testing.T, array1, array2 []*models.EquipmentResponse) bool {
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
