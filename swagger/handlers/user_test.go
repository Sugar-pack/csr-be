package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/users"

	servicemock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/services"
)

type UserTestSuite struct {
	suite.Suite
	logger  *zap.Logger
	service *servicemock.UserService
	user    *User
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

func (s *UserTestSuite) SetupTest() {
	s.logger, _ = zap.NewDevelopment()
	s.service = &servicemock.UserService{}
	s.user = &User{
		logger: s.logger,
	} // it will be rewritten after full user handler refactoring
}

func (s *UserTestSuite) TestUser_LoginUserFunc_InternalErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.LoginUserFunc(s.service)
	data := users.LoginParams{
		HTTPRequest: &request,
		Login: &models.LoginInfo{
			Login:    &login,
			Password: &password,
		},
	}
	err := errors.New("some error")
	s.service.On("GenerateAccessToken", ctx, login, password).Return("", true, err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.service.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_LoginUserFunc_UnauthorizedErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.LoginUserFunc(s.service)
	data := users.LoginParams{
		HTTPRequest: &request,
		Login: &models.LoginInfo{
			Login:    &login,
			Password: &password,
		},
	}
	err := errors.New("some error")
	s.service.On("GenerateAccessToken", ctx, login, password).Return("", false, err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)
	s.service.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_LoginUserFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.LoginUserFunc(s.service)
	data := users.LoginParams{
		HTTPRequest: &request,
		Login: &models.LoginInfo{
			Login:    &login,
			Password: &password,
		},
	}
	s.service.On("GenerateAccessToken", ctx, login, password).Return("", false, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.service.AssertExpectations(t)
}
