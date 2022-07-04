package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/suite"
	"gotest.tools/assert"

	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/pet_size"
)

type PetSizeTestSuite struct {
	suite.Suite
	logger      *zap.Logger
	petSizeRepo *repomock.PetSizeRepository
	petSize     *PetSize
}

func InvalidPetSize(t *testing.T) *ent.PetSize {
	t.Helper()
	return &ent.PetSize{
		ID:   1,
		Name: "no edges",
	}
}

func ValidPetSize(t *testing.T) *ent.PetSize {
	t.Helper()
	return &ent.PetSize{
		ID:   1,
		Name: "test pet name",
		Size: "test size",
		Edges: ent.PetSizeEdges{
			Equipments: []*ent.Equipment{},
		},
	}
}

func isEqual(t *testing.T, first *ent.PetSize, second *ent.PetSize) bool {
	t.Helper()
	if first.ID == second.ID && first.Name == second.Name && first.Size == second.Size {
		return true
	}
	return false
}

func TestPetSizeSuite(t *testing.T) {
	suite.Run(t, new(PetSizeTestSuite))
}

func (s *PetSizeTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.petSizeRepo = &repomock.PetSizeRepository{}
	s.petSize = NewPetSize(s.logger)
}

func (s *PetSizeTestSuite) TestPetSize_CreatePetSizeFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet name"
	size := "test size"
	handlerFunc := s.petSize.CreatePetSizeFunc(s.petSizeRepo)
	petSizeToAdd := models.PetSize{
		Name: &name,
		Size: &size,
	}
	data := pet_size.CreateNewPetSizeParams{
		HTTPRequest: &request,
		NewPetSize:  &petSizeToAdd,
	}

	petSizeToReturn := ValidPetSize(t)

	s.petSizeRepo.On("CreatePetSize", ctx, petSizeToAdd).Return(petSizeToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)

	actualPetSize := ent.PetSize{}
	body := responseRecorder.Body.Bytes()
	err := json.Unmarshal(body, &actualPetSize)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	eq := isEqual(t, petSizeToReturn, &actualPetSize)
	assert.Equal(t, true, eq)
}

func (s *PetSizeTestSuite) TestPetSize_CreatePetSizeFunc_ErrFromRepo() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet name"
	size := "test size"
	handlerFunc := s.petSize.CreatePetSizeFunc(s.petSizeRepo)

	petSizeToAdd := models.PetSize{
		Name: &name,
		Size: &size,
	}
	data := pet_size.CreateNewPetSizeParams{
		HTTPRequest: &request,
		NewPetSize:  &petSizeToAdd,
	}
	err := errors.New("test")

	s.petSizeRepo.On("CreatePetSize", ctx, petSizeToAdd).Return(nil, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}

func (s *PetSizeTestSuite) TestPetSize_CreatePetSizeFunc_ErrRespNil() {
	t := s.T()
	var toReturn *ent.PetSize
	request := http.Request{}
	ctx := request.Context()
	name := "test pet name"
	size := "test size"
	handlerFunc := s.petSize.CreatePetSizeFunc(s.petSizeRepo)

	petSizeToAdd := models.PetSize{
		Name: &name,
		Size: &size,
	}
	data := pet_size.CreateNewPetSizeParams{
		HTTPRequest: &request,
		NewPetSize:  &petSizeToAdd,
	}

	s.petSizeRepo.On("CreatePetSize", ctx, petSizeToAdd).Return(toReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}

func (s *PetSizeTestSuite) TestPetSize_GetAllPetSizeFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	handlerFunc := s.petSize.GetAllPetSizeFunc(s.petSizeRepo)
	data := pet_size.GetAllPetSizeParams{
		HTTPRequest: &request,
	}
	var petSizeToReturn []*ent.PetSize
	for i := 0; i < 10; i++ {
		petSizeToReturn = append(petSizeToReturn, ValidPetSize(t))
	}
	s.petSizeRepo.On("AllPetSizes", ctx).Return(petSizeToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)

	actualPetSize := []*ent.PetSize{}
	body := responseRecorder.Body.Bytes()
	err := json.Unmarshal(body, &actualPetSize)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	for i := 0; i < 10; i++ {
		eq := isEqual(t, petSizeToReturn[i], actualPetSize[i])
		assert.Equal(t, true, eq)
	}
}

func (s *PetSizeTestSuite) TestPetSize_GetAllPetSizeFunc_ErrFromRepo() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	handlerFunc := s.petSize.GetAllPetSizeFunc(s.petSizeRepo)
	data := pet_size.GetAllPetSizeParams{
		HTTPRequest: &request,
	}
	err := errors.New("test")

	s.petSizeRepo.On("AllPetSizes", ctx).Return(nil, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}

