package handlers

import (
	"errors"
	"math"
	"net/http"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/users"
)

func SetUserHandler(logger *zap.Logger, api *operations.BeAPI,
	tokenManager domain.TokenManager, regConfirmService domain.RegistrationConfirmService) {
	userRepo := repositories.NewUserRepository()
	userHandler := NewUser(logger)

	api.UsersLoginHandler = userHandler.LoginUserFunc(tokenManager)
	api.UsersRefreshHandler = userHandler.Refresh(tokenManager)
	api.UsersPostUserHandler = userHandler.PostUserFunc(userRepo, regConfirmService)
	api.UsersGetCurrentUserHandler = userHandler.GetUserFunc(userRepo)
	api.UsersPatchUserHandler = userHandler.PatchUserFunc(userRepo)
	api.UsersGetUserHandler = userHandler.GetUserById(userRepo)
	api.UsersGetAllUsersHandler = userHandler.GetUsersList(userRepo)
	api.UsersAssignRoleToUserHandler = userHandler.AssignRoleToUserFunc(userRepo)
	api.UsersDeleteUserHandler = userHandler.DeleteUserByID(userRepo)
}

type User struct {
	logger *zap.Logger
}

func NewUser(logger *zap.Logger) *User {
	return &User{
		logger: logger,
	}
}

func (c User) LoginUserFunc(service domain.TokenManager) users.LoginHandlerFunc {
	return func(p users.LoginParams) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		login := *p.Login.Login
		password := *p.Login.Password
		accessToken, refreshToken, isInternalErr, err := service.GenerateTokens(ctx, login, password)
		if err != nil {
			if isInternalErr {
				return users.NewLoginDefault(http.StatusInternalServerError)
			}
			return users.NewLoginUnauthorized().WithPayload("Invalid login or password")
		}

		return users.NewLoginOK().WithPayload(&models.TokenPair{
			AccessToken:  &accessToken,
			RefreshToken: &refreshToken,
		})
	}
}

func (c User) PostUserFunc(repository domain.UserRepository, regConfirmService domain.RegistrationConfirmService) users.PostUserHandlerFunc {
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

func (c User) Refresh(manager domain.TokenManager) users.RefreshHandlerFunc {
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

func (c User) GetUserFunc(repository domain.UserRepository) users.GetCurrentUserHandlerFunc {
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

func (c User) PatchUserFunc(repository domain.UserRepository) users.PatchUserHandlerFunc {
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

func (c User) AssignRoleToUserFunc(repository domain.UserRepository) users.AssignRoleToUserHandlerFunc {
	return func(p users.AssignRoleToUserParams, access interface{}) middleware.Responder {
		isAdmin, err := authentication.IsAdmin(access)
		if err != nil {
			c.logger.Error("error while getting authorization", zap.Error(err))
			return users.NewAssignRoleToUserDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't get authorization"}})
		}
		if !isAdmin {
			c.logger.Error("user is not admin", zap.Any("access", access))
			return users.NewAssignRoleToUserDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "You don't have rights to add new status"}})
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

func (c User) GetUserById(repository domain.UserRepository) users.GetUserHandlerFunc {
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

func (c User) GetUsersList(repository domain.UserRepository) users.GetAllUsersHandlerFunc {
	return func(p users.GetAllUsersParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		limit := utils.GetValueByPointerOrDefaultValue(p.Limit, math.MaxInt)
		offset := utils.GetValueByPointerOrDefaultValue(p.Offset, 0)
		orderBy := utils.GetValueByPointerOrDefaultValue(p.OrderBy, utils.AscOrder)
		orderColumn := utils.GetValueByPointerOrDefaultValue(p.OrderColumn, user.FieldID)
		total, err := repository.UsersListTotal(ctx)
		if err != nil {
			c.logger.Error("failed get user total amount", zap.Error(err))
			return users.NewGetAllUsersDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("failed to get user total amount"))
		}
		var allUsers []*ent.User
		if total > 0 {
			allUsers, err = repository.UserList(ctx, int(limit), int(offset), orderBy, orderColumn)
			if err != nil {
				c.logger.Error("failed get user list", zap.Error(err))
				return users.NewGetAllUsersDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("failed to get user list"))
			}
		}

		usersToResponse := make([]*models.UserInfo, len(allUsers))
		for i, element := range allUsers {
			userToResponse, errMap := mapUserInfo(element)
			if errMap != nil {
				c.logger.Error("map user error", zap.Error(errMap))
				return users.NewGetAllUsersDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("map user error"))
			}
			usersToResponse[i] = userToResponse
		}
		totalUsers := int64(total)
		listUsers := &models.GetListUsers{
			Items: usersToResponse,
			Total: &totalUsers,
		}

		return users.NewGetAllUsersOK().WithPayload(listUsers)
	}
}

func (c User) DeleteUserByID(repo domain.UserRepository) users.DeleteUserHandlerFunc {
	return func(p users.DeleteUserParams, access interface{}) middleware.Responder {
		// #todo: add test

		isAdmin, err := authentication.IsAdmin(access)
		if err != nil {
			c.logger.Error("error while getting authorization", zap.Error(err))
			return users.NewDeleteUserDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't get authorization"}})
		}
		if !isAdmin {
			c.logger.Error("user is not admin", zap.Any("access", access))
			return users.NewDeleteUserDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "You don't have rights"}})
		}

		ctx := p.HTTPRequest.Context()
		userToDelete, err := repo.GetUserByID(ctx, int(p.UserID))
		if err != nil {
			c.logger.Error("getting user failed", zap.Error(err))
			return users.NewDeleteUserDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Can't get user by id"))
		}

		if userToDelete.IsBlocked != true {
			c.logger.Error("user must be blocked before delete", zap.Any("access", access))
			return users.NewDeleteUserDefault(http.StatusConflict).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "User must be blocked before delete"}})
		}

		err = repo.Delete(ctx, int(p.UserID))
		if err != nil {
			c.logger.Error("Error while deleting user by id", zap.Error(err))
			return users.NewDeleteUserDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "Error while deleting user",
					},
				})
		}
		return users.NewDeleteUserOK().WithPayload("User deleted")
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
