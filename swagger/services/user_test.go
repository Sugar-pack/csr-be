package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
)

type UserServiceTestSuite struct {
	suite.Suite
	tokenRepository *repomock.TokenRepository
	userRepository  *repomock.UserRepository
	userService     UserService
	jwtSecret       string
	logger          *zap.Logger
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) SetupTest() {
	s.tokenRepository = &repomock.TokenRepository{}
	s.userRepository = &repomock.UserRepository{}
	s.jwtSecret = "secret"
	s.logger = zap.NewNop()
	s.userService = NewUserService(s.userRepository, s.tokenRepository, s.jwtSecret, s.logger)
}

func (s *UserServiceTestSuite) TestUserService_GenerateAccessToken_UserNotFound() {
	t := s.T()
	login := "login"
	password := "password"
	ctx := context.Background()

	err := &ent.NotFoundError{}

	s.userRepository.On("GetUserByLogin", ctx, login).Return(nil, err)
	token, b, errGen := s.userService.GenerateAccessToken(ctx, login, password)
	assert.Error(t, errGen)
	assert.True(t, ent.IsNotFound(errGen))
	assert.Empty(t, token)
	assert.False(t, b)
	s.userRepository.AssertExpectations(t)
}

func (s *UserServiceTestSuite) TestUserService_GenerateAccessToken_RepoErr() {
	t := s.T()
	login := "login"
	password := "password"
	ctx := context.Background()

	err := errors.New("error")

	s.userRepository.On("GetUserByLogin", ctx, login).Return(nil, err)
	token, isInternalErr, errGen := s.userService.GenerateAccessToken(ctx, login, password)
	assert.Error(t, errGen)
	assert.False(t, ent.IsNotFound(errGen))
	assert.Empty(t, token)
	assert.True(t, isInternalErr)
	s.userRepository.AssertExpectations(t)
}

func (s *UserServiceTestSuite) TestUserService_GenerateAccessToken_HashCompareErr() {
	t := s.T()
	login := "login"
	password := "password"
	ctx := context.Background()

	user := &ent.User{
		Login:    login,
		Password: "wrong_password",
	}

	s.userRepository.On("GetUserByLogin", ctx, login).Return(user, nil)
	token, isInternalErr, errGen := s.userService.GenerateAccessToken(ctx, login, password)
	assert.Error(t, errGen)
	assert.False(t, ent.IsNotFound(errGen))
	assert.Empty(t, token)
	assert.False(t, isInternalErr)
	s.userRepository.AssertExpectations(t)
}

func (s *UserServiceTestSuite) TestUserService_GenerateAccessToken_TokenRepoErr() {
	t := s.T()
	login := "login"
	password := "password"
	ctx := context.Background()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	user := &ent.User{
		ID:       1,
		Login:    login,
		Password: string(hashedPassword),
		Edges: ent.UserEdges{
			Role: &ent.Role{
				ID:   1,
				Name: "admin",
			},
			Groups: []*ent.Group{
				{
					ID: 1,
				},
			},
		},
	}
	err = errors.New("error")
	s.userRepository.On("GetUserByLogin", ctx, login).Return(user, nil)
	s.tokenRepository.On("CreateTokens", ctx, user.ID, mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).Return(err)
	token, isInternalErr, errGen := s.userService.GenerateAccessToken(ctx, login, password)
	assert.Error(t, errGen)
	assert.False(t, ent.IsNotFound(errGen))
	assert.Empty(t, token)
	assert.True(t, isInternalErr)
	s.userRepository.AssertExpectations(t)
	s.tokenRepository.AssertExpectations(t)
}

func (s *UserServiceTestSuite) TestUserService_GenerateAccessToken_OK() {
	t := s.T()
	login := "login"
	password := "password"
	ctx := context.Background()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	user := &ent.User{
		ID:       1,
		Login:    login,
		Password: string(hashedPassword),
		Edges: ent.UserEdges{
			Role: &ent.Role{
				ID:   1,
				Name: "admin",
			},
			Groups: []*ent.Group{
				{
					ID: 1,
				},
			},
		},
	}
	s.userRepository.On("GetUserByLogin", ctx, login).Return(user, nil)
	s.tokenRepository.On("CreateTokens", ctx, user.ID, mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).Return(nil)
	token, isInternalErr, errGen := s.userService.GenerateAccessToken(ctx, login, password)
	assert.NoError(t, errGen)
	assert.NotEmpty(t, token)
	assert.False(t, isInternalErr)
	s.userRepository.AssertExpectations(t)
	s.tokenRepository.AssertExpectations(t)
}
