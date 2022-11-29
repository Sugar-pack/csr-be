package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/pet_kind"
)

func TestSetPetKindHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:petkindhandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	SetPetKindHandler(logger, api)
	assert.NotEmpty(t, api.PetKindGetAllPetKindsHandler)
	assert.NotEmpty(t, api.PetKindEditPetKindHandler)
	assert.NotEmpty(t, api.PetKindDeletePetKindHandler)
	assert.NotEmpty(t, api.PetKindCreateNewPetKindHandler)
	assert.NotEmpty(t, api.PetKindGetPetKindHandler)
}

type PetKindTestSuite struct {
	suite.Suite
	logger      *zap.Logger
	petKindRepo *mocks.PetKindRepository
	petKind     *PetKind
}

func InvalidPetKind(t *testing.T) *ent.PetKind {
	t.Helper()
	return &ent.PetKind{
		ID:   1,
		Name: "no edges",
	}
}

func ValidPetKind(t *testing.T) *ent.PetKind {
	t.Helper()
	return &ent.PetKind{
		ID:   1,
		Name: "test pet kind name",
		Edges: ent.PetKindEdges{
			Equipments: []*ent.Equipment{},
		},
	}
}

func isEqualPetKind(t *testing.T, first *ent.PetKind, second *ent.PetKind) bool {
	t.Helper()
	if first.ID == second.ID && first.Name == second.Name {
		return true
	}
	return false
}

func TestPetKindSuite(t *testing.T) {
	suite.Run(t, new(PetKindTestSuite))
}

func (s *PetKindTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.petKindRepo = &mocks.PetKindRepository{}
	s.petKind = NewPetKind(s.logger)
}

func (s *PetKindTestSuite) TestPetKind_CreatePetKindFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet kind name"
	handlerFunc := s.petKind.CreatePetKindFunc(s.petKindRepo)
	petKindToAdd := models.PetKind{
		Name: &name,
	}
	data := pet_kind.CreateNewPetKindParams{
		HTTPRequest: &request,
		NewPetKind:  &petKindToAdd,
	}

	petKindToReturn := ValidPetKind(t)

	s.petKindRepo.On("Create", ctx, petKindToAdd).Return(petKindToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)

	actualPetKind := ent.PetKind{}
	body := responseRecorder.Body.Bytes()
	err := json.Unmarshal(body, &actualPetKind)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	eq := isEqualPetKind(t, petKindToReturn, &actualPetKind)
	assert.Equal(t, true, eq)
}

func (s *PetKindTestSuite) TestPetKind_CreatePetKindFunc_ErrFromRepo() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet kind name"
	handlerFunc := s.petKind.CreatePetKindFunc(s.petKindRepo)

	petKindToAdd := models.PetKind{
		Name: &name,
	}
	data := pet_kind.CreateNewPetKindParams{
		HTTPRequest: &request,
		NewPetKind:  &petKindToAdd,
	}
	err := errors.New("test")

	s.petKindRepo.On("Create", ctx, petKindToAdd).Return(nil, err)

	access := "dummy access"
	resp := handlerFunc.Handle(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)
}

func (s *PetKindTestSuite) TestPetKind_CreatePetKindFunc_ErrRespNil() {
	t := s.T()
	var toReturn *ent.PetKind
	request := http.Request{}
	ctx := request.Context()
	name := "test pet kind name"
	handlerFunc := s.petKind.CreatePetKindFunc(s.petKindRepo)

	petKindToAdd := models.PetKind{
		Name: &name,
	}
	data := pet_kind.CreateNewPetKindParams{
		HTTPRequest: &request,
		NewPetKind:  &petKindToAdd,
	}

	s.petKindRepo.On("Create", ctx, petKindToAdd).Return(toReturn, nil)

	access := "dummy access"
	resp := handlerFunc.Handle(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)
}

func (s *PetKindTestSuite) TestPetKind_GetAllPetKindFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	handlerFunc := s.petKind.GetAllPetKindFunc(s.petKindRepo)
	data := pet_kind.GetAllPetKindsParams{
		HTTPRequest: &request,
	}
	var petKindToReturn []*ent.PetKind
	for i := 0; i < 10; i++ {
		petKindToReturn = append(petKindToReturn, ValidPetKind(t))
	}
	s.petKindRepo.On("GetAll", ctx).Return(petKindToReturn, nil)

	access := "dummy access"
	resp := handlerFunc.Handle(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)

	actualPetKind := []*ent.PetKind{}
	body := responseRecorder.Body.Bytes()
	err := json.Unmarshal(body, &actualPetKind)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	for i := 0; i < 10; i++ {
		eq := isEqualPetKind(t, petKindToReturn[i], actualPetKind[i])
		assert.Equal(t, true, eq)
	}
}

func (s *PetKindTestSuite) TestPetKind_GetAllPetKindFunc_ErrFromRepo() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	handlerFunc := s.petKind.GetAllPetKindFunc(s.petKindRepo)
	data := pet_kind.GetAllPetKindsParams{
		HTTPRequest: &request,
	}
	err := errors.New("test")

	s.petKindRepo.On("GetAll", ctx).Return(nil, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)
}

