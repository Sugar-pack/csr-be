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

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	servicemock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/services"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/users"
)

type UserTestSuite struct {
	suite.Suite
	logger            *zap.Logger
	service           *servicemock.UserService
	regConfirmService *servicemock.RegistrationConfirm
	user              *User
	userRepository    *repomock.UserRepository
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

func (s *UserTestSuite) SetupTest() {
	s.logger, _ = zap.NewDevelopment()
	s.service = &servicemock.UserService{}
	s.regConfirmService = &servicemock.RegistrationConfirm{}
	s.userRepository = &repomock.UserRepository{}
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

func (s *UserTestSuite) TestUser_PostUserFunc_LoginExistErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.PostUserFunc(s.userRepository, s.regConfirmService)
	data := users.PostUserParams{
		HTTPRequest: &request,
		Data: &models.UserRegister{
			Login:    &login,
			Password: &password,
		},
	}
	err := &ent.ConstraintError{}
	s.userRepository.On("CreateUser", ctx, data.Data).Return(nil, err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusExpectationFailed, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_PostUserFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.PostUserFunc(s.userRepository, s.regConfirmService)
	data := users.PostUserParams{
		HTTPRequest: &request,
		Data: &models.UserRegister{
			Login:    &login,
			Password: &password,
		},
	}
	err := errors.New("some error")
	s.userRepository.On("CreateUser", ctx, data.Data).Return(nil, err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_PostUserFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.PostUserFunc(s.userRepository, s.regConfirmService)
	data := users.PostUserParams{
		HTTPRequest: &request,
		Data: &models.UserRegister{
			Login:    &login,
			Password: &password,
		},
	}
	user := &ent.User{
		ID:    1,
		Login: login,
	}
	s.userRepository.On("CreateUser", ctx, data.Data).Return(user, nil)
	s.regConfirmService.On("SendConfirmationLink", ctx, login).Return(nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}
