package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/registration_confirm"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/services"
)

func SetRegistrationHandler(logger *zap.Logger, api *operations.BeAPI, service services.RegistrationConfirm) {
	regConfirmHandler := NewRegistrationConfirmHandler(logger, service)

	api.RegistrationConfirmSendRegistrationConfirmLinkByLoginHandler = regConfirmHandler.SendRegistrationConfirmLinkByLoginFunc()
	api.RegistrationConfirmVerifyRegistrationConfirmTokenHandler = regConfirmHandler.VerifyRegistrationConfirmTokenFunc()
}

type registrationConfirmHandler struct {
	logger     *zap.Logger
	regConfirm services.RegistrationConfirm
}

func NewRegistrationConfirmHandler(logger *zap.Logger, regConfirmService services.RegistrationConfirm) *registrationConfirmHandler {
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
				WithPayload(buildStringPayload("Login is required"))
		}
		err := rc.regConfirm.SendConfirmationLink(ctx, login)
		if err != nil {
			rc.logger.Error("Error while sending registration confirmation link", zap.Error(err))
			switch err {
			case services.ErrRegistrationAlreadyConfirmed:
				return registration_confirm.NewSendRegistrationConfirmLinkByLoginDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Registration is already confirmed."))
			case services.ErrUserNotFound:
				return registration_confirm.NewSendRegistrationConfirmLinkByLoginDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Can't find this user, registration confirmation link wasn't send"))
			default:
				return registration_confirm.NewSendRegistrationConfirmLinkByLoginDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Can't send registration confirmation link. Please try again later"))
			}
		}
		if !rc.regConfirm.IsSendRequired() {
			return registration_confirm.NewSendRegistrationConfirmLinkByLoginOK().WithPayload("Confirmation link was not sent to email, sending parameter was set to false and not required")
		}
		return registration_confirm.NewSendRegistrationConfirmLinkByLoginOK().WithPayload("Confirmation link was sent")
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
				WithPayload(buildStringPayload("Failed to verify confirmation token. Please try again later"))
		}
		return registration_confirm.NewVerifyRegistrationConfirmTokenOK().WithPayload("You have successfully confirmed registration")
	}
}