func (s *PetSizeTestSuite) TestPetSize_GetAllPetSizeFunc_ErrRespNil() {
	t := s.T()
	var toReturn []*ent.PetSize
	request := http.Request{}
	ctx := request.Context()
	handlerFunc := s.petSize.GetAllPetSizeFunc(s.petSizeRepo)
	data := pet_size.GetAllPetSizeParams{
		HTTPRequest: &request,
	}
	s.petSizeRepo.On("AllPetSizes", ctx).Return(toReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}

func (s *PetSizeTestSuite) TestPetSize_DeletePetSizeFunc_Err() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	idToDelete := 1
	handlerFunc := s.petSize.DeletePetSizeByID(s.petSizeRepo)
	data := pet_size.DeletePetSizeParams{
		HTTPRequest: &request,
		PetSizeID:   int64(idToDelete),
	}
	err := errors.New("test")

	s.petSizeRepo.On("DeletePetSizeByID", ctx, idToDelete).Return(err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}

func (s *PetSizeTestSuite) TestPetSize_DeletePetSizeFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	idToDelete := 1
	handlerFunc := s.petSize.DeletePetSizeByID(s.petSizeRepo)
	data := pet_size.DeletePetSizeParams{
		HTTPRequest: &request,
		PetSizeID:   int64(idToDelete),
	}
	s.petSizeRepo.On("DeletePetSizeByID", ctx, idToDelete).Return(nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}

func (s *PetSizeTestSuite) TestPetSize_GetPetSizeByIDFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	idToGet := 1
	handlerFunc := s.petSize.GetPetSizeByID(s.petSizeRepo)
	data := pet_size.GetPetSizeParams{
		HTTPRequest: &request,
		PetSizeID:   int64(idToGet),
	}
	petSizeToReturn := ValidPetSize(t)
	s.petSizeRepo.On("PetSizeByID", ctx, idToGet).Return(petSizeToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)

	actualPetSize := ent.PetSize{}
	body := responseRecorder.Body.Bytes()
	err := json.Unmarshal(body, &actualPetSize)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	eq := isEqual(t, petSizeToReturn, &actualPetSize)
	assert.Equal(t, true, eq)
}

func (s *PetSizeTestSuite) TestPetSize_GetPetSizeByIDFunc_ErrFromRepo() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	idToGet := 1
	handlerFunc := s.petSize.GetPetSizeByID(s.petSizeRepo)
	data := pet_size.GetPetSizeParams{
		HTTPRequest: &request,
		PetSizeID:   int64(idToGet),
	}

	err := errors.New("test")
	s.petSizeRepo.On("PetSizeByID", ctx, idToGet).Return(nil, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}

func (s *PetSizeTestSuite) TestPetSize_GetPetSizeByIDFunc_ErrRespNil() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	idToGet := 1
	handlerFunc := s.petSize.GetPetSizeByID(s.petSizeRepo)
	data := pet_size.GetPetSizeParams{
		HTTPRequest: &request,
		PetSizeID:   int64(idToGet),
	}
	s.petSizeRepo.On("PetSizeByID", ctx, idToGet).Return(nil, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}

func (s *PetSizeTestSuite) TestPetSize_EditPetSizeFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet name"
	size := "test size"
	handlerFunc := s.petSize.UpdatePetSizeByID(s.petSizeRepo)
	petSizeToUpdate := &models.PetSize{
		Name: &name,
		Size: &size,
	}
	data := pet_size.EditPetSizeParams{
		HTTPRequest: &request,
		EditPetSize: petSizeToUpdate,
	}

	petSizeToReturn := ValidPetSize(t)

	s.petSizeRepo.On("UpdatePetSizeByID", ctx, int(petSizeToUpdate.ID), petSizeToUpdate).Return(petSizeToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)

	actualPetSize := ent.PetSize{}
	body := responseRecorder.Body.Bytes()
	err := json.Unmarshal(body, &actualPetSize)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	eq := isEqual(t, petSizeToReturn, &actualPetSize)
	assert.Equal(t, true, eq)
}

func (s *PetSizeTestSuite) TestPetSize_EditPetSizeFunc_ErrRespNil() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet name"
	size := "test size"
	handlerFunc := s.petSize.UpdatePetSizeByID(s.petSizeRepo)
	petSizeToUpdate := &models.PetSize{
		Name: &name,
		Size: &size,
	}
	data := pet_size.EditPetSizeParams{
		HTTPRequest: &request,
		EditPetSize: petSizeToUpdate,
	}
	s.petSizeRepo.On("UpdatePetSizeByID", ctx, int(petSizeToUpdate.ID), petSizeToUpdate).Return(nil, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}

func (s *PetSizeTestSuite) TestPetSize_EditPetSizeFunc_ErrFromRepo() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet name"
	size := "test size"
	handlerFunc := s.petSize.UpdatePetSizeByID(s.petSizeRepo)
	petSizeToUpdate := &models.PetSize{
		Name: &name,
		Size: &size,
	}
	data := pet_size.EditPetSizeParams{
		HTTPRequest: &request,
		EditPetSize: petSizeToUpdate,
	}
	err := errors.New("test")
	s.petSizeRepo.On("UpdatePetSizeByID", ctx, int(petSizeToUpdate.ID), petSizeToUpdate).Return(nil, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}
