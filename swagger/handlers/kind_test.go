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
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/kinds"
)

type KindTestSuite struct {
	suite.Suite
	logger     *zap.Logger
	repository *repomock.KindRepository
	handler    *Kind
}

func TestKindSuite(t *testing.T) {
	suite.Run(t, new(KindTestSuite))
}

func (s *KindTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.repository = &repomock.KindRepository{}
	s.handler = NewKind(s.logger)
}

func (s *KindTestSuite) TestKind_CreateKind_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	maxReservationTime := int64(100)
	maxReservationUnit := int64(1)
	kindName := "test"
	newKind := models.CreateNewKind{
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnit,
		Name:                &kindName,
	}
	data := kinds.CreateNewKindParams{
		HTTPRequest: &request,
		NewKind:     &newKind,
	}
	err := errors.New("test")
	s.repository.On("CreateKind", ctx, newKind).Return(nil, err)

	handlerFunc := s.handler.CreateNewKindFunc(s.repository)
	resp := handlerFunc.Handle(data)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_CreateKind_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	maxReservationTime := int64(100)
	maxReservationUnit := int64(1)
	kindName := "test"
	newKind := models.CreateNewKind{
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnit,
		Name:                &kindName,
	}
	data := kinds.CreateNewKindParams{
		HTTPRequest: &request,
		NewKind:     &newKind,
	}

	kindToReturn := &ent.Kind{
		ID:                  1,
		Name:                kindName,
		MaxReservationTime:  maxReservationTime,
		MaxReservationUnits: maxReservationUnit,
	}
	s.repository.On("CreateKind", ctx, newKind).Return(kindToReturn, nil)

	handlerFunc := s.handler.CreateNewKindFunc(s.repository)
	resp := handlerFunc.Handle(data)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	returnedKind := models.CreateNewKindResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedKind)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, kindToReturn.ID, int(returnedKind.Data.ID))
	assert.Equal(t, kindToReturn.Name, *returnedKind.Data.Name)
	assert.Equal(t, kindToReturn.MaxReservationTime, returnedKind.Data.MaxReservationTime)
	assert.Equal(t, kindToReturn.MaxReservationUnits, returnedKind.Data.MaxReservationUnits)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetAllKinds_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := kinds.GetAllKindsParams{
		HTTPRequest: &request,
	}

	err := errors.New("test")
	s.repository.On("AllKind", ctx).Return(nil, err)

	handlerFunc := s.handler.GetAllKindsFunc(s.repository)
	resp := handlerFunc.Handle(data)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetAllKinds_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := kinds.GetAllKindsParams{
		HTTPRequest: &request,
	}

	var kindsToReturn []*ent.Kind
	kindToReturn := &ent.Kind{
		ID:                  1,
		Name:                "test",
		MaxReservationTime:  100,
		MaxReservationUnits: 1,
	}
	kindsToReturn = append(kindsToReturn, kindToReturn)
	s.repository.On("AllKind", ctx).Return(kindsToReturn, nil)

	handlerFunc := s.handler.GetAllKindsFunc(s.repository)
	resp := handlerFunc.Handle(data)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var returnedKinds []models.Kind
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedKinds)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(kindsToReturn), len(returnedKinds))
	assert.Equal(t, kindToReturn.ID, int(returnedKinds[0].ID))
	assert.Equal(t, kindToReturn.Name, *returnedKinds[0].Name)
	assert.Equal(t, kindToReturn.MaxReservationTime, returnedKinds[0].MaxReservationTime)
	assert.Equal(t, kindToReturn.MaxReservationUnits, returnedKinds[0].MaxReservationUnits)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetKindByID_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := kinds.GetKindByIDParams{
		HTTPRequest: &request,
		KindID:      1,
	}

	err := errors.New("test")
	s.repository.On("KindByID", ctx, int(data.KindID)).Return(nil, err)

	handlerFunc := s.handler.GetKindByIDFunc(s.repository)
	resp := handlerFunc.Handle(data)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_GetKindByID_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := kinds.GetKindByIDParams{
		HTTPRequest: &request,
		KindID:      1,
	}

	kindToReturn := &ent.Kind{
		ID:                  1,
		Name:                "test",
		MaxReservationTime:  100,
		MaxReservationUnits: 1,
	}
	s.repository.On("KindByID", ctx, int(data.KindID)).Return(kindToReturn, nil)

	handlerFunc := s.handler.GetKindByIDFunc(s.repository)
	resp := handlerFunc.Handle(data)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	returnedKind := models.GetKindByIDResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedKind)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, kindToReturn.ID, int(returnedKind.Data.ID))
	assert.Equal(t, kindToReturn.Name, *returnedKind.Data.Name)
	assert.Equal(t, kindToReturn.MaxReservationTime, returnedKind.Data.MaxReservationTime)
	assert.Equal(t, kindToReturn.MaxReservationUnits, returnedKind.Data.MaxReservationUnits)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_DeleteKind_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := kinds.DeleteKindParams{
		HTTPRequest: &request,
		KindID:      1,
	}

	err := errors.New("test")
	s.repository.On("DeleteKindByID", ctx, int(data.KindID)).Return(err)

	handlerFunc := s.handler.DeleteKindFunc(s.repository)
	resp := handlerFunc.Handle(data)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_DeleteKind_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := kinds.DeleteKindParams{
		HTTPRequest: &request,
		KindID:      1,
	}

	s.repository.On("DeleteKindByID", ctx, int(data.KindID)).Return(nil)

	handlerFunc := s.handler.DeleteKindFunc(s.repository)
	resp := handlerFunc.Handle(data)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_PatchKind_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	patch := models.PatchKind{
		Name: "test",
	}
	data := kinds.PatchKindParams{
		HTTPRequest: &request,
		KindID:      1,
		PatchKind:   &patch,
	}

	err := errors.New("test")
	s.repository.On("UpdateKind", ctx, int(data.KindID), patch).Return(nil, err)

	handlerFunc := s.handler.PatchKindFunc(s.repository)
	resp := handlerFunc.Handle(data)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.repository.AssertExpectations(t)
}

func (s *KindTestSuite) TestKind_PatchKind_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	patch := models.PatchKind{
		Name: "test",
	}
	data := kinds.PatchKindParams{
		HTTPRequest: &request,
		KindID:      1,
		PatchKind:   &patch,
	}

	updatedKind := &ent.Kind{
		ID:                  1,
		Name:                patch.Name,
		MaxReservationTime:  100,
		MaxReservationUnits: 1,
	}
	s.repository.On("UpdateKind", ctx, int(data.KindID), patch).Return(updatedKind, nil)

	handlerFunc := s.handler.PatchKindFunc(s.repository)
	resp := handlerFunc.Handle(data)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	returnedKind := models.PatchKindResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &returnedKind)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, updatedKind.ID, int(returnedKind.Data.ID))
	assert.Equal(t, updatedKind.Name, *returnedKind.Data.Name)
	assert.Equal(t, updatedKind.MaxReservationTime, returnedKind.Data.MaxReservationTime)
	assert.Equal(t, updatedKind.MaxReservationUnits, returnedKind.Data.MaxReservationUnits)

	s.repository.AssertExpectations(t)
}
