package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	clientmock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/email"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
)

type RegistrationConfirmTestSuite struct {
	suite.Suite
	logger            *zap.Logger
	userRepository    *repomock.UserRepository
	regConfirmRepo    *repomock.RegistrationConfirmRepository
	emailClient       *clientmock.Sender
	regConfirmService RegistrationConfirm
}

func TestRegistrationConfirmSuite(t *testing.T) {
	s := new(RegistrationConfirmTestSuite)
	suite.Run(t, s)
}

func (s *RegistrationConfirmTestSuite) SetupTest() {
	s.userRepository = &repomock.UserRepository{}
	s.regConfirmRepo = &repomock.RegistrationConfirmRepository{}
	s.emailClient = &clientmock.Sender{}
	s.logger = zap.NewExample()
	ttl := time.Hour
	service := NewRegistrationConfirmService(s.emailClient, s.userRepository, s.regConfirmRepo, s.logger, &ttl)
	s.regConfirmService = service
}

func (s *RegistrationConfirmTestSuite) TestPasswordReset_SendConfirmationLink_UserByLoginErr() {
	t := s.T()
	ctx := context.Background()
	login := "login"
	err := errors.New("error")
	s.userRepository.On("UserByLogin", ctx, login).Return(nil, err)
	errReturn := s.regConfirmService.SendConfirmationLink(ctx, login)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.userRepository.AssertExpectations(t)
}

func (s *RegistrationConfirmTestSuite) TestPasswordReset_SendConfirmationLink_CreateTokenErr() {
	t := s.T()
	ctx := context.Background()
	login := "login"
	user := &ent.User{ID: 1}
	err := errors.New("error")
	s.userRepository.On("UserByLogin", ctx, login).Return(user, nil)
	s.regConfirmRepo.On("CreateToken", ctx, mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Time"), user.ID).Return(err)
	errReturn := s.regConfirmService.SendConfirmationLink(ctx, login)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.userRepository.AssertExpectations(t)
	s.regConfirmRepo.AssertExpectations(t)
}

func (s *RegistrationConfirmTestSuite) TestPasswordReset_SendConfirmationLink_SendLinkErr() {
	t := s.T()
	ctx := context.Background()
	login := "login"
	user := &ent.User{ID: 1, Email: "email"}
	err := errors.New("error")
	s.userRepository.On("UserByLogin", ctx, login).Return(user, nil)
	s.regConfirmRepo.On("CreateToken", ctx, mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Time"), user.ID).Return(nil)
	s.emailClient.On("SendRegistrationConfirmLink", user.Email, user.Login,
		mock.AnythingOfType("string")).Return(err)
	errReturn := s.regConfirmService.SendConfirmationLink(ctx, login)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.userRepository.AssertExpectations(t)
	s.regConfirmRepo.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
}

func (s *RegistrationConfirmTestSuite) TestPasswordReset_SendConfirmationLink_OK() {
	t := s.T()
	ctx := context.Background()
	login := "login"
	user := &ent.User{ID: 1, Email: "email"}
	s.userRepository.On("UserByLogin", ctx, login).Return(user, nil)
	s.regConfirmRepo.On("CreateToken", ctx, mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Time"), user.ID).Return(nil)
	s.emailClient.On("SendRegistrationConfirmLink", user.Email, user.Login,
		mock.AnythingOfType("string")).Return(nil)
	errReturn := s.regConfirmService.SendConfirmationLink(ctx, login)
	assert.NoError(t, errReturn)
	s.userRepository.AssertExpectations(t)
	s.regConfirmRepo.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
}

func (s *RegistrationConfirmTestSuite) TestPasswordReset_VerifyConfirmationToken_GetTokenErr() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	err := errors.New("error")
	s.regConfirmRepo.On("GetToken", ctx, token).Return(nil, err)
	errReturn := s.regConfirmService.VerifyConfirmationToken(ctx, token)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.regConfirmRepo.AssertExpectations(t)
}

func (s *RegistrationConfirmTestSuite) TestPasswordReset_VerifyConfirmationToken_TokenExpired() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.RegistrationConfirm{
		TTL:   time.Now().Add(-1 * time.Hour),
		Token: token,
		Edges: ent.RegistrationConfirmEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	s.regConfirmRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.regConfirmRepo.On("DeleteToken", ctx, token).Return(nil)
	errReturn := s.regConfirmService.VerifyConfirmationToken(ctx, token)
	assert.Error(t, errReturn)
	s.regConfirmRepo.AssertExpectations(t)
}

func (s *RegistrationConfirmTestSuite) TestPasswordReset_VerifyConfirmationToken_DeleteExpiredErr() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.RegistrationConfirm{
		TTL:   time.Now().Add(-1 * time.Hour),
		Token: token,
		Edges: ent.RegistrationConfirmEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	err := errors.New("error")
	s.regConfirmRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.regConfirmRepo.On("DeleteToken", ctx, token).Return(err)
	errReturn := s.regConfirmService.VerifyConfirmationToken(ctx, token)
	assert.Error(t, errReturn)
	s.regConfirmRepo.AssertExpectations(t)
}

func (s *RegistrationConfirmTestSuite) TestPasswordReset_VerifyConfirmationToken_ChangeTxErr() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.RegistrationConfirm{
		TTL:   time.Now().Add(1 * time.Hour),
		Token: token,
		Edges: ent.RegistrationConfirmEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	err := errors.New("error")
	s.regConfirmRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.userRepository.On("ConfirmRegistration", ctx, returnToken.Edges.Users.Login).Return(err)
	errReturn := s.regConfirmService.VerifyConfirmationToken(ctx, token)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.regConfirmRepo.AssertExpectations(t)
	s.userRepository.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
}

func (s *RegistrationConfirmTestSuite) TestPasswordReset_VerifyConfirmationToken_DeleteTokenErr() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.RegistrationConfirm{
		TTL:   time.Now().Add(1 * time.Hour),
		Token: token,
		Edges: ent.RegistrationConfirmEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	s.regConfirmRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.userRepository.On("ConfirmRegistration", ctx, returnToken.Edges.Users.Login).Return(nil)
	s.regConfirmRepo.On("DeleteToken", ctx, token).Return(errors.New("error"))
	errReturn := s.regConfirmService.VerifyConfirmationToken(ctx, token)
	assert.NoError(t, errReturn)
	s.regConfirmRepo.AssertExpectations(t)
	s.userRepository.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
}

func (s *RegistrationConfirmTestSuite) TestPasswordReset_VerifyConfirmationToken_OK() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.RegistrationConfirm{
		TTL:   time.Now().Add(1 * time.Hour),
		Token: token,
		Edges: ent.RegistrationConfirmEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	s.regConfirmRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.userRepository.On("ConfirmRegistration", ctx, returnToken.Edges.Users.Login).Return(nil)
	s.regConfirmRepo.On("DeleteToken", ctx, token).Return(nil)
	errReturn := s.regConfirmService.VerifyConfirmationToken(ctx, token)
	assert.NoError(t, errReturn)
	s.regConfirmRepo.AssertExpectations(t)
	s.userRepository.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
}
