package handlers

import (
	"net/http"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"
)

type Blocker struct {
	client *ent.Client
	logger *zap.Logger
}

func NewBlocker(client *ent.Client, logger *zap.Logger) *Blocker {
	return &Blocker{
		client: client,
		logger: logger,
	}
}

func (b Blocker) BlockUserFunc(repository repositories.BlockerRepository) users.BlockUserHandlerFunc {
	return func(u users.BlockUserParams) middleware.Responder {
		userId := int(u.UserID)
		context := u.HTTPRequest.Context()
		err := repository.SetIsBlockedUser(context, userId, true)
		if err != nil {
			return users.NewBlockUserDefault(http.StatusNotFound).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		return users.NewBlockUserOK()
	}
}

func (b Blocker) UnblockUserFunc(repository repositories.BlockerRepository) users.UnblockUserHandlerFunc {
	return func(u users.UnblockUserParams) middleware.Responder {
		userId := int(u.UserID)
		context := u.HTTPRequest.Context()

		err := repository.SetIsBlockedUser(context, userId, false)
		if err != nil {
			return users.NewUnblockUserDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		return users.NewUnblockUserOK()
	}
}
