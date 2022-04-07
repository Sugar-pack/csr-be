package handlers

import (
	"net/http"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/users"
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

func (b Blocker) BlockUserFunc() users.BlockUserHandlerFunc {
	return func(u users.BlockUserParams) middleware.Responder {
		userId := int(u.UserID)
		context := u.HTTPRequest.Context()
		user, err := b.client.User.Get(context, userId)
		if err != nil {
			return users.NewBlockUserDefault(http.StatusNotFound).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		user, err = b.client.User.UpdateOne(user).SetIsBlocked(true).Save(context)
		if err != nil {
			return users.NewBlockUserDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		return users.NewBlockUserOK()
	}
}

func (b Blocker) UnblockUserFunc() users.UnblockUserHandlerFunc {
	return func(u users.UnblockUserParams) middleware.Responder {
		userId := int(u.UserID)
		context := u.HTTPRequest.Context()
		user, err := b.client.User.Get(context, userId)
		if err != nil {
			return users.NewUnblockUserDefault(http.StatusNotFound).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		user, err = b.client.User.UpdateOne(user).SetIsBlocked(false).Save(context)
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
