package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/registration_confirm"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/messages"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/services"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func SetRegistrationHandler(logger *zap.Logger, api *operations.BeAPI, service domain.RegistrationConfirmService) {
	regConfirmHandler := NewRegistrationConfirmHandler(logger, service)

	api.RegistrationConfirmSendRegistrationConfirmLinkByLoginHandler = regConfirmHandler.SendRegistrationConfirmLinkByLoginFunc()
	api.RegistrationConfirmVerifyRegistrationConfirmTokenHandler = regConfirmHandler.VerifyRegistrationConfirmTokenFunc()
}

type registrationConfirmHandler struct {
	logger     *zap.Logger
	regConfirm domain.RegistrationConfirmService
}

func NewRegistrationConfirmHandler(logger *zap.Logger, regConfirmService domain.RegistrationConfirmService) *registrationConfirmHandler {
	return &registrationConfirmHandler{
		logger:     logger,
		regConfirm: regConfirmService,
	}
}

func (rc registrationConfirmHandler) SendRegistrationConfirmLinkByLoginFunc() registration_confirm.SendRegistrationConfirmLinkByLoginHandlerFunc {
	return func(s registration_confirm.SendRegistrationConfirmLinkByLoginParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		login := *s.Login.Data.Login
		if login == "" {
			rc.logger.Warn("Login is empty")
			return registration_confirm.NewSendRegistrationConfirmLinkByLoginDefault(http.StatusBadRequest).
				WithPayload(buildBadRequestErrorPayload(messages.ErrLoginRequired, ""))
		}
		err := rc.regConfirm.SendConfirmationLink(ctx, login)
		if err != nil {
			rc.logger.Error("Error while sending registration confirmation link", zap.Error(err))
			switch err {
			case services.ErrRegistrationAlreadyConfirmed:
				return registration_confirm.NewSendRegistrationConfirmLinkByLoginDefault(http.StatusInternalServerError).
					WithPayload(buildInternalErrorPayload(messages.ErrRegistrationAlreadyConfirmed, ""))
			case services.ErrUserNotFound:
				return registration_confirm.NewSendRegistrationConfirmLinkByLoginDefault(http.StatusInternalServerError).
					WithPayload(buildInternalErrorPayload(messages.ErrRegistrationCannotFindUser, ""))
			default:
				return registration_confirm.NewSendRegistrationConfirmLinkByLoginDefault(http.StatusInternalServerError).
					WithPayload(buildInternalErrorPayload(messages.ErrRegistrationCannotSend, ""))
			}
		}
		if !rc.regConfirm.IsSendRequired() {
			return registration_confirm.NewSendRegistrationConfirmLinkByLoginOK().WithPayload(
				models.RegistrationConfirmResponse(messages.MsgConfirmationNotRequired))
		}
		return registration_confirm.NewSendRegistrationConfirmLinkByLoginOK().WithPayload(
			models.RegistrationConfirmResponse(messages.MsgConfirmationSent))
	}
}

func (rc registrationConfirmHandler) VerifyRegistrationConfirmTokenFunc() registration_confirm.VerifyRegistrationConfirmTokenHandlerFunc {
	return func(s registration_confirm.VerifyRegistrationConfirmTokenParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		token := s.Token
		err := rc.regConfirm.VerifyConfirmationToken(ctx, token)
		if err != nil {
			rc.logger.Error("Failed to verify confirmation token", zap.Error(err))
			return registration_confirm.NewVerifyRegistrationConfirmTokenDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrFailedToConfirm, ""))
		}
		return registration_confirm.NewVerifyRegistrationConfirmTokenOK().WithPayload(
			models.RegistrationConfirmResponse(messages.MsgRegistrationConfirmed))
	}
}
