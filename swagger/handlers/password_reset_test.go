package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"

	"github.com/stretchr/testify/assert"

	"github.com/go-openapi/runtime"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	psmock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/passwordservice"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/password_reset"
)

type PasswordResetHandlerTestSuite struct {
	suite.Suite
	logger          *zap.Logger
	passwordService *psmock.PasswordReset
	handler         *passwordResetHandler
}

func TestNewPasswordResetRepository(t *testing.T) {
	s := new(PasswordResetHandlerTestSuite)
	suite.Run(t, s)
}

func (s *PasswordResetHandlerTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.passwordService = &psmock.PasswordReset{}
	s.handler = NewPasswordReset(s.logger, s.passwordService)
}

func (s *PasswordResetHandlerTestSuite) TestPasswordResetHandler_GetPasswordResetLinkFunc_ServiceErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	token := "token"
	params := password_reset.GetPasswordResetLinkParams{
		HTTPRequest: &request,
		Token:       token,
	}
	err := errors.New("error")
	s.passwordService.On("VerifyTokenAndSendPassword", ctx, token).Return(err)
	handlerFunc := s.handler.GetPasswordResetLinkFunc()
	resp := handlerFunc.Handle(params)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.passwordService.AssertExpectations(t)
}

func (s *PasswordResetHandlerTestSuite) TestPasswordResetHandler_GetPasswordResetLinkFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	token := "token"
	params := password_reset.GetPasswordResetLinkParams{
		HTTPRequest: &request,
		Token:       token,
	}
	s.passwordService.On("VerifyTokenAndSendPassword", ctx, token).Return(nil)
	handlerFunc := s.handler.GetPasswordResetLinkFunc()
	resp := handlerFunc.Handle(params)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.passwordService.AssertExpectations(t)
}

func (s *PasswordResetHandlerTestSuite) TestPasswordResetHandler_SendLinkByLoginFunc_EmptyLogin() {
	t := s.T()
	request := http.Request{}
	login := ""
	params := password_reset.SendLinkByLoginParams{
		HTTPRequest: &request,
		Login:       &models.SendPasswordResetLinkRequest{Data: &models.Login{Login: &login}},
	}
	handlerFunc := s.handler.SendLinkByLoginFunc()
	resp := handlerFunc.Handle(params)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	s.passwordService.AssertExpectations(t)
}

func (s *PasswordResetHandlerTestSuite) TestPasswordResetHandler_GSendLinkByLoginFunc_ServiceErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	login := "login"
	params := password_reset.SendLinkByLoginParams{
		HTTPRequest: &request,
		Login:       &models.SendPasswordResetLinkRequest{Data: &models.Login{Login: &login}},
	}
	err := errors.New("error")
	s.passwordService.On("SendResetPasswordLink", ctx, login).Return(err)
	handlerFunc := s.handler.SendLinkByLoginFunc()
	resp := handlerFunc.Handle(params)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.passwordService.AssertExpectations(t)
}

func (s *PasswordResetHandlerTestSuite) TestPasswordResetHandler_GSendLinkByLoginFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	login := "login"
	params := password_reset.SendLinkByLoginParams{
		HTTPRequest: &request,
		Login:       &models.SendPasswordResetLinkRequest{Data: &models.Login{Login: &login}},
	}
	s.passwordService.On("SendResetPasswordLink", ctx, login).Return(nil)
	handlerFunc := s.handler.SendLinkByLoginFunc()
	resp := handlerFunc.Handle(params)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.passwordService.AssertExpectations(t)
}
