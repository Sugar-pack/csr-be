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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	eqStatusName "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/equipment_status_name"
)

func TestSetEquipmentStatusNameHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:eqstatushandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	SetEquipmentStatusNameHandler(logger, api)
	require.NotEmpty(t, api.EquipmentStatusNamePostEquipmentStatusNameHandler)
	require.NotEmpty(t, api.EquipmentStatusNameGetEquipmentStatusNameHandler)
	require.NotEmpty(t, api.EquipmentStatusNameGetEquipmentStatusNameHandler)
	require.NotEmpty(t, api.EquipmentStatusNameDeleteEquipmentStatusNameHandler)
}

type EquipmentStatusNameTestSuite struct {
	suite.Suite
	logger     *zap.Logger
	repository *mocks.EquipmentStatusNameRepository
	handler    *EquipmentStatusName
}

func ValidStatus(t *testing.T) *ent.EquipmentStatusName {
	t.Helper()
	return &ent.EquipmentStatusName{
		ID:   1,
		Name: "available",
	}

}

func TestStatusNameSuite(t *testing.T) {
	suite.Run(t, new(EquipmentStatusNameTestSuite))
}

func (s *EquipmentStatusNameTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.repository = &mocks.EquipmentStatusNameRepository{}
	s.handler = NewEquipmentStatusName(s.logger)
}

func (s *EquipmentStatusNameTestSuite) TestStatus_PostStatus_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusName := "statusName"
	data := eqStatusName.PostEquipmentStatusNameParams{
		HTTPRequest: &request,
		Name: &models.NewEquipmentStatusName{
			Name: &statusName,
		},
	}
	err := errors.New("test")
	s.repository.On("Create", ctx, statusName).Return(nil, err)

	handlerFunc := s.handler.PostEquipmentStatusNameFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *EquipmentStatusNameTestSuite) TestStatus_PostStatus_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusName := "statusName"
	data := eqStatusName.PostEquipmentStatusNameParams{
		HTTPRequest: &request,
		Name: &models.NewEquipmentStatusName{
			Name: &statusName,
		},
	}
	statusToReturn := &ent.EquipmentStatusName{
		ID: 1,
	}
	s.repository.On("Create", ctx, statusName).Return(statusToReturn, nil)

	handlerFunc := s.handler.PostEquipmentStatusNameFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseStatus := models.SuccessEquipmentStatusNameOperationResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseStatus)
	if err != nil {
		t.Errorf("Error unmarshalling response: %v", err)
	}
	require.Equal(t, statusToReturn.ID, int(responseStatus.Data.ID))

	s.repository.AssertExpectations(t)
}

func (s *EquipmentStatusNameTestSuite) TestStatus_ListEquipmentStatusNames_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := eqStatusName.ListEquipmentStatusNamesParams{
		HTTPRequest: &request,
	}
	err := errors.New("test")
	s.repository.On("GetAll", ctx).Return(nil, err)

	handlerFunc := s.handler.ListEquipmentStatusNamesFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *EquipmentStatusNameTestSuite) TestStatus_ListEquipmentStatusNames_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := eqStatusName.ListEquipmentStatusNamesParams{
		HTTPRequest: &request,
	}
	var statusesToReturn []*ent.EquipmentStatusName
	statusToReturn := &ent.EquipmentStatusName{
		ID: 1,
	}
	statusesToReturn = append(statusesToReturn, statusToReturn)
	s.repository.On("GetAll", ctx).Return(statusesToReturn, nil)

	handlerFunc := s.handler.ListEquipmentStatusNamesFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var responseStatus []models.EquipmentStatusName
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseStatus)
	if err != nil {
		t.Errorf("Error unmarshalling response: %v", err)
	}
	require.Equal(t, len(statusesToReturn), len(responseStatus))
	require.Equal(t, statusToReturn.ID, int(responseStatus[0].ID))

	s.repository.AssertExpectations(t)
}

func (s *EquipmentStatusNameTestSuite) TestStatus_GetEquipmentStatusName_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusID := 1
	data := eqStatusName.GetEquipmentStatusNameParams{
		HTTPRequest: &request,
		StatusID:    int64(statusID),
	}
	err := errors.New("test")
	s.repository.On("Get", ctx, statusID).Return(nil, err)

	handlerFunc := s.handler.GetEquipmentStatusNameFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *EquipmentStatusNameTestSuite) TestStatus_GetEquipmentStatusName_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusID := 1
	data := eqStatusName.GetEquipmentStatusNameParams{
		HTTPRequest: &request,
		StatusID:    int64(statusID),
	}
	statusToReturn := &ent.EquipmentStatusName{
		ID: 1,
	}
	s.repository.On("Get", ctx, statusID).Return(statusToReturn, nil)

	handlerFunc := s.handler.GetEquipmentStatusNameFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	responseStatus := models.SuccessEquipmentStatusNameOperationResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseStatus)
	if err != nil {
		t.Errorf("Error unmarshalling response: %v", err)
	}
	require.Equal(t, statusToReturn.ID, int(responseStatus.Data.ID))

	s.repository.AssertExpectations(t)
}

func (s *EquipmentStatusNameTestSuite) TestStatus_DeleteEquipmentStatusName_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusID := 1
	data := eqStatusName.DeleteEquipmentStatusNameParams{
		HTTPRequest: &request,
		StatusID:    int64(statusID),
	}
	err := errors.New("test")
	s.repository.On("Delete", ctx, statusID).Return(nil, err)

	handlerFunc := s.handler.DeleteEquipmentStatusNameFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *EquipmentStatusNameTestSuite) TestStatus_DeleteEquipmentStatusName_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	statusID := 1
	data := eqStatusName.DeleteEquipmentStatusNameParams{
		HTTPRequest: &request,
		StatusID:    int64(statusID),
	}
	statusToReturn := &ent.EquipmentStatusName{
		ID: 1,
	}
	s.repository.On("Delete", ctx, statusID).Return(statusToReturn, nil)

	handlerFunc := s.handler.DeleteEquipmentStatusNameFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	responseStatus := models.SuccessEquipmentStatusNameOperationResponse{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseStatus)
	if err != nil {
		t.Errorf("Error unmarshalling response: %v", err)
	}
	require.Equal(t, statusToReturn.ID, int(responseStatus.Data.ID))

	s.repository.AssertExpectations(t)
}