func (s *PetKindTestSuite) TestPetKind_GetAllPetKindFunc_ErrRespNil() {
	t := s.T()
	var toReturn []*ent.PetKind
	request := http.Request{}
	ctx := request.Context()
	handlerFunc := s.petKind.GetAllPetKindFunc(s.petKindRepo)
	data := pet_kind.GetAllPetKindsParams{
		HTTPRequest: &request,
	}
	s.petKindRepo.On("GetAll", ctx).Return(toReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)
}

func (s *PetKindTestSuite) TestPetKind_DeletePetKindFunc_Err() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	idToDelete := 1
	handlerFunc := s.petKind.DeletePetKindByID(s.petKindRepo)
	data := pet_kind.DeletePetKindParams{
		HTTPRequest: &request,
		PetKindID:   int64(idToDelete),
	}
	err := errors.New("test")

	s.petKindRepo.On("Delete", ctx, idToDelete).Return(err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)
}

func (s *PetKindTestSuite) TestPetKind_DeletePetKindFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	idToDelete := 1
	handlerFunc := s.petKind.DeletePetKindByID(s.petKindRepo)
	data := pet_kind.DeletePetKindParams{
		HTTPRequest: &request,
		PetKindID:   int64(idToDelete),
	}
	s.petKindRepo.On("Delete", ctx, idToDelete).Return(nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)
}

func (s *PetKindTestSuite) TestPetKind_GetPetKindByIDFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	idToGet := 1
	handlerFunc := s.petKind.GetPetKindsByID(s.petKindRepo)
	data := pet_kind.GetPetKindParams{
		HTTPRequest: &request,
		PetKindID:   int64(idToGet),
	}
	petKindToReturn := ValidPetKind(t)
	s.petKindRepo.On("GetByID", ctx, idToGet).Return(petKindToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)

	actualPetKind := ent.PetKind{}
	body := responseRecorder.Body.Bytes()
	err := json.Unmarshal(body, &actualPetKind)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	eq := isEqualPetKind(t, petKindToReturn, &actualPetKind)
	assert.Equal(t, true, eq)
}

func (s *PetKindTestSuite) TestPetKind_GetPetKindByIDFunc_ErrFromRepo() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	idToGet := 1
	handlerFunc := s.petKind.GetPetKindsByID(s.petKindRepo)
	data := pet_kind.GetPetKindParams{
		HTTPRequest: &request,
		PetKindID:   int64(idToGet),
	}

	err := errors.New("test")
	s.petKindRepo.On("GetByID", ctx, idToGet).Return(nil, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)
}

func (s *PetKindTestSuite) TestPetKind_GetPetKindByIDFunc_ErrRespNil() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	idToGet := 1
	handlerFunc := s.petKind.GetPetKindsByID(s.petKindRepo)
	data := pet_kind.GetPetKindParams{
		HTTPRequest: &request,
		PetKindID:   int64(idToGet),
	}
	s.petKindRepo.On("GetByID", ctx, idToGet).Return(nil, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)
}

func (s *PetKindTestSuite) TestPetKind_EditPetKindFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet kind name"
	handlerFunc := s.petKind.UpdatePetKindByID(s.petKindRepo)
	petKindToUpdate := &models.PetKind{
		Name: &name,
	}
	data := pet_kind.EditPetKindParams{
		HTTPRequest: &request,
		EditPetKind: petKindToUpdate,
	}

	petKindToReturn := ValidPetKind(t)

	s.petKindRepo.On("Update", ctx, int(petKindToUpdate.ID), petKindToUpdate).Return(petKindToReturn, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)

	actualPetKind := ent.PetKind{}
	body := responseRecorder.Body.Bytes()
	err := json.Unmarshal(body, &actualPetKind)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	eq := isEqualPetKind(t, petKindToReturn, &actualPetKind)
	assert.Equal(t, true, eq)
}

func (s *PetKindTestSuite) TestPetKind_EditPetKindFunc_ErrRespNil() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet kind name"
	handlerFunc := s.petKind.UpdatePetKindByID(s.petKindRepo)
	petKindToUpdate := &models.PetKind{
		Name: &name,
	}
	data := pet_kind.EditPetKindParams{
		HTTPRequest: &request,
		EditPetKind: petKindToUpdate,
	}
	s.petKindRepo.On("Update", ctx, int(petKindToUpdate.ID), petKindToUpdate).Return(nil, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)
}

func (s *PetKindTestSuite) TestPetKind_EditPetKindFunc_ErrFromRepo() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet kind name"
	handlerFunc := s.petKind.UpdatePetKindByID(s.petKindRepo)
	petKindToUpdate := &models.PetKind{
		Name: &name,
	}
	data := pet_kind.EditPetKindParams{
		HTTPRequest: &request,
		EditPetKind: petKindToUpdate,
	}
	err := errors.New("test")
	s.petKindRepo.On("Update", ctx, int(petKindToUpdate.ID), petKindToUpdate).Return(nil, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petKindRepo.AssertExpectations(t)
}
