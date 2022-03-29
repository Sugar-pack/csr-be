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

func (c User) GetUserById() users.GetUserHandlerFunc {
	return func(p users.GetUserParams) middleware.Responder {
		id := int(p.UserID)
		c, err := c.client.User.Get(p.HTTPRequest.Context(), id)
		if err != nil {
			return users.NewGetUserDefault(http.StatusNotFound).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		id64 := int64(id)
		passportDate := c.PassportIssueDate.String()
		typeString := c.Type.String()
		role, err := c.QueryRole().Only(p.HTTPRequest.Context())
		if err != nil {
			return users.NewGetUserDefault(http.StatusNotFound).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		roleId := int64(role.ID)
		roleResp := models.GetUserByIDRole{
			ID:   roleId,
			Name: role.Name,
		}

		return users.NewGetUserCreated().WithPayload(&models.GetUserByID{
			Email:             &c.Email,
			ID:                &id64,
			IsBlocked:         &c.IsBlocked,
			Login:             &c.Login,
			Name:              &c.Name,
			OrgName:           c.OrgName,
			PassportAuthority: c.PassportAuthority,
			PassportIssueDate: &passportDate,
			PassportNumber:    c.PassportNumber,
			PassportSeries:    c.PassportSeries,
			Patronomic:        c.Patronymic,
			PhoneNumber:       c.Phone,
			Role:              &roleResp,
			Surname:           c.Surname,
			Type:              &typeString,
		})
	}
}

func (c User) GetUsersList() users.GetAllUsersHandlerFunc {
	return func(p users.GetAllUsersParams) middleware.Responder {
		all, err := c.client.User.Query().All(p.HTTPRequest.Context())
		if err != nil {
			return users.NewGetAllUsersDefault(http.StatusNotFound).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		listUsers := models.GetListUsers{}
		for _, element := range all {

			id64 := int64(element.ID)
			passportDate := element.PassportIssueDate.String()
			typeString := element.Type.String()

			role, err := element.QueryRole().Only(p.HTTPRequest.Context())
			if err != nil {
				return users.NewGetAllUsersDefault(http.StatusNotFound).WithPayload(&models.Error{
					Data: &models.ErrorData{
						Message: err.Error(),
					},
				})
			}
			roleId := int64(role.ID)
			roleResp := models.GetUserByIDRole{
				ID:   roleId,
				Name: role.Name,
			}

			listUsers = append(listUsers, &models.GetUserByID{
				Email:             &element.Email,
				ID:                &id64,
				IsBlocked:         &element.IsBlocked,
				Login:             &element.Login,
				Name:              &element.Name,
				OrgName:           element.OrgName,
				PassportAuthority: element.PassportAuthority,
				PassportIssueDate: &passportDate,
				PassportNumber:    element.PassportNumber,
				PassportSeries:    element.PassportSeries,
				Patronomic:        element.Patronymic,
				PhoneNumber:       element.Phone,
				Role:              &roleResp,
				Surname:           element.Surname,
				Type:              &typeString,
			})
		}
		return users.NewGetAllUsersCreated().WithPayload(listUsers)
	}
}
