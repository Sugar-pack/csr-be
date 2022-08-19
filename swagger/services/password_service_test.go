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
	utilmock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/utils"
)

type PasswordResetTestSuite struct {
	suite.Suite
	txMock            *repomock.Transaction
	logger            *zap.Logger
	userRepository    *repomock.UserRepository
	passwordRepo      *repomock.PasswordResetRepository
	emailClient       *clientmock.Sender
	passwordGenerator *utilmock.PasswordGenerator
	passwordService   PasswordReset
}

func TestPasswordClientSuite(t *testing.T) {
	s := new(PasswordResetTestSuite)
	suite.Run(t, s)
}

func (s *PasswordResetTestSuite) SetupTest() {
	s.txMock = &repomock.Transaction{}
	s.userRepository = &repomock.UserRepository{}
	s.passwordRepo = &repomock.PasswordResetRepository{}
	s.emailClient = &clientmock.Sender{}
	s.passwordGenerator = &utilmock.PasswordGenerator{}
	s.logger = zap.NewExample()
	ttl := time.Hour
	service := NewPasswordResetService(s.emailClient, s.userRepository, s.passwordRepo, s.logger, &ttl, s.passwordGenerator)
	s.passwordService = service
}

func (s *PasswordResetTestSuite) TestPasswordReset_SendResetPasswordLink_UserByLoginErr() {
	t := s.T()
	ctx := context.Background()
	login := "login"
	err := errors.New("error")
	s.userRepository.On("UserByLogin", ctx, login).Return(nil, err)
	errReturn := s.passwordService.SendResetPasswordLink(ctx, login)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.userRepository.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_SendResetPasswordLink_CreateTokenErr() {
	t := s.T()
	ctx := context.Background()
	login := "login"
	user := &ent.User{ID: 1}
	err := errors.New("error")
	s.userRepository.On("UserByLogin", ctx, login).Return(user, nil)
	s.passwordRepo.On("CreateToken", ctx, mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Time"), user.ID).Return(err)
	errReturn := s.passwordService.SendResetPasswordLink(ctx, login)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.userRepository.AssertExpectations(t)
	s.passwordRepo.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_SendResetPasswordLink_SendLinkErr() {
	t := s.T()
	ctx := context.Background()
	login := "login"
	user := &ent.User{ID: 1, Email: "email"}
	err := errors.New("error")
	s.userRepository.On("UserByLogin", ctx, login).Return(user, nil)

	s.passwordRepo.On("CreateToken", ctx, mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Time"), user.ID).Return(nil)

	s.emailClient.On("SendResetLink", user.Email, user.Login,
		mock.AnythingOfType("string")).Return(err)

	errReturn := s.passwordService.SendResetPasswordLink(ctx, login)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.userRepository.AssertExpectations(t)
	s.passwordRepo.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_SendResetPasswordLink_OK() {
	t := s.T()
	ctx := context.Background()
	login := "login"
	user := &ent.User{ID: 1, Email: "email"}

	s.userRepository.On("UserByLogin", ctx, login).Return(user, nil)

	s.passwordRepo.On("CreateToken", ctx, mock.AnythingOfType("string"),
		mock.AnythingOfType("time.Time"), user.ID).Return(nil)

	s.emailClient.On("SendResetLink", user.Email, user.Login,
		mock.AnythingOfType("string")).Return(nil)
	s.emailClient.On("IsSendRequired").Return(false)

	errReturn := s.passwordService.SendResetPasswordLink(ctx, login)
	assert.NoError(t, errReturn)
	s.userRepository.AssertExpectations(t)
	s.passwordRepo.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_VerifyTokenAndSendPassword_GetTokenErr() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	err := errors.New("error")
	s.passwordRepo.On("GetToken", ctx, token).Return(nil, err)
	errReturn := s.passwordService.VerifyTokenAndSendPassword(ctx, token)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.passwordRepo.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_VerifyTokenAndSendPassword_TokenExpired() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.PasswordReset{
		TTL:   time.Now().Add(-1 * time.Hour),
		Token: token,
		Edges: ent.PasswordResetEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	s.passwordRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.passwordRepo.On("DeleteToken", ctx, token).Return(nil)
	errReturn := s.passwordService.VerifyTokenAndSendPassword(ctx, token)
	assert.Error(t, errReturn)
	s.passwordRepo.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_VerifyTokenAndSendPassword_DeleteExpiredErr() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.PasswordReset{
		TTL:   time.Now().Add(-1 * time.Hour),
		Token: token,
		Edges: ent.PasswordResetEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	err := errors.New("error")
	s.passwordRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.passwordRepo.On("DeleteToken", ctx, token).Return(err)
	errReturn := s.passwordService.VerifyTokenAndSendPassword(ctx, token)
	assert.Error(t, errReturn)
	s.passwordRepo.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_VerifyTokenAndSendPassword_PasswordGenErr() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.PasswordReset{
		TTL:   time.Now().Add(1 * time.Hour),
		Token: token,
		Edges: ent.PasswordResetEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	err := errors.New("error")
	newPassword := "new password"
	s.passwordRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.passwordGenerator.On("NewPassword").Return(newPassword, err)

	errReturn := s.passwordService.VerifyTokenAndSendPassword(ctx, token)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.passwordRepo.AssertExpectations(t)
	s.userRepository.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
	s.passwordRepo.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_VerifyTokenAndSendPassword_ChangeTxErr() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.PasswordReset{
		TTL:   time.Now().Add(1 * time.Hour),
		Token: token,
		Edges: ent.PasswordResetEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	err := errors.New("error")
	newPassword := "new password"
	s.passwordRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.userRepository.On("ChangePasswordByLogin", ctx, returnToken.Edges.Users.Login,
		mock.AnythingOfType("string")).Return(s.txMock, err)
	s.passwordGenerator.On("NewPassword").Return(newPassword, nil)

	errReturn := s.passwordService.VerifyTokenAndSendPassword(ctx, token)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.passwordRepo.AssertExpectations(t)
	s.userRepository.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
	s.passwordRepo.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_VerifyTokenAndSendPassword_RollbackErr() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.PasswordReset{
		TTL:   time.Now().Add(1 * time.Hour),
		Token: token,
		Edges: ent.PasswordResetEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	err := errors.New("error")
	newPassword := "new password"
	s.passwordRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.passwordGenerator.On("NewPassword").Return(newPassword, nil)
	s.userRepository.On("ChangePasswordByLogin", ctx, returnToken.Edges.Users.Login,
		mock.AnythingOfType("string")).Return(s.txMock, nil)
	s.txMock.On("Rollback").Return(err)
	s.emailClient.On("SendNewPassword", returnToken.Edges.Users.Email,
		returnToken.Edges.Users.Login, newPassword).Return(err)
	errReturn := s.passwordService.VerifyTokenAndSendPassword(ctx, token)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.passwordRepo.AssertExpectations(t)
	s.userRepository.AssertExpectations(t)
	s.txMock.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
	s.passwordRepo.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_VerifyTokenAndSendPassword_SendEmailErr() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.PasswordReset{
		TTL:   time.Now().Add(1 * time.Hour),
		Token: token,
		Edges: ent.PasswordResetEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	err := errors.New("error")
	newPassword := "new password"
	s.passwordRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.passwordGenerator.On("NewPassword").Return(newPassword, nil)
	s.userRepository.On("ChangePasswordByLogin", ctx, returnToken.Edges.Users.Login,
		mock.AnythingOfType("string")).Return(s.txMock, nil)
	s.txMock.On("Rollback").Return(nil)
	s.emailClient.On("SendNewPassword", returnToken.Edges.Users.Email,
		returnToken.Edges.Users.Login, newPassword).Return(err)
	errReturn := s.passwordService.VerifyTokenAndSendPassword(ctx, token)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.passwordRepo.AssertExpectations(t)
	s.userRepository.AssertExpectations(t)
	s.txMock.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
	s.passwordRepo.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_VerifyTokenAndSendPassword_CommitErr() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.PasswordReset{
		TTL:   time.Now().Add(1 * time.Hour),
		Token: token,
		Edges: ent.PasswordResetEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	err := errors.New("error")
	newPassword := "new password"
	s.passwordRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.passwordGenerator.On("NewPassword").Return(newPassword, nil)
	s.userRepository.On("ChangePasswordByLogin", ctx, returnToken.Edges.Users.Login,
		mock.AnythingOfType("string")).Return(s.txMock, nil)
	s.txMock.On("Commit").Return(err)
	s.emailClient.On("SendNewPassword", returnToken.Edges.Users.Email,
		returnToken.Edges.Users.Login, newPassword).Return(nil)
	errReturn := s.passwordService.VerifyTokenAndSendPassword(ctx, token)
	assert.Error(t, errReturn)
	assert.Equal(t, err, errReturn)
	s.passwordRepo.AssertExpectations(t)
	s.userRepository.AssertExpectations(t)
	s.txMock.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
	s.passwordRepo.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_VerifyTokenAndSendPassword_DeleteTokenErr() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.PasswordReset{
		TTL:   time.Now().Add(1 * time.Hour),
		Token: token,
		Edges: ent.PasswordResetEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	newPassword := "new password"
	s.passwordRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.passwordGenerator.On("NewPassword").Return(newPassword, nil)
	s.userRepository.On("ChangePasswordByLogin", ctx, returnToken.Edges.Users.Login,
		mock.AnythingOfType("string")).Return(s.txMock, nil)
	s.txMock.On("Commit").Return(nil)
	s.emailClient.On("SendNewPassword", returnToken.Edges.Users.Email,
		returnToken.Edges.Users.Login, newPassword).Return(nil)
	s.passwordRepo.On("DeleteToken", ctx, token).Return(errors.New("error"))
	s.emailClient.On("IsSendRequired").Return(false)
	errReturn := s.passwordService.VerifyTokenAndSendPassword(ctx, token)
	assert.NoError(t, errReturn)
	s.passwordRepo.AssertExpectations(t)
	s.userRepository.AssertExpectations(t)
	s.txMock.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
	s.passwordRepo.AssertExpectations(t)
}

func (s *PasswordResetTestSuite) TestPasswordReset_VerifyTokenAndSendPassword_OK() {
	t := s.T()
	ctx := context.Background()
	token := "token"
	returnToken := &ent.PasswordReset{
		TTL:   time.Now().Add(1 * time.Hour),
		Token: token,
		Edges: ent.PasswordResetEdges{
			Users: &ent.User{Login: "login"},
		},
	}
	newPassword := "new password"
	s.passwordRepo.On("GetToken", ctx, token).Return(returnToken, nil)
	s.passwordGenerator.On("NewPassword").Return(newPassword, nil)
	s.userRepository.On("ChangePasswordByLogin", ctx, returnToken.Edges.Users.Login,
		mock.AnythingOfType("string")).Return(s.txMock, nil)
	s.txMock.On("Commit").Return(nil)
	s.emailClient.On("SendNewPassword", returnToken.Edges.Users.Email,
		returnToken.Edges.Users.Login, newPassword).Return(nil)
	s.passwordRepo.On("DeleteToken", ctx, token).Return(nil)
	s.emailClient.On("IsSendRequired").Return(false)
	errReturn := s.passwordService.VerifyTokenAndSendPassword(ctx, token)
	assert.NoError(t, errReturn)
	s.passwordRepo.AssertExpectations(t)
	s.userRepository.AssertExpectations(t)
	s.txMock.AssertExpectations(t)
	s.emailClient.AssertExpectations(t)
	s.passwordRepo.AssertExpectations(t)
}
