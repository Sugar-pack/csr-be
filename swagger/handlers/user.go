package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/services"

	"github.com/go-openapi/runtime/middleware"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/token"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

const (
	accessExpireTime  = 15 * time.Minute
	refreshExpireTime = 148 * time.Hour
)

type User struct {
	client *ent.Client
	logger *zap.Logger
}

func generateJWT(ctx context.Context, user *ent.User, jwtSecretKey string) (string, error) {
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
	claims["exp"] = time.Now().Add(accessExpireTime).Unix()

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

func (c User) LoginUserFunc(service services.UserService) users.LoginHandlerFunc {
	return func(p users.LoginParams) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		login := *p.Login.Login
		password := *p.Login.Password
		accessToken, isInternalErr, err := service.GenerateAccessToken(ctx, login, password)
		if err != nil {
			if isInternalErr {
				return users.NewLoginDefault(http.StatusInternalServerError)
			}
			return users.NewLoginUnauthorized().WithPayload("Invalid login or password")
		}
		return users.NewLoginOK().WithPayload(&models.AccessToken{AccessToken: &accessToken})
	}
}

func (c User) PostUserFunc(repository repositories.UserRepository) users.PostUserHandlerFunc {
	return func(p users.PostUserParams) middleware.Responder {
		createdUser, err := repository.CreateUser(p.HTTPRequest.Context(), p.Data)
		if err != nil {
			if ent.IsConstraintError(err) {
				return users.NewPostUserDefault(http.StatusExpectationFailed).WithPayload(
					buildErrorPayload(errors.New("login is already used")),
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

func (c User) Refresh(jwtSecretKey string) users.RefreshHandlerFunc {
	return func(p users.RefreshParams) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		claims := jwt.MapClaims{}
		refreshToken, err := jwt.ParseWithClaims(*p.RefreshToken.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("error decoding token")
			}
			return []byte(jwtSecretKey), nil
		})

		if errors.Is(err, jwt.ErrTokenExpired) {
			_, err = c.client.Token.Delete().Where(token.RefreshToken(refreshToken.Raw)).Exec(ctx)
			if err != nil {
				c.logger.Error("delete tokens error", zap.Error(err))
				return users.NewRefreshDefault(http.StatusInternalServerError).WithPayload(&models.Error{Data: &models.ErrorData{
					Message: "delete tokens error",
				}})
			}

			c.logger.Error("refresh token is expired", zap.Error(err))
			return users.NewRefreshDefault(http.StatusInternalServerError).WithPayload(&models.Error{Data: &models.ErrorData{
				Message: "refresh token is expired",
			}})
		}

		if err != nil {
			c.logger.Error("not valid refresh token", zap.Error(err))
			return users.NewRefreshDefault(http.StatusInternalServerError).WithPayload(&models.Error{Data: &models.ErrorData{
				Message: "not valid refresh token",
			}})
		}

		if refreshToken.Valid {
			if refreshToken.Raw != *p.RefreshToken.RefreshToken {
				c.logger.Error("invalid refresh token")
				return users.NewRefreshDefault(http.StatusInternalServerError).WithPayload(&models.Error{Data: &models.ErrorData{
					Message: "invalid refresh token",
				}})
			}

			userID := int(claims["id"].(float64))

			currentUser, err := c.client.User.Get(ctx, userID) // get current user
			if err != nil {
				c.logger.Error("user not found", zap.Error(err))
				return users.NewRefreshDefault(http.StatusInternalServerError).WithPayload(&models.Error{Data: &models.ErrorData{
					Message: "user not found",
				}})
			}

			newAccessToken, err := generateJWT(ctx, currentUser, jwtSecretKey)
			if err != nil {
				c.logger.Error("generate JWT token error")
				return users.NewRefreshDefault(http.StatusInternalServerError).WithPayload(&models.Error{Data: &models.ErrorData{
					Message: "generate JWT token error",
				}})
			}

			_, err = c.client.Token.Update().Where(token.RefreshToken(refreshToken.Raw)).SetAccessToken(newAccessToken).Save(ctx)
			if err != nil {
				log.Printf("update JWT token error: %v", err)
				return users.NewRefreshDefault(http.StatusInternalServerError).WithPayload(&models.Error{Data: &models.ErrorData{
					Message: "update JWT token error",
				}})
			}

			return users.NewRefreshOK().WithPayload(&models.AccessToken{AccessToken: &newAccessToken})
		}

		c.logger.Error("validating refresh token token error", zap.Error(err))
		return users.NewRefreshDefault(http.StatusInternalServerError).WithPayload(&models.Error{Data: &models.ErrorData{
			Message: "validating refresh token error",
		}})
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
		ctx := p.HTTPRequest.Context()
		id := int(p.UserID)
		user, err := c.client.User.Get(ctx, id)
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
		role, err := user.QueryRole().Only(ctx)
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
		ctx := p.HTTPRequest.Context()
		all, err := c.client.User.Query().All(ctx)
		if err != nil {
			c.logger.Error("failed to query users")
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

			role, err := element.QueryRole().Only(ctx)
			if err != nil {
				c.logger.Error("failed to query role")
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
