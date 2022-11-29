package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type UserServiceTestSuite struct {
	suite.Suite
	tokenRepository *mocks.TokenRepository
	userRepository  *mocks.UserRepository
	userService     domain.TokenManager
	jwtSecret       string
	logger          *zap.Logger
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) SetupTest() {
	s.tokenRepository = &mocks.TokenRepository{}
	s.userRepository = &mocks.UserRepository{}
	s.jwtSecret = "secret"
	s.logger = zap.NewNop()
	s.userService = NewTokenManager(s.userRepository, s.tokenRepository, s.jwtSecret, s.logger)
}

func (s *UserServiceTestSuite) TestUserService_GenerateAccessToken_UserNotFound() {
	t := s.T()
	login := "login"
	password := "password"
	ctx := context.Background()

	err := &ent.NotFoundError{}

	s.userRepository.On("GetUserByLogin", ctx, login).Return(nil, err)
	accessToken, refreshToken, isInternalErr, errGen := s.userService.GenerateTokens(ctx, login, password)
	assert.Error(t, errGen)
	assert.True(t, ent.IsNotFound(errGen))
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.False(t, isInternalErr)
	s.userRepository.AssertExpectations(t)
}

func (s *UserServiceTestSuite) TestUserService_GenerateAccessToken_RepoErr() {
	t := s.T()
	login := "login"
	password := "password"
	ctx := context.Background()

	err := errors.New("error")

	s.userRepository.On("GetUserByLogin", ctx, login).Return(nil, err)
	accessToken, refreshToken, isInternalErr, errGen := s.userService.GenerateTokens(ctx, login, password)
	assert.Error(t, errGen)
	assert.False(t, ent.IsNotFound(errGen))
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
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
	accessToken, refreshToken, isInternalErr, errGen := s.userService.GenerateTokens(ctx, login, password)
	assert.Error(t, errGen)
	assert.False(t, ent.IsNotFound(errGen))
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
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
	accessToken, refreshToken, isInternalErr, errGen := s.userService.GenerateTokens(ctx, login, password)
	assert.Error(t, errGen)
	assert.False(t, ent.IsNotFound(errGen))
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
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
	accessToken, refreshToken, isInternalErr, errGen := s.userService.GenerateTokens(ctx, login, password)
	assert.NoError(t, errGen)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.False(t, isInternalErr)
	s.userRepository.AssertExpectations(t)
	s.tokenRepository.AssertExpectations(t)
}
