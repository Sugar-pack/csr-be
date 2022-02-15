package handlers

import (
	"fmt"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/users"
	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"
	"net/http"
)

type User struct {
	client *ent.Client
	logger *zap.Logger
}

func NewUser(client *ent.Client, logger *zap.Logger) *User {
	return &User{
		client: client,
		logger: logger,
	}
}

func (c User) PostUserFunc() users.PostUserHandlerFunc {
	return func(p users.PostUserParams) middleware.Responder {
		e, err := c.client.User.Create().SetName("testClient").Save(p.HTTPRequest.Context())
		if err != nil {
			return users.NewPostUserDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		id := fmt.Sprintf("%d", e.ID)
		return users.NewPostUserCreated().WithPayload(&models.CreateUserResponse{
			Data: &models.CreateUserResponseData{
				ID: &id,
			},
		})
	}
}

func (c User) GetUserFunc() users.GetCurrentUserHandlerFunc {
	return func(p users.GetCurrentUserParams, _ interface{}) middleware.Responder {
		return users.NewGetCurrentUserOK()
	}
}

func (c User) PatchUserFunc() users.PatchUserHandlerFunc {
	return func(p users.PatchUserParams, _ interface{}) middleware.Responder {
		return users.NewPatchUserNoContent()
	}
}
