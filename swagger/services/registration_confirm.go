package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/email"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Scope of custom errors
var (
	ErrRegistrationAlreadyConfirmed = fmt.Errorf("registration is already confirmed")
	ErrUserNotFound                 = fmt.Errorf("error while getting user by login")
)

type registrationConfirm struct {
	email.Sender
	repositories.UserRepository
	repositories.RegistrationConfirmRepository
	logger *zap.Logger
	ttl    *time.Duration
}

func NewRegistrationConfirmService(emailClient email.Sender, userRepository repositories.UserRepository,
	registrationConfirmRepository repositories.RegistrationConfirmRepository, logger *zap.Logger, ttl *time.Duration) RegistrationConfirm {
	return &registrationConfirm{
		Sender:                        emailClient,
		UserRepository:                userRepository,
		RegistrationConfirmRepository: registrationConfirmRepository,
		logger:                        logger,
		ttl:                           ttl,
	}
}

type RegistrationConfirm interface {
	SendConfirmationLink(ctx context.Context, login string) error
	VerifyConfirmationToken(ctx context.Context, token string) error
	IsSendRequired() bool
}

func (rc *registrationConfirm) IsSendRequired() bool {
	return rc.Sender.IsSendRequired()
}

func (rc *registrationConfirm) SendConfirmationLink(ctx context.Context, login string) error {
	rc.logger.Info("registration confirmation service: send confirmation link", zap.String("login", login))
	token := uuid.New().String()
	user, err := rc.UserRepository.UserByLogin(ctx, login)
	if err != nil {
		err = ErrUserNotFound
		rc.logger.Error("Error while getting user by login", zap.String("login", login), zap.Error(err))
		return err
	}
	if user.IsConfirmed {
		err = ErrRegistrationAlreadyConfirmed
		rc.logger.Error("Error registration is already confirmed", zap.String("login", login), zap.Error(err))
		return err
	}
	err = rc.RegistrationConfirmRepository.CreateToken(ctx, token, time.Now().Add(*rc.ttl), user.ID)
	if err != nil {
		rc.logger.Error("Error while creating token", zap.String("login", login), zap.Error(err))
		return err
	}
	err = rc.Sender.SendRegistrationConfirmLink(user.Email, user.Login, token)
	if err != nil {
		rc.logger.Error("Error while sending confirmation link to email", zap.String("login", login), zap.Error(err))
		return err
	}
	if rc.Sender.IsSendRequired() {
		rc.logger.Info("registration confirmation service: send registration confirmation link")
	} else {
		rc.logger.Info("registration confirmation service: no send, send wasn't required")
	}
	return nil
}

func (rc *registrationConfirm) VerifyConfirmationToken(ctx context.Context, tokenToVerify string) error {
	rc.logger.Info("registration confirmation service: verify token", zap.String("token", tokenToVerify))
	token, err := rc.GetToken(ctx, tokenToVerify)
	if err != nil {
		rc.logger.Error("Error while getting token", zap.String("token", tokenToVerify), zap.Error(err))
		return err
	}
	if token.TTL.Before(time.Now()) {
		rc.logger.Error("Token is expired", zap.String("token", tokenToVerify))
		errDelete := rc.DeleteToken(ctx, tokenToVerify)
		if errDelete != nil {
			return fmt.Errorf("error while deleting expired token: %w", errDelete)
		}
		return errors.New("token expired")
	}
	login := token.Edges.Users.Login
	err = rc.ConfirmRegistration(ctx, login)
	if err != nil {
		rc.logger.Error("Error while confirming registration", zap.String("login", login), zap.Error(err))
		return err
	}

	errDelete := rc.DeleteToken(ctx, tokenToVerify)
	if errDelete != nil {
		rc.logger.Warn("Error while deleting token", zap.String("token", tokenToVerify), zap.Error(errDelete))
	}
	rc.logger.Info("registration confirmation service: verified token")
	return nil
}
