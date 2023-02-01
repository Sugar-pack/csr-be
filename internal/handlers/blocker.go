package handlers

import (
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/authentication"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func SetBlockerHandler(logger *zap.Logger, api *operations.BeAPI) {
	blockerRepo := repositories.NewBlockerRepository()
	blockerHandler := NewBlocker(logger)
	api.UsersBlockUserHandler = blockerHandler.BlockUserFunc(blockerRepo)
	api.UsersUnblockUserHandler = blockerHandler.UnblockUserFunc(blockerRepo)
}

type Blocker struct {
	logger *zap.Logger
}

func NewBlocker(logger *zap.Logger) *Blocker {
	return &Blocker{
		logger: logger,
	}
}

func (b Blocker) BlockUserFunc(repository domain.BlockerRepository) users.BlockUserHandlerFunc {
	return func(u users.BlockUserParams, access interface{}) middleware.Responder {
		isManager, err := authentication.IsManager(access)
		if err != nil {
			b.logger.Error("error while getting authorization", zap.Error(err))
			return users.NewBlockUserDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't get authorization"}})
		}
		if !isManager {
			b.logger.Error("user is not manager", zap.Any("access", access))
			return users.NewBlockUserDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "You don't have rights"}})
		}

		userId := int(u.UserID)
		context := u.HTTPRequest.Context()
		err = repository.SetIsBlockedUser(context, userId, true)
		if err != nil {
			b.logger.Error("block user failed", zap.Error(err))
			return users.NewBlockUserDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		return users.NewBlockUserOK()
	}
}

func (b Blocker) UnblockUserFunc(repository domain.BlockerRepository) users.UnblockUserHandlerFunc {
	return func(u users.UnblockUserParams, access interface{}) middleware.Responder {
		isManager, err := authentication.IsManager(access)
		if err != nil {
			b.logger.Error("error while getting authorization", zap.Error(err))
			return users.NewUnblockUserDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't get authorization"}})
		}
		if !isManager {
			b.logger.Error("user is not manager", zap.Any("access", access))
			return users.NewUnblockUserDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "You don't have rights"}})
		}

		userId := int(u.UserID)
		context := u.HTTPRequest.Context()

		err = repository.SetIsBlockedUser(context, userId, false)
		if err != nil {
			b.logger.Error("unblock user failed", zap.Error(err))
			return users.NewUnblockUserDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		return users.NewUnblockUserOK()
	}
}
