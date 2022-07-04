package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/status"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
)

type StatusTestSuite struct {
	suite.Suite
	logger     *zap.Logger
	repository *repomock.EquipmentStatusRepository
	handler    *Status
}

func TestStatusSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}

func (s *StatusTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.repository = &repomock.EquipmentStatusRepository{}
	s.handler = NewStatus(s.logger)
}

func (s *StatusTestSuite) TestStatus_PostStatus_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusName := "statusName"
	data := status.PostStatusParams{
		HTTPRequest: &request,
		Name: &models.StatusName{
			Name: &statusName,
		},
	}
	err := errors.New("test")
	s.repository.On("Create", ctx, statusName).Return(nil, err)

	handlerFunc := s.handler.PostStatusFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *StatusTestSuite) TestStatus_PostStatus_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusName := "statusName"
	data := status.PostStatusParams{
		HTTPRequest: &request,
		Name: &models.StatusName{
			Name: &statusName,
		},
	}
	statusToReturn := &ent.Statuses{
		ID: 1,
	}
	s.repository.On("Create", ctx, statusName).Return(statusToReturn, nil)

	handlerFunc := s.handler.PostStatusFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseStatus := models.SuccessStatusOperationResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseStatus)
	if err != nil {
		t.Errorf("Error unmarshalling response: %v", err)
	}
	assert.Equal(t, statusToReturn.ID, int(*responseStatus.Data.ID))

	s.repository.AssertExpectations(t)
}

func (s *StatusTestSuite) TestStatus_GetStatuses_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := status.GetStatusesParams{
		HTTPRequest: &request,
	}
	err := errors.New("test")
	s.repository.On("GetAll", ctx).Return(nil, err)

	handlerFunc := s.handler.GetStatusesFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *StatusTestSuite) TestStatus_GetStatuses_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := status.GetStatusesParams{
		HTTPRequest: &request,
	}
	var statusesToReturn []*ent.Statuses
	statusToReturn := &ent.Statuses{
		ID: 1,
	}
	statusesToReturn = append(statusesToReturn, statusToReturn)
	s.repository.On("GetAll", ctx).Return(statusesToReturn, nil)

	handlerFunc := s.handler.GetStatusesFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var responseStatus []models.Status
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseStatus)
	if err != nil {
		t.Errorf("Error unmarshalling response: %v", err)
	}
	assert.Equal(t, len(statusesToReturn), len(responseStatus))
	assert.Equal(t, statusToReturn.ID, int(*responseStatus[0].ID))

	s.repository.AssertExpectations(t)
}

func (s *StatusTestSuite) TestStatus_GetStatus_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusID := 1
	data := status.GetStatusParams{
		HTTPRequest: &request,
		StatusID:    int64(statusID),
	}
	err := errors.New("test")
	s.repository.On("Get", ctx, statusID).Return(nil, err)

	handlerFunc := s.handler.GetStatusFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *StatusTestSuite) TestStatus_GetStatus_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusName := "statusName"
	data := status.PostStatusParams{
		HTTPRequest: &request,
		Name: &models.StatusName{
			Name: &statusName,
		},
	}
	statusToReturn := &ent.Statuses{
		ID: 1,
	}
	s.repository.On("Create", ctx, statusName).Return(statusToReturn, nil)

	handlerFunc := s.handler.PostStatusFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseStatus := models.SuccessStatusOperationResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseStatus)
	if err != nil {
		t.Errorf("Error unmarshalling response: %v", err)
	}
	assert.Equal(t, statusToReturn.ID, int(*responseStatus.Data.ID))

	s.repository.AssertExpectations(t)
}

func (s *StatusTestSuite) TestStatus_DeleteStatus_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusID := 1
	data := status.DeleteStatusParams{
		HTTPRequest: &request,
		StatusID:    int64(statusID),
	}
	err := errors.New("test")
	s.repository.On("Delete", ctx, statusID).Return(nil, err)

	handlerFunc := s.handler.DeleteStatusFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *StatusTestSuite) TestStatus_DeleteStatus_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusID := 1
	data := status.DeleteStatusParams{
		HTTPRequest: &request,
		StatusID:    int64(statusID),
	}
	statusToReturn := &ent.Statuses{
		ID: 1,
	}
	s.repository.On("Delete", ctx, statusID).Return(statusToReturn, nil)

	handlerFunc := s.handler.DeleteStatusFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseStatus := models.SuccessStatusOperationResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseStatus)
	if err != nil {
		t.Errorf("Error unmarshalling response: %v", err)
	}
	assert.Equal(t, statusToReturn.ID, int(*responseStatus.Data.ID))

	s.repository.AssertExpectations(t)
}
