package handlers

import (
	"net/http"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/email_confirm"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"
)

func SetEmailConfirmHandler(logger *zap.Logger, api *operations.BeAPI, service domain.ChangeEmailService) {
	emailConfirmHandler := NewEmailConfirmHandler(logger, service)
	api.EmailConfirmVerifyEmailConfirmTokenHandler = emailConfirmHandler.VerifyEmailConfirmTokenFunc()
}

type emailConfirmHandler struct {
	logger       *zap.Logger
	emailConfirm domain.ChangeEmailService
}

func NewEmailConfirmHandler(logger *zap.Logger, changeEmailService domain.ChangeEmailService) *emailConfirmHandler {
	return &emailConfirmHandler{
		logger:       logger,
		emailConfirm: changeEmailService,
	}
}

func (e emailConfirmHandler) VerifyEmailConfirmTokenFunc() email_confirm.VerifyEmailConfirmTokenHandlerFunc {
	return func(s email_confirm.VerifyEmailConfirmTokenParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		token := s.Token
		err := e.emailConfirm.VerifyTokenAndChangeEmail(ctx, token)
		if err != nil {
			e.logger.Error("Failed to verify email confirmation token", zap.Error(err))
			return email_confirm.NewVerifyEmailConfirmTokenDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Failed to verify email confirmation token. Please try again later"))
		}

		return email_confirm.NewVerifyEmailConfirmTokenOK().WithPayload("You have successfully confirmed new email")
	}
}
