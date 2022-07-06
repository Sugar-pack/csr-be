package handlers

import (
	"errors"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/services"
)

func SetUserHandler(client *ent.Client, logger *zap.Logger, api *operations.BeAPI,
	tokenManager services.TokenManager, regConfirmService services.RegistrationConfirm) {
	userRepo := repositories.NewUserRepository(client)
	userHandler := NewUser(logger)

	api.UsersLoginHandler = userHandler.LoginUserFunc(tokenManager)
	api.UsersRefreshHandler = userHandler.Refresh(tokenManager)
	api.UsersPostUserHandler = userHandler.PostUserFunc(userRepo, regConfirmService)
	api.UsersGetCurrentUserHandler = userHandler.GetUserFunc(userRepo)
	api.UsersPatchUserHandler = userHandler.PatchUserFunc(userRepo)
	api.UsersGetUserHandler = userHandler.GetUserById(userRepo)
	api.UsersGetAllUsersHandler = userHandler.GetUsersList(userRepo)
	api.UsersAssignRoleToUserHandler = userHandler.AssignRoleToUserFunc(userRepo)
}

type User struct {
	logger *zap.Logger
}

func NewUser(logger *zap.Logger) *User {
	return &User{
		logger: logger,
	}
}

func (c User) LoginUserFunc(service services.TokenManager) users.LoginHandlerFunc {
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

func (c User) PostUserFunc(repository repositories.UserRepository, regConfirmService services.RegistrationConfirm) users.PostUserHandlerFunc {
	return func(p users.PostUserParams) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		createdUser, err := repository.CreateUser(ctx, p.Data)
		if err != nil {
			if ent.IsConstraintError(err) {
				return users.NewPostUserDefault(http.StatusExpectationFailed).WithPayload(
					buildErrorPayload(errors.New("login is already used")),
				)
			}
			return users.NewPostUserDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		id := int64(createdUser.ID)

		err = regConfirmService.SendConfirmationLink(ctx, createdUser.Login)
		if err != nil {
			c.logger.Error("error sending registration confirmation link", zap.Error(err))
		}

		return users.NewPostUserCreated().WithPayload(&models.CreateUserResponse{
			Data: &models.CreateUserResponseData{
				ID:    &id,
				Login: &createdUser.Login,
			},
		})
	}
}

func (c User) Refresh(manager services.TokenManager) users.RefreshHandlerFunc {
	return func(p users.RefreshParams) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		refreshToken := *p.RefreshToken.RefreshToken
		newToken, isValid, err := manager.RefreshToken(ctx, refreshToken)
		if isValid {
			c.logger.Info("token invalid", zap.String("token", refreshToken))
			return users.NewRefreshDefault(http.StatusBadRequest).
				WithPayload(buildStringPayload("token invalid"))
		}
		if err != nil {
			c.logger.Error("Error while refreshing token", zap.Error(err))
			return users.NewRefreshDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while refreshing token"))
		}
		return users.NewRefreshOK().WithPayload(&models.AccessToken{AccessToken: &newToken})
	}
}

func (c User) GetUserFunc(repository repositories.UserRepository) users.GetCurrentUserHandlerFunc {
	return func(p users.GetCurrentUserParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		userId, err := authentication.GetUserId(access)
		if err != nil {
			c.logger.Error("get user id error", zap.Error(err))
			return users.NewGetCurrentUserDefault(http.StatusUnauthorized).WithPayload(&models.Error{Data: &models.ErrorData{
				Message: "get user id error",
			}})
		}
		user, err := repository.GetUserByID(ctx, userId)
		if err != nil {
			c.logger.Error("get user by id error", zap.Error(err))
			return users.NewGetCurrentUserDefault(http.StatusInternalServerError).WithPayload(&models.Error{Data: &models.ErrorData{
				Message: "cant find user by id",
			}})
		}

		result, err := mapUserInfo(user)
		if err != nil {
			c.logger.Error("map user error", zap.Error(err))
			return users.NewGetCurrentUserDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("map user error"))
		}

		return users.NewGetCurrentUserOK().WithPayload(result)
	}
}

func (c User) PatchUserFunc(repository repositories.UserRepository) users.PatchUserHandlerFunc {
	return func(p users.PatchUserParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		userId, err := authentication.GetUserId(access)
		if err != nil {
			c.logger.Error("get user id error", zap.Error(err))
			return users.NewPatchUserDefault(http.StatusUnauthorized).
				WithPayload(buildStringPayload("get user id error"))
		}
		err = repository.UpdateUserByID(ctx, userId, p.UserPatch)
		if err != nil {
			c.logger.Error("get user by id error", zap.Error(err))
			return users.NewPatchUserDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("cant update user"))
		}
		return users.NewPatchUserNoContent()
	}
}

func (c User) AssignRoleToUserFunc(repository repositories.UserRepository) users.AssignRoleToUserHandlerFunc {
	return func(p users.AssignRoleToUserParams, access interface{}) middleware.Responder {
		_, err := authentication.IsAdmin(access)
		if err != nil {
			c.logger.Error("user is not admin", zap.Error(err))
			return users.NewAssignRoleToUserDefault(http.StatusForbidden).WithPayload(buildErrorPayload(err))
		}
		ctx := p.HTTPRequest.Context()
		userId := int(p.UserID)
		roleId := int(*p.Data.RoleID)

		err = repository.SetUserRole(ctx, userId, roleId)
		if err != nil {
			c.logger.Error("set user role error", zap.Error(err))
			return users.NewAssignRoleToUserDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		return users.NewAssignRoleToUserOK().WithPayload("role assigned")
	}
}

func (c User) GetUserById(repository repositories.UserRepository) users.GetUserHandlerFunc {
	return func(p users.GetUserParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		id := int(p.UserID)
		foundUser, err := repository.GetUserByID(ctx, id)
		if err != nil {
			return users.NewGetUserDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("cant find user by id"))
		}

		userToResponse, err := mapUserInfo(foundUser)
		if err != nil {
			c.logger.Error("map user error", zap.Error(err))
			return users.NewGetUserDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("map user error"))
		}

		return users.NewGetUserOK().WithPayload(userToResponse)
	}
}

func (c User) GetUsersList(repository repositories.UserRepository) users.GetAllUsersHandlerFunc {
	return func(p users.GetAllUsersParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		all, err := repository.UserList(ctx)
		if err != nil {
			c.logger.Error("failed get user list", zap.Error(err))
			return users.NewGetAllUsersDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("failed to get user list"))
		}
		listUsers := models.GetListUsers{}
		for _, element := range all {
			userToResponse, errMap := mapUserInfo(element)
			if errMap != nil {
				c.logger.Error("map user error", zap.Error(errMap))
				return users.NewGetAllUsersDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("map user error"))
			}
			listUsers = append(listUsers, userToResponse)
		}

		return users.NewGetAllUsersOK().WithPayload(listUsers)
	}
}

func mapUserInfo(user *ent.User) (*models.UserInfo, error) {
	userID := int64(user.ID)
	passportDate := user.PassportIssueDate.String()
	if user.Edges.Role == nil {
		return nil, errors.New("role is nil")
	}
	userRole := user.Edges.Role
	userRoleInfo := models.UserInfoRole{
		ID:   int64(userRole.ID),
		Name: userRole.Name,
	}
	typeString := user.Type.String()
	result := &models.UserInfo{
		Email:             &user.Email,
		ID:                &userID,
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
		Role:              &userRoleInfo,
		Surname:           user.Surname,
		Type:              &typeString,
	}
	return result, nil
}
