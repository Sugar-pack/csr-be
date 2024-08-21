package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type emailChange struct {
	domain.Sender
	domain.UserRepository
	domain.EmailConfirmRepository
	logger *zap.Logger
	ttl    time.Duration
}

func NewEmailChangeService(senderClient domain.Sender, userRepository domain.UserRepository,
	emailConfirmRepository domain.EmailConfirmRepository, logger *zap.Logger,
) domain.ChangeEmailService {
	return &emailChange{
		Sender:                 senderClient,
		UserRepository:         userRepository,
		EmailConfirmRepository: emailConfirmRepository,
		logger:                 logger,
	}
}

func (e *emailChange) SendEmailConfirmationLink(ctx context.Context, login, email string) error {
	e.logger.Info("new email confirmation service: send confirmation link", zap.String("login", login))
	token := uuid.New().String()
	user, err := e.UserRepository.UserByLogin(ctx, login)
	if err != nil {
		e.logger.Error(
			"Error while getting user by login", zap.String("login", login),
			zap.Error(ErrUserNotFound),
		)
		return err
	}

	err = e.EmailConfirmRepository.CreateToken(ctx, token, time.Now().Add(e.ttl), user.ID, email)
	if err != nil {
		e.logger.Error(
			"Error while creating token during sending confirmation link to new email",
			zap.String("login", login), zap.Error(err),
		)
		return err
	}

	err = e.Sender.SendEmailConfirmationLink(email, user.Login, token)
	if err != nil {
		e.logger.Error(
			"Error while sending confirmation link to new email", zap.String("login", login), zap.Error(err),
		)
		return err
	}

	return nil
}

func (e *emailChange) VerifyTokenAndChangeEmail(ctx context.Context, tokenToVerify string) error {
	e.logger.Info("change email service: verify token and change email", zap.String("token", tokenToVerify))
	token, err := e.GetToken(ctx, tokenToVerify)
	if err != nil {
		e.logger.Error("Error while getting token during changing email",
			zap.String("token", tokenToVerify), zap.Error(err))
		return err
	}

	if token.TTL.Before(time.Now()) {
		e.logger.Error("Token is expired", zap.String("token", tokenToVerify))
		errDelete := e.DeleteToken(ctx, tokenToVerify)
		if errDelete != nil {
			return fmt.Errorf("error while deleting expired token: %w", errDelete)
		}
		return errors.New("token expired")
	}

	login := token.Edges.Users.Login
	err = e.ChangeEmailByLogin(ctx, login, token.Email)
	if err != nil {
		e.logger.Error("Error while changing email", zap.String("login", login), zap.Error(err))
		return err
	}

	newLogin := token.Email
	err = e.UpdateLogin(ctx, login, newLogin)
	if err != nil {
		e.logger.Error("Error while changing login", zap.String("login", login), zap.Error(err))
		return err
	}

	errDelete := e.DeleteToken(ctx, tokenToVerify)
	if errDelete != nil {
		e.logger.Warn("Error while deleting token during changing email",
			zap.String("token", tokenToVerify), zap.Error(errDelete),
		)
	}
	return nil
}
