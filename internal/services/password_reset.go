package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type passwordReset struct {
	domain.Sender
	domain.UserRepository
	domain.PasswordResetRepository
	domain.PasswordGenerator
	logger *zap.Logger
	ttl    time.Duration
}

func NewPasswordResetService(emailClient domain.Sender, userRepository domain.UserRepository,
	passwordResetRepository domain.PasswordResetRepository, logger *zap.Logger, ttl time.Duration,
	passwordGenerator domain.PasswordGenerator) domain.PasswordResetService {
	return &passwordReset{
		Sender:                  emailClient,
		UserRepository:          userRepository,
		PasswordResetRepository: passwordResetRepository,
		logger:                  logger,
		ttl:                     ttl,
		PasswordGenerator:       passwordGenerator,
	}
}

func (p *passwordReset) SendResetPasswordLink(ctx context.Context, login string) error {
	p.logger.Info("password reset service: send reset password link", zap.String("login", login))
	token := uuid.New().String()
	user, err := p.UserRepository.UserByLogin(ctx, login)
	if err != nil {
		p.logger.Error("Error while getting user by login", zap.String("login", login), zap.Error(err))
		return err
	}
	err = p.PasswordResetRepository.CreateToken(ctx, token, time.Now().Add(p.ttl), user.ID)
	if err != nil {
		p.logger.Error("Error while creating token", zap.String("login", login), zap.Error(err))
		return err
	}
	err = p.Sender.SendResetLink(user.Email, user.Login, token)
	if err != nil {
		p.logger.Error("Error while sending reset link to email", zap.String("login", login), zap.Error(err))
		return err
	}
	if !p.Sender.IsSendRequired() {
		p.logger.Info("password reset service: reset password link wasn't send, sending parameter is set to false and send email is not required")
	} else {
		p.logger.Info("password reset service: send reset password link")
	}
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
	password, err := p.PasswordGenerator.NewPassword()
	if err != nil {
		p.logger.Error("Error while generating password", zap.Error(err))
		return err
	}
	err = p.ChangePasswordByLogin(ctx, login, password)
	if err != nil {
		p.logger.Error("Error while changing password", zap.String("login", login), zap.Error(err))
		return err
	}
	err = p.SendNewPassword(token.Edges.Users.Email, login, password)
	if err != nil {
		p.logger.Error("Error while sending new password to email", zap.String("login", login), zap.Error(err))
		return err
	}
	errDelete := p.DeleteToken(ctx, tokenToVerify)
	if errDelete != nil {
		p.logger.Warn("Error while deleting token", zap.String("token", tokenToVerify), zap.Error(errDelete))
	}
	if !p.Sender.IsSendRequired() {
		p.logger.Info("password reset service: verified token, password wasn't send, sending parameter is set to false and send email is not required")
	} else {
		p.logger.Info("password reset service: verified token and send password")
	}
	return nil
}
