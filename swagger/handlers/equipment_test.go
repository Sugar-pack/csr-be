package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	servicesmock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/services"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/equipment"
)

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

func ValidEquipment(t *testing.T) *ent.Equipment {
	t.Helper()
	return &ent.Equipment{
		ID:   1,
		Name: "test equipment",
		Edges: ent.EquipmentEdges{
			Kind:   &ent.Kind{},
			Status: &ent.Statuses{},
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

	resp := handlerFunc(data)
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

	resp := handlerFunc(data)
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

	equipmentToReturn := ValidEquipment(t)

	s.equipmentRepo.On("CreateEquipment", ctx, equipmentToAdd).Return(equipmentToReturn, nil)

	resp := handlerFunc(data)
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

	resp := handlerFunc(data)
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

	resp := handlerFunc(data)
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
	equipmentToReturn := ValidEquipment(t)

	s.equipmentRepo.On("EquipmentByID", ctx, int(equipmentId)).Return(equipmentToReturn, nil)

	resp := handlerFunc(data)
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

	equipmentToReturn := ValidEquipment(t)
	s.equipmentRepo.On("EquipmentByID", ctx, int(equipmentId)).Return(equipmentToReturn, nil)
	s.equipmentRepo.On("DeleteEquipmentByID", ctx, int(equipmentId)).Return(err)

	resp := handlerFunc(data)
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

	equipmentToReturn := ValidEquipment(t)
	s.equipmentRepo.On("EquipmentByID", ctx, int(equipmentId)).Return(equipmentToReturn, nil)
	s.equipmentRepo.On("DeleteEquipmentByID", ctx, int(equipmentId)).Return(nil)
	s.equipmentRepo.On("DeleteEquipmentPhoto", ctx, equipmentToReturn.Edges.Photo.ID).Return(nil)
	s.fileManager.On("DeleteFile", equipmentToReturn.Edges.Photo.FileName).Return(nil)

	resp := handlerFunc(data)
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
	s.equipmentRepo.On("AllEquipments", ctx).Return(nil, err)

	resp := handlerFunc(data)
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
	var equipmentToReturn []*ent.Equipment
	s.equipmentRepo.On("AllEquipments", ctx).Return(equipmentToReturn, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_ListEquipmentFunc_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.ListEquipmentFunc(s.equipmentRepo)
	data := equipment.GetAllEquipmentParams{
		HTTPRequest: &request,
	}
	var equipmentToReturn []*ent.Equipment
	equipmentToReturn = append(equipmentToReturn, InvalidEquipment(t))
	s.equipmentRepo.On("AllEquipments", ctx).Return(equipmentToReturn, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_ListEquipmentFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.ListEquipmentFunc(s.equipmentRepo)
	data := equipment.GetAllEquipmentParams{
		HTTPRequest: &request,
	}
	var equipmentToReturn []*ent.Equipment
	equipmentToReturn = append(equipmentToReturn, ValidEquipment(t))
	s.equipmentRepo.On("AllEquipments", ctx).Return(equipmentToReturn, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	actualEquipment := make([]models.Equipment, len(equipmentToReturn))
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &actualEquipment)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Equal(t, equipmentToReturn[0].Name, *actualEquipment[0].Name)
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

	s.equipmentRepo.On("EquipmentsByFilter", ctx, equipmentFilter).Return(nil, err)

	resp := handlerFunc(data)
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
	var equipmentToReturn []*ent.Equipment

	s.equipmentRepo.On("EquipmentsByFilter", ctx, equipmentFilter).Return(equipmentToReturn, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_FindEquipmentFunc_MapErr() {
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
	var equipmentToReturn []*ent.Equipment
	equipmentToReturn = append(equipmentToReturn, InvalidEquipment(t))

	s.equipmentRepo.On("EquipmentsByFilter", ctx, equipmentFilter).Return(equipmentToReturn, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_FindEquipmentFunc_EmptyList() {
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
	var equipmentToReturn []*ent.Equipment

	s.equipmentRepo.On("EquipmentsByFilter", ctx, equipmentFilter).Return(equipmentToReturn, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)
}

func (s *EquipmentTestSuite) TestEquipment_FindEquipmentFunc_OK() {
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
	var equipmentToReturn []*ent.Equipment
	equipmentToReturn = append(equipmentToReturn, ValidEquipment(t))

	s.equipmentRepo.On("EquipmentsByFilter", ctx, equipmentFilter).Return(equipmentToReturn, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.equipmentRepo.AssertExpectations(t)

	actualEquipment := make([]models.Equipment, len(equipmentToReturn))
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &actualEquipment)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	assert.Equal(t, equipmentToReturn[0].Name, *actualEquipment[0].Name)
}

func (s *EquipmentTestSuite) TestEquipment_EditEquipmentFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.equipment.EditEquipmentFunc(s.equipmentRepo)
	equipmentId := int64(1)
	equipmentKindUpdate := int64(10)
	equipmentUpdate := &models.Equipment{
		Kind: &equipmentKindUpdate,
	}
	data := equipment.EditEquipmentParams{
		HTTPRequest:   &request,
		EquipmentID:   equipmentId,
		EditEquipment: equipmentUpdate,
	}
	err := errors.New("test error")

	s.equipmentRepo.On("UpdateEquipmentByID", ctx, int(equipmentId), equipmentUpdate).
		Return(nil, err)

	resp := handlerFunc(data)
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
	equipmentKindUpdate := int64(10)
	equipmentUpdate := &models.Equipment{
		Kind: &equipmentKindUpdate,
	}
	data := equipment.EditEquipmentParams{
		HTTPRequest:   &request,
		EquipmentID:   equipmentId,
		EditEquipment: equipmentUpdate,
	}
	equipmentToReturn := InvalidEquipment(t)

	s.equipmentRepo.On("UpdateEquipmentByID", ctx, int(equipmentId), equipmentUpdate).
		Return(equipmentToReturn, nil)

	resp := handlerFunc(data)
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
		Kind: &equipmentKindUpdate,
	}
	data := equipment.EditEquipmentParams{
		HTTPRequest:   &request,
		EquipmentID:   equipmentId,
		EditEquipment: equipmentUpdate,
	}
	equipmentToReturn := ValidEquipment(t)

	s.equipmentRepo.On("UpdateEquipmentByID", ctx, int(equipmentId), equipmentUpdate).
		Return(equipmentToReturn, nil)

	resp := handlerFunc(data)
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
