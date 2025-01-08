package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/email_confirm"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type EmailConfirmHandlerTestSuite struct {
	suite.Suite
	logger       *zap.Logger
	emailService *mocks.ChangeEmailService
	handler      *emailConfirmHandler
}

func TestEmailConfirmHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:emailhandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}

	api := operations.NewBeAPI(swaggerSpec)

	emailService := &mocks.ChangeEmailService{}

	SetEmailConfirmHandler(logger, api, emailService)

	require.NotNil(t, api.EmailConfirmVerifyEmailConfirmTokenHandler)
}

func TestNewEmailConfirmRepository(t *testing.T) {
	s := new(EmailConfirmHandlerTestSuite)
	suite.Run(t, s)
}

func (s *EmailConfirmHandlerTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.emailService = &mocks.ChangeEmailService{}
	s.handler = NewEmailConfirmHandler(s.logger, s.emailService)
}

func (s *EmailConfirmHandlerTestSuite) TestEmailConfirmHandler_VerifyEmailConfirmationLinkFunc_Ok() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	token := "token"
	params := email_confirm.VerifyEmailConfirmTokenParams{
		HTTPRequest: &request,
		Token:       token,
	}

	s.emailService.On("VerifyTokenAndChangeEmail", ctx, token).Return(nil)
	handlerFunc := s.handler.VerifyEmailConfirmTokenFunc()
	resp := handlerFunc.Handle(params)
	handlerFunc.Handle(params)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.emailService.AssertExpectations(t)
}

func (s *EmailConfirmHandlerTestSuite) TestEmailConfirmHandler_VerifyEmailConfirmationLinkFunc_ServiceErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	token := "token"
	params := email_confirm.VerifyEmailConfirmTokenParams{
		HTTPRequest: &request,
		Token:       token,
	}

	err := errors.New("error")
	s.emailService.On("VerifyTokenAndChangeEmail", ctx, token).Return(err)
	handlerFunc := s.handler.VerifyEmailConfirmTokenFunc()
	resp := handlerFunc.Handle(params)
	handlerFunc.Handle(params)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.emailService.AssertExpectations(t)
}
