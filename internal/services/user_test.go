package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	require.Error(t, errGen)
	require.True(t, ent.IsNotFound(errGen))
	require.Empty(t, accessToken)
	require.Empty(t, refreshToken)
	require.False(t, isInternalErr)
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
	require.Error(t, errGen)
	require.False(t, ent.IsNotFound(errGen))
	require.Empty(t, accessToken)
	require.Empty(t, refreshToken)
	require.True(t, isInternalErr)
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
	require.Error(t, errGen)
	require.False(t, ent.IsNotFound(errGen))
	require.Empty(t, accessToken)
	require.Empty(t, refreshToken)
	require.False(t, isInternalErr)
	s.userRepository.AssertExpectations(t)
}

func (s *UserServiceTestSuite) TestUserService_GenerateAccessToken_DeletedUserError() {
	t := s.T()
	login := "login"
	password := "password"
	ctx := context.Background()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	user := &ent.User{
		ID:               1,
		IsDeleted: true,
		Login:            login,
		Password:         string(hashedPassword),
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
	accessToken, refreshToken, isInternalErr, errGen := s.userService.GenerateTokens(ctx, login, password)
	require.Error(t, errGen)
	require.Empty(t, accessToken)
	require.Empty(t, refreshToken)
	require.False(t, isInternalErr)
	s.userRepository.AssertExpectations(t)
	s.tokenRepository.AssertExpectations(t)
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
	require.Error(t, errGen)
	require.False(t, ent.IsNotFound(errGen))
	require.Empty(t, accessToken)
	require.Empty(t, refreshToken)
	require.True(t, isInternalErr)
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
	require.NoError(t, errGen)
	require.NotEmpty(t, accessToken)
	require.NotEmpty(t, refreshToken)
	require.False(t, isInternalErr)
	s.userRepository.AssertExpectations(t)
	s.tokenRepository.AssertExpectations(t)
}
