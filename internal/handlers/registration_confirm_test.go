package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/registration_confirm"
)

func TestSetRegistrationHandler(t *testing.T) {
	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	regConfService := &mocks.RegistrationConfirmService{}
	api := operations.NewBeAPI(swaggerSpec)
	SetRegistrationHandler(logger, api, regConfService)
	require.NotEmpty(t, api.RegistrationConfirmSendRegistrationConfirmLinkByLoginHandler)
	require.NotEmpty(t, api.RegistrationConfirmVerifyRegistrationConfirmTokenHandler)
}

type RegistrationConfirmHandlerTestSuite struct {
	suite.Suite
	logger            *zap.Logger
	regConfirmService *mocks.RegistrationConfirmService
	handler           *registrationConfirmHandler
}

func TestNewRegistrationConfirmHandler(t *testing.T) {
	s := new(RegistrationConfirmHandlerTestSuite)
	suite.Run(t, s)
}

func (s *RegistrationConfirmHandlerTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.regConfirmService = &mocks.RegistrationConfirmService{}
	s.handler = NewRegistrationConfirmHandler(s.logger, s.regConfirmService)
}

func (s *RegistrationConfirmHandlerTestSuite) TestRegistrationConfirmHandler_SendRegistrationConfirmLinkByLoginFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	login := "login"
	params := registration_confirm.SendRegistrationConfirmLinkByLoginParams{
		HTTPRequest: &request,
		Login:       &models.SendRegistrationConfirmLinkRequest{Data: &models.Login{Login: &login}},
	}
	s.regConfirmService.On("SendConfirmationLink", ctx, login).Return(nil)
	s.regConfirmService.On("IsSendRequired").Return(false)
	handlerFunc := s.handler.SendRegistrationConfirmLinkByLoginFunc()
	resp := handlerFunc.Handle(params)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.regConfirmService.AssertExpectations(t)
}

func (s *RegistrationConfirmHandlerTestSuite) TestRegistrationConfirmHandler_SendRegistrationConfirmLinkByLoginFunc_EmptyLogin() {
	t := s.T()
	request := http.Request{}
	login := ""
	params := registration_confirm.SendRegistrationConfirmLinkByLoginParams{
		HTTPRequest: &request,
		Login:       &models.SendRegistrationConfirmLinkRequest{Data: &models.Login{Login: &login}},
	}
	handlerFunc := s.handler.SendRegistrationConfirmLinkByLoginFunc()
	resp := handlerFunc.Handle(params)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	s.regConfirmService.AssertExpectations(t)
}

func (s *RegistrationConfirmHandlerTestSuite) TestRegistrationConfirmHandler_SendRegistrationConfirmLinkByLoginFunc_ServiceErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	login := "login"
	params := registration_confirm.SendRegistrationConfirmLinkByLoginParams{
		HTTPRequest: &request,
		Login:       &models.SendRegistrationConfirmLinkRequest{Data: &models.Login{Login: &login}},
	}
	err := errors.New("error")
	s.regConfirmService.On("SendConfirmationLink", ctx, login).Return(err)
	handlerFunc := s.handler.SendRegistrationConfirmLinkByLoginFunc()
	resp := handlerFunc.Handle(params)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.regConfirmService.AssertExpectations(t)
}

func (s *RegistrationConfirmHandlerTestSuite) TestRegistrationConfirmHandler_VerifyRegistrationConfirmTokenFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	token := "token"
	params := registration_confirm.VerifyRegistrationConfirmTokenParams{
		HTTPRequest: &request,
		Token:       token,
	}
	s.regConfirmService.On("VerifyConfirmationToken", ctx, token).Return(nil)
	handlerFunc := s.handler.VerifyRegistrationConfirmTokenFunc()
	resp := handlerFunc.Handle(params)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.regConfirmService.AssertExpectations(t)
}

func (s *RegistrationConfirmHandlerTestSuite) TestRegistrationConfirmHandler_VerifyRegistrationConfirmTokenFunc_ServiceErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	token := "token"
	params := registration_confirm.VerifyRegistrationConfirmTokenParams{
		HTTPRequest: &request,
		Token:       token,
	}
	err := errors.New("error")
	s.regConfirmService.On("VerifyConfirmationToken", ctx, token).Return(err)
	handlerFunc := s.handler.VerifyRegistrationConfirmTokenFunc()
	resp := handlerFunc.Handle(params)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.regConfirmService.AssertExpectations(t)
}
