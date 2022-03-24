package handlers

import (
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

		id := int64(e.ID)
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

func (c User) AssignRoleToUserFunc() users.AssignRoleToUserHandlerFunc {
	return func(p users.AssignRoleToUserParams) middleware.Responder {
		context := p.HTTPRequest.Context()
		userId := int(p.UserID)
		roleId := int(*p.Data.RoleID)
		user, err := c.client.User.Get(context, userId)
		if err != nil {
			return users.NewAssignRoleToUserDefault(http.StatusNotFound).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		role, err := c.client.Role.Get(context, roleId)
		if err != nil {
			return users.NewAssignRoleToUserDefault(http.StatusNotFound).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		user, err = c.client.User.UpdateOne(user).SetRole(role).Save(context)
		if err != nil {
			return users.NewAssignRoleToUserDefault(http.StatusNotFound).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		userIdInt64 := int64(user.ID)
		roleIdInt64 := int64(role.ID)
		return users.NewAssignRoleToUserOK().WithPayload(&models.GetUserResponse{
			Data: &models.User{
				CreateTime: nil,
				ID:         &userIdInt64,
				RoleID:     &roleIdInt64,
			},
		})
	}
}
