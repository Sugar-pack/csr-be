package handlers

import (
	"context"
	"errors"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
	"github.com/go-openapi/runtime/middleware"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type User struct {
	client *ent.Client
	logger *zap.Logger
}

func generateJWT(user *ent.User, jwtSecretKey string, ctx context.Context) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["id"] = user.ID
	claims["login"] = user.Login
	claims["role"] = nil
	claims["group"] = nil
	role, err := user.QueryRole().First(ctx)
	if err == nil {
		claims["role"] = map[string]interface{}{
			"id":   role.ID,
			"slug": role.Slug,
		}
	}
	group, err := user.QueryGroups().First(ctx)
	if err == nil {
		claims["group"] = map[string]interface{}{
			"id": group.ID,
		}
	}
	claims["exp"] = time.Now().Add(time.Minute * 300).Unix()

	tokenString, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func NewUser(client *ent.Client, logger *zap.Logger) *User {
	return &User{
		client: client,
		logger: logger,
	}
}

func (c User) LoginUserFunc(jwtSecretKey string) users.LoginHandlerFunc {
	return func(p users.LoginParams) middleware.Responder {
		login := p.Login.Login
		foundUser, err := c.client.User.Query().Where(user.Login(*login)).First(p.HTTPRequest.Context())
		if ent.IsNotFound(err) {
			return users.NewLoginNotFound()
		}
		if err != nil {
			return users.NewLoginDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}
		err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(*p.Login.Password))
		if err != nil {
			return users.NewLoginNotFound()
		}

		token, err := generateJWT(foundUser, jwtSecretKey, p.HTTPRequest.Context())
		if err != nil {
			return users.NewLoginDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		return users.NewLoginOK().WithPayload(&models.LoginSuccessResponse{
			Data: &models.LoginSuccessResponseData{Token: &token},
		})
	}
}

func (c User) PostUserFunc(repository repositories.UserRepository) users.PostUserHandlerFunc {
	return func(p users.PostUserParams) middleware.Responder {
		createdUser, err := repository.CreateUser(p.HTTPRequest.Context(), p.Data)
		if err != nil {
			if ent.IsConstraintError(err) {
				return users.NewPostUserDefault(http.StatusExpectationFailed).WithPayload(
					buildErrorPayload(errors.New("This login is already used")),
				)
			}
			return users.NewPostUserDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		id := int64(createdUser.ID)
		return users.NewPostUserCreated().WithPayload(&models.CreateUserResponse{
			Data: &models.CreateUserResponseData{
				ID:    &id,
				Login: &createdUser.Login,
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

func (c User) AssignRoleToUserFunc(repository repositories.UserRepository) users.AssignRoleToUserHandlerFunc {
	return func(p users.AssignRoleToUserParams, access interface{}) middleware.Responder {
		_, err := authentication.IsAdmin(access)
		if err != nil {
			return users.NewAssignRoleToUserDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}
		ctx := p.HTTPRequest.Context()
		userId := int(p.UserID)
		roleId := int(*p.Data.RoleID)

		foundUser, err := repository.SetUserRole(ctx, userId, roleId)
		if err != nil {
			return users.NewAssignRoleToUserDefault(http.StatusNotFound).WithPayload(buildErrorPayload(err))
		}

		userIdInt64 := int64(foundUser.ID)
		roleIdInt64 := int64(roleId)
		return users.NewAssignRoleToUserOK().WithPayload(&models.GetUserResponse{
			Data: &models.User{
				CreateTime: nil,
				ID:         &userIdInt64,
				RoleID:     &roleIdInt64,
				Login:      &foundUser.Login,
			},
		})
	}
}

func (c User) GetUserById() users.GetUserHandlerFunc {
	return func(p users.GetUserParams) middleware.Responder {
		id := int(p.UserID)
		user, err := c.client.User.Get(p.HTTPRequest.Context(), id)
		if err != nil {
			return users.NewGetUserDefault(http.StatusNotFound).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		id64 := int64(id)
		passportDate := user.PassportIssueDate.String()
		typeString := user.Type.String()
		role, err := user.QueryRole().Only(p.HTTPRequest.Context())
		if err != nil {
			return users.NewGetAllUsersDefault(http.StatusNotFound).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		roleResp := models.GetUserByIDRole{
			ID:   int64(role.ID),
			Name: role.Name,
		}

		return users.NewGetUserCreated().WithPayload(&models.GetUserByID{
			Email:             &user.Email,
			ID:                &id64,
			IsBlocked:         &user.IsBlocked,
			Login:             &user.Login,
			Name:              &user.Name,
			OrgName:           user.OrgName,
			PassportAuthority: user.PassportAuthority,
			PassportIssueDate: &passportDate,
			PassportNumber:    user.PassportNumber,
			PassportSeries:    user.PassportSeries,
			Patronymic:        user.Patronymic,
			PhoneNumber:       user.Phone,
			Role:              &roleResp,
			Surname:           user.Surname,
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
			roleResp := models.GetUserByIDRole{
				ID:   int64(role.ID),
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
				Patronymic:        element.Patronymic,
				PhoneNumber:       element.Phone,
				Role:              &roleResp,
				Surname:           element.Surname,
				Type:              &typeString,
			})
		}

		return users.NewGetAllUsersCreated().WithPayload(listUsers)
	}
}
