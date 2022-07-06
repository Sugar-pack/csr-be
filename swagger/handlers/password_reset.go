package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/password_reset"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/services"
)

func SetPasswordResetHandler(logger *zap.Logger, api *operations.BeAPI, passwordService services.PasswordReset) {
	PasswordResetHandler := NewPasswordReset(logger, passwordService)

	api.PasswordResetSendLinkByLoginHandler = PasswordResetHandler.SendLinkByLoginFunc()
	api.PasswordResetGetPasswordResetLinkHandler = PasswordResetHandler.GetPasswordResetLinkFunc()
}

type passwordResetHandler struct {
	logger        *zap.Logger
	passwordReset services.PasswordReset
}

func NewPasswordReset(logger *zap.Logger, passwordService services.PasswordReset) *passwordResetHandler {
	return &passwordResetHandler{
		logger:        logger,
		passwordReset: passwordService,
	}
}

func (c passwordResetHandler) SendLinkByLoginFunc() password_reset.SendLinkByLoginHandlerFunc {
	return func(s password_reset.SendLinkByLoginParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		login := *s.Login.Data.Login
		if login == "" {
			c.logger.Warn("Login is empty")
			return password_reset.NewSendLinkByLoginDefault(http.StatusBadRequest).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "Login is required",
					},
				})
		}
		err := c.passwordReset.SendResetPasswordLink(ctx, login)
		if err != nil {
			c.logger.Error("Error while sending reset password link", zap.Error(err))
			return password_reset.NewSendLinkByLoginDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "Can't send reset password link. Please try again later",
					},
				})
		}
		return password_reset.NewSendLinkByLoginOK().WithPayload("Reset link sent")
	}
}

func (c passwordResetHandler) GetPasswordResetLinkFunc() password_reset.GetPasswordResetLinkHandlerFunc {
	return func(s password_reset.GetPasswordResetLinkParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		token := s.Token
		err := c.passwordReset.VerifyTokenAndSendPassword(ctx, token)
		if err != nil {
			c.logger.Error("Failed to verify token or send email", zap.Error(err))
			return password_reset.NewGetPasswordResetLinkDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: "Failed to verify token. Please try again later",
				},
			})
		}
		return password_reset.NewGetPasswordResetLinkOK().WithPayload("Password successfully reset. Check your email")
	}
}
