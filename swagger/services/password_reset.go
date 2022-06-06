package services

import (
	"context"
	"errors"
	"fmt"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/utils"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/email"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

type passwordReset struct {
	email.Sender
	repositories.UserRepository
	repositories.PasswordResetRepository
	logger *zap.Logger
	ttl    *time.Duration
}

func NewPasswordResetService(emailClient email.Sender, userRepository repositories.UserRepository,
	passwordResetRepository repositories.PasswordResetRepository, logger *zap.Logger, ttl *time.Duration) PasswordReset {
	return &passwordReset{
		Sender:                  emailClient,
		UserRepository:          userRepository,
		PasswordResetRepository: passwordResetRepository,
		logger:                  logger,
		ttl:                     ttl,
	}
}

type PasswordReset interface {
	SendResetPasswordLink(ctx context.Context, login string) error
	VerifyTokenAndSendPassword(ctx context.Context, tokenToVerify string) error
}

func (p *passwordReset) SendResetPasswordLink(ctx context.Context, login string) error {
	p.logger.Info("password reset service: send reset password link", zap.String("login", login))
	token := uuid.New().String()
	user, err := p.UserRepository.UserByLogin(ctx, login)
	if err != nil {
		p.logger.Error("Error while getting user by login", zap.String("login", login), zap.Error(err))
		return err
	}
	err = p.PasswordResetRepository.CreateToken(ctx, token, time.Now().Add(*p.ttl), user.ID)
	if err != nil {
		p.logger.Error("Error while creating token", zap.String("login", login), zap.Error(err))
		return err
	}
	err = p.Sender.SendResetLink(user.Email, user.Login, token)
	if err != nil {
		p.logger.Error("Error while sending reset link to email", zap.String("login", login), zap.Error(err))
		return err
	}
	p.logger.Info("password reset service: send reset password link")
	return nil
}

func (p *passwordReset) VerifyTokenAndSendPassword(ctx context.Context, tokenToVerify string) error {
	p.logger.Info("password reset service: verify token and send password", zap.String("token", tokenToVerify))
	token, err := p.GetToken(ctx, tokenToVerify)
	if err != nil {
		p.logger.Error("Error while getting token", zap.String("token", tokenToVerify), zap.Error(err))
		return err
	}
	if token.TTL.Before(time.Now()) {
		p.logger.Error("Token is expired", zap.String("token", tokenToVerify))
		errDelete := p.DeleteToken(ctx, tokenToVerify)
		if errDelete != nil {
			return fmt.Errorf("error while deleting expired token: %w", errDelete)
		}
		return errors.New("token expired")
	}
	login := token.Edges.Users.Login
	password, err := utils.GenerateRandomResetPassword()
	if err != nil {
		p.logger.Error("Error while generating password", zap.Error(err))
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		p.logger.Error("Error while hashing password", zap.String("password", password), zap.Error(err))
		return err
	}
	tx, err := p.ChangePasswordByLogin(ctx, login, string(hashedPassword))
	if err != nil {
		p.logger.Error("Error while changing password", zap.String("login", login), zap.Error(err))
		return err
	}
	err = p.SendNewPassword(token.Edges.Users.Email, login, password)
	if err != nil {
		p.logger.Error("Error while sending new password to email", zap.String("login", login), zap.Error(err))
		errRollback := tx.Rollback()
		if errRollback != nil {
			p.logger.Error("Error while rollback", zap.Error(errRollback))
			return errRollback
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		p.logger.Error("Error while changing password", zap.String("login", login), zap.Error(err))
		return err
	}
	errDelete := p.DeleteToken(ctx, tokenToVerify)
	if errDelete != nil {
		p.logger.Warn("Error while deleting token", zap.String("token", tokenToVerify), zap.Error(errDelete))
	}
	p.logger.Info("password reset service: verified token and send password")
	return nil
}
