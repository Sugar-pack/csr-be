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
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/pet_size"
)

func TestSetPetSizeHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:petsizehandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	SetPetSizeHandler(logger, api)
	require.NotEmpty(t, api.PetSizeGetAllPetSizeHandler)
	require.NotEmpty(t, api.PetSizeEditPetSizeHandler)
	require.NotEmpty(t, api.PetSizeCreateNewPetSizeHandler)
	require.NotEmpty(t, api.PetSizeDeletePetSizeHandler)
	require.NotEmpty(t, api.PetSizeGetPetSizeHandler)

}

type PetSizeTestSuite struct {
	suite.Suite
	logger      *zap.Logger
	petSizeRepo *mocks.PetSizeRepository
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
		ID:          1,
		Name:        "test pet name",
		Size:        "test size",
		IsUniversal: false,
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
	s.petSizeRepo = &mocks.PetSizeRepository{}
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

	s.petSizeRepo.On("Create", ctx, petSizeToAdd).Return(petSizeToReturn, nil)

	var petSizesToReturn []*ent.PetSize
	for i := 0; i < 10; i++ {
		ps := ValidPetSize(t)
		ps.Name = ps.Name + fmt.Sprintf("-%v", i)
		petSizesToReturn = append(petSizesToReturn, ps)
	}
	s.petSizeRepo.On("GetAll", ctx).Return(petSizesToReturn, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusCreated, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)

	actualPetSize := ent.PetSize{}
	body := responseRecorder.Body.Bytes()
	err := json.Unmarshal(body, &actualPetSize)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	eq := isEqual(t, petSizeToReturn, &actualPetSize)
	require.Equal(t, true, eq)
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

	s.petSizeRepo.On("Create", ctx, petSizeToAdd).Return(nil, err)
	var petSizesToReturn []*ent.PetSize
	for i := 0; i < 10; i++ {
		ps := ValidPetSize(t)
		ps.Name = ps.Name + fmt.Sprintf("-%v", i)
		petSizesToReturn = append(petSizesToReturn, ps)
	}
	s.petSizeRepo.On("GetAll", ctx).Return(petSizesToReturn, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}

func (s *PetSizeTestSuite) TestPetSize_CreatePetSizeFunc_ErrRespGetAll() {
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

	err := errors.New("Error while creating pet size")
	s.petSizeRepo.On("GetAll", ctx).Return(nil, err)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
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

	s.petSizeRepo.On("Create", ctx, petSizeToAdd).Return(toReturn, nil)
	var petSizesToReturn []*ent.PetSize
	for i := 0; i < 10; i++ {
		ps := ValidPetSize(t)
		ps.Name = ps.Name + fmt.Sprintf("-%v", i)
		petSizesToReturn = append(petSizesToReturn, ps)
	}
	s.petSizeRepo.On("GetAll", ctx).Return(petSizesToReturn, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
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
	s.petSizeRepo.On("GetAll", ctx).Return(petSizeToReturn, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)

	actualPetSize := []*ent.PetSize{}
	body := responseRecorder.Body.Bytes()
	err := json.Unmarshal(body, &actualPetSize)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	for i := 0; i < 10; i++ {
		eq := isEqual(t, petSizeToReturn[i], actualPetSize[i])
		require.Equal(t, true, eq)
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

	s.petSizeRepo.On("GetAll", ctx).Return(nil, err)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
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
	s.petSizeRepo.On("GetAll", ctx).Return(toReturn, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
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

	s.petSizeRepo.On("Delete", ctx, idToDelete).Return(err)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
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
	s.petSizeRepo.On("Delete", ctx, idToDelete).Return(nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
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
	s.petSizeRepo.On("GetByID", ctx, idToGet).Return(petSizeToReturn, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)

	actualPetSize := ent.PetSize{}
	body := responseRecorder.Body.Bytes()
	err := json.Unmarshal(body, &actualPetSize)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	eq := isEqual(t, petSizeToReturn, &actualPetSize)
	require.Equal(t, true, eq)
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
	s.petSizeRepo.On("GetByID", ctx, idToGet).Return(nil, err)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
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
	s.petSizeRepo.On("GetByID", ctx, idToGet).Return(nil, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}

func (s *PetSizeTestSuite) TestPetSize_EditPetSizeFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet name"
	size := "test size"
	id := 0
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

	s.petSizeRepo.On("Update", ctx, id, petSizeToUpdate).Return(petSizeToReturn, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)

	actualPetSize := ent.PetSize{}
	body := responseRecorder.Body.Bytes()
	err := json.Unmarshal(body, &actualPetSize)
	if err != nil {
		t.Errorf("unable to unmarshal response body: %v", err)
	}
	eq := isEqual(t, petSizeToReturn, &actualPetSize)
	require.Equal(t, true, eq)
}

func (s *PetSizeTestSuite) TestPetSize_EditPetSizeFunc_ErrRespNil() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet name"
	size := "test size"
	id := 0
	handlerFunc := s.petSize.UpdatePetSizeByID(s.petSizeRepo)
	petSizeToUpdate := &models.PetSize{
		Name: &name,
		Size: &size,
	}
	data := pet_size.EditPetSizeParams{
		HTTPRequest: &request,
		EditPetSize: petSizeToUpdate,
	}
	s.petSizeRepo.On("Update", ctx, id, petSizeToUpdate).Return(nil, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}

func (s *PetSizeTestSuite) TestPetSize_EditPetSizeFunc_ErrFromRepo() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	name := "test pet name"
	size := "test size"
	handlerFunc := s.petSize.UpdatePetSizeByID(s.petSizeRepo)
	id := 0
	petSizeToUpdate := &models.PetSize{
		Name: &name,
		Size: &size,
	}
	data := pet_size.EditPetSizeParams{
		HTTPRequest: &request,
		EditPetSize: petSizeToUpdate,
	}
	err := errors.New("test")
	s.petSizeRepo.On("Update", ctx, id, petSizeToUpdate).Return(nil, err)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.petSizeRepo.AssertExpectations(t)
}
