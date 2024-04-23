package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/roles"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
)

func TestSetUserHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:userhandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	tokenManager := &mocks.TokenManager{}
	registrationConfirm := &mocks.RegistrationConfirmService{}
	SetUserHandler(logger, api, tokenManager, registrationConfirm, nil)

	require.NotEmpty(t, api.UsersLoginHandler)
	require.NotEmpty(t, api.UsersRefreshHandler)
	require.NotEmpty(t, api.UsersPostUserHandler)
	require.NotEmpty(t, api.UsersGetCurrentUserHandler)
	require.NotEmpty(t, api.UsersPatchUserHandler)
	require.NotEmpty(t, api.UsersGetUserHandler)
	require.NotEmpty(t, api.UsersGetAllUsersHandler)
	require.NotEmpty(t, api.UsersAssignRoleToUserHandler)
	require.NotEmpty(t, api.UsersDeleteCurrentUserHandler)
}

type UserTestSuite struct {
	suite.Suite
	logger              *zap.Logger
	service             *mocks.TokenManager
	user                *User
	userRepository      *mocks.UserRepository
	changeEmailService  *mocks.ChangeEmailService
	registrationConfirm *mocks.RegistrationConfirmService
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

func (s *UserTestSuite) SetupTest() {
	s.logger, _ = zap.NewDevelopment()
	s.service = &mocks.TokenManager{}
	s.registrationConfirm = &mocks.RegistrationConfirmService{}
	s.userRepository = &mocks.UserRepository{}
	s.changeEmailService = &mocks.ChangeEmailService{}
	s.user = NewUser(s.logger)
}

func (s *UserTestSuite) TestUser_LoginUserFunc_InternalErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.LoginUserFunc(s.service)
	data := users.LoginParams{
		HTTPRequest: &request,
		Login: &models.LoginInfo{
			Login:    &login,
			Password: &password,
		},
	}
	err := errors.New("some error")
	s.service.On("GenerateTokens", ctx, login, password).Return("", "", true, err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.service.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_LoginUserFunc_UnauthorizedErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.LoginUserFunc(s.service)
	data := users.LoginParams{
		HTTPRequest: &request,
		Login: &models.LoginInfo{
			Login:    &login,
			Password: &password,
		},
	}
	err := errors.New("some error")
	s.service.On("GenerateTokens", ctx, login, password).Return("", "", false, err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusUnauthorized, responseRecorder.Code)
	s.service.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_LoginUserFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.LoginUserFunc(s.service)
	data := users.LoginParams{
		HTTPRequest: &request,
		Login: &models.LoginInfo{
			Login:    &login,
			Password: &password,
		},
	}
	accessToken := "accessToken"
	refreshToken := "refreshToken"
	s.service.On("GenerateTokens", ctx, login, password).Return(accessToken, refreshToken, false, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var tokenPair models.TokenPair
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &tokenPair)
	if err != nil {
		t.Errorf("unable to unmarshal response: %v", err)
	}
	require.Equal(t, accessToken, *tokenPair.AccessToken)
	require.Equal(t, refreshToken, *tokenPair.RefreshToken)

	s.service.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_PostUserFunc_LoginExistErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.PostUserFunc(s.userRepository, s.registrationConfirm)
	data := users.PostUserParams{
		HTTPRequest: &request,
		Data: &models.UserRegister{
			Login:    &login,
			Password: &password,
		},
	}
	err := &ent.ConstraintError{}
	s.userRepository.On("CreateUser", ctx, data.Data).Return(nil, err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusExpectationFailed, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_PostUserFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.PostUserFunc(s.userRepository, s.registrationConfirm)
	data := users.PostUserParams{
		HTTPRequest: &request,
		Data: &models.UserRegister{
			Login:    &login,
			Password: &password,
		},
	}
	err := errors.New("some error")
	s.userRepository.On("CreateUser", ctx, data.Data).Return(nil, err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_PostUserFunc_RegConfirmErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.PostUserFunc(s.userRepository, s.registrationConfirm)
	data := users.PostUserParams{
		HTTPRequest: &request,
		Data: &models.UserRegister{
			Login:    &login,
			Password: &password,
		},
	}
	user := &ent.User{
		ID:    1,
		Login: login,
	}
	s.userRepository.On("CreateUser", ctx, data.Data).Return(user, nil)
	err := errors.New("some error")
	s.registrationConfirm.On("SendConfirmationLink", ctx, login).Return(err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusCreated, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_PostUserFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	login := "login"
	password := "password"
	handlerFunc := s.user.PostUserFunc(s.userRepository, s.registrationConfirm)
	data := users.PostUserParams{
		HTTPRequest: &request,
		Data: &models.UserRegister{
			Login:    &login,
			Password: &password,
		},
	}
	user := &ent.User{
		ID:    1,
		Login: login,
	}
	s.userRepository.On("CreateUser", ctx, data.Data).Return(user, nil)
	s.registrationConfirm.On("SendConfirmationLink", ctx, login).Return(nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusCreated, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_Refresh_InvalidToken_InvalidToken() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	token := "token"
	handlerFunc := s.user.Refresh(s.service)
	data := users.RefreshParams{
		HTTPRequest: &request,
		RefreshToken: &models.RefreshToken{
			RefreshToken: &token,
		},
	}
	s.service.On("RefreshToken", ctx, token).Return("", "", true, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	s.service.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_Refresh_InvalidToken_ServiceErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	token := "token"
	handlerFunc := s.user.Refresh(s.service)
	data := users.RefreshParams{
		HTTPRequest: &request,
		RefreshToken: &models.RefreshToken{
			RefreshToken: &token,
		},
	}
	err := errors.New("test error")
	s.service.On("RefreshToken", ctx, token).Return("", "", false, err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.service.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_Refresh_InvalidToken_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	token := "token"
	handlerFunc := s.user.Refresh(s.service)
	data := users.RefreshParams{
		HTTPRequest: &request,
		RefreshToken: &models.RefreshToken{
			RefreshToken: &token,
		},
	}
	newToken := "new token"
	s.service.On("RefreshToken", ctx, token).Return(newToken, newToken, false, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	responseToken := &models.AccessToken{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), responseToken)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, newToken, *responseToken.AccessToken)

	s.service.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUserFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	principal := &models.Principal{
		ID: int64(userID),
	}

	handlerFunc := s.user.GetUserFunc(s.userRepository)
	data := users.GetCurrentUserParams{
		HTTPRequest: &request,
	}
	err := errors.New("some error")
	s.userRepository.On("GetUserByID", ctx, userID).Return(nil, err)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUserFunc_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	principal := &models.Principal{
		ID: int64(userID),
	}

	handlerFunc := s.user.GetUserFunc(s.userRepository)
	data := users.GetCurrentUserParams{
		HTTPRequest: &request,
	}
	user := &ent.User{}
	s.userRepository.On("GetUserByID", ctx, userID).Return(user, nil)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUserFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	principal := &models.Principal{
		ID: int64(userID),
	}

	handlerFunc := s.user.GetUserFunc(s.userRepository)
	data := users.GetCurrentUserParams{
		HTTPRequest: &request,
	}
	user := &ent.User{
		ID: userID,
		Edges: ent.UserEdges{
			Role: &ent.Role{},
		},
	}
	s.userRepository.On("GetUserByID", ctx, userID).Return(user, nil)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_PatchUser_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	patch := &models.PatchUserRequest{
		Name: "new name",
	}
	userID := 1
	principal := &models.Principal{
		ID: int64(userID),
	}

	handlerFunc := s.user.PatchUserFunc(s.userRepository)
	data := users.PatchUserParams{
		HTTPRequest: &request,
		UserPatch:   patch,
	}

	err := errors.New("some error")
	s.userRepository.On("UpdateUserByID", ctx, userID, patch).Return(err)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_PatchUser_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	patch := &models.PatchUserRequest{
		Name: "new name",
	}
	userID := 1
	principal := &models.Principal{
		ID: int64(userID),
	}

	handlerFunc := s.user.PatchUserFunc(s.userRepository)
	data := users.PatchUserParams{
		HTTPRequest: &request,
		UserPatch:   patch,
	}

	s.userRepository.On("UpdateUserByID", ctx, userID, patch).Return(nil)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusNoContent, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_AssignRoleToUserFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	principal := &models.Principal{
		ID:   int64(userID),
		Role: roles.Admin,
	}

	handlerFunc := s.user.AssignRoleToUserFunc(s.userRepository)
	roleID := int64(1)
	userToChangeID := int64(1)
	data := users.AssignRoleToUserParams{
		HTTPRequest: &request,
		Data: &models.AssignRoleToUser{
			RoleID: &roleID,
		},
		UserID: userToChangeID,
	}
	err := errors.New("some error")
	s.userRepository.On("SetUserRole", ctx, int(userToChangeID), int(roleID)).Return(err)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_AssignRoleToUserFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	principal := &models.Principal{
		ID:   int64(userID),
		Role: roles.Admin,
	}

	handlerFunc := s.user.AssignRoleToUserFunc(s.userRepository)
	roleID := int64(1)
	userToChangeID := int64(1)
	data := users.AssignRoleToUserParams{
		HTTPRequest: &request,
		Data: &models.AssignRoleToUser{
			RoleID: &roleID,
		},
		UserID: userToChangeID,
	}
	s.userRepository.On("SetUserRole", ctx, int(userToChangeID), int(roleID)).Return(nil)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUsersList_RepositoryErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.user.GetUsersList(s.userRepository)
	data := users.GetAllUsersParams{
		HTTPRequest: &request,
	}
	err := errors.New("some err")
	s.userRepository.On("UsersListTotal", ctx).Return(0, err)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUsersList_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := user.FieldID

	handlerFunc := s.user.GetUsersList(s.userRepository)
	data := users.GetAllUsersParams{
		HTTPRequest: &request,
	}
	usersToReturn := []*ent.User{
		{
			ID: 1,
		},
	}
	s.userRepository.On("UsersListTotal", ctx).Return(1, nil)
	s.userRepository.On("UserList", ctx, limit, offset, orderBy, orderColumn).
		Return(usersToReturn, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUsersList_NotFound() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.user.GetUsersList(s.userRepository)
	data := users.GetAllUsersParams{
		HTTPRequest: &request,
	}
	s.userRepository.On("UsersListTotal", ctx).Return(0, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	responseUsers := &models.GetListUsers{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), responseUsers)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 0, int(*responseUsers.Total))
	require.Equal(t, 0, len(responseUsers.Items))
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUsersList_EmptyParams() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := user.FieldID

	handlerFunc := s.user.GetUsersList(s.userRepository)
	data := users.GetAllUsersParams{
		HTTPRequest: &request,
	}
	usersToReturn := []*ent.User{
		validUser(t, 1),
	}
	s.userRepository.On("UsersListTotal", ctx).Return(1, nil)
	s.userRepository.On("UserList", ctx, limit, offset, orderBy, orderColumn).
		Return(usersToReturn, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	responseUsers := &models.GetListUsers{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), responseUsers)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(usersToReturn), int(*responseUsers.Total))
	require.Equal(t, len(usersToReturn), len(responseUsers.Items))

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUsersList_LimitGreaterThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := int64(5)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := user.FieldID

	handlerFunc := s.user.GetUsersList(s.userRepository)
	data := users.GetAllUsersParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderColumn: &orderColumn,
		OrderBy:     &orderBy,
	}
	usersToReturn := []*ent.User{
		validUser(t, 1),
		validUser(t, 2),
		validUser(t, 3),
	}
	s.userRepository.On("UsersListTotal", ctx).Return(len(usersToReturn), nil)
	s.userRepository.On("UserList", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(usersToReturn, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	responseUsers := &models.GetListUsers{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), responseUsers)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(usersToReturn), int(*responseUsers.Total))
	require.Equal(t, len(usersToReturn), len(responseUsers.Items))
	require.GreaterOrEqual(t, int(limit), len(responseUsers.Items))

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUsersList_LimitLessThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := int64(3)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := user.FieldID

	handlerFunc := s.user.GetUsersList(s.userRepository)
	data := users.GetAllUsersParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderColumn: &orderColumn,
		OrderBy:     &orderBy,
	}
	usersToReturn := []*ent.User{
		validUser(t, 1),
		validUser(t, 2),
		validUser(t, 3),
		validUser(t, 4),
		validUser(t, 5),
		validUser(t, 6),
	}
	s.userRepository.On("UsersListTotal", ctx).Return(len(usersToReturn), nil)
	s.userRepository.On("UserList", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(usersToReturn[:limit], nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	responseUsers := &models.GetListUsers{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), responseUsers)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(usersToReturn), int(*responseUsers.Total))
	require.Greater(t, len(usersToReturn), len(responseUsers.Items))
	require.GreaterOrEqual(t, int(limit), len(responseUsers.Items))

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUsersList_SecondPage() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := int64(5)
	offset := int64(5)
	orderBy := utils.AscOrder
	orderColumn := user.FieldID

	handlerFunc := s.user.GetUsersList(s.userRepository)
	data := users.GetAllUsersParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderColumn: &orderColumn,
		OrderBy:     &orderBy,
	}
	usersToReturn := []*ent.User{
		validUser(t, 1),
		validUser(t, 2),
		validUser(t, 3),
		validUser(t, 4),
		validUser(t, 5),
		validUser(t, 6),
	}
	s.userRepository.On("UsersListTotal", ctx).Return(len(usersToReturn), nil)
	s.userRepository.On("UserList", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(usersToReturn[offset:], nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	responseUsers := &models.GetListUsers{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), responseUsers)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(usersToReturn), int(*responseUsers.Total))
	require.Greater(t, len(usersToReturn), len(responseUsers.Items))
	require.GreaterOrEqual(t, int(limit), len(responseUsers.Items))
	require.Equal(t, len(usersToReturn)-int(offset), len(responseUsers.Items))

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUsersList_SeveralPages() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	limit := int64(5)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := user.FieldID

	handlerFunc := s.user.GetUsersList(s.userRepository)
	data := users.GetAllUsersParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderColumn: &orderColumn,
		OrderBy:     &orderBy,
	}
	usersToReturn := []*ent.User{
		validUser(t, 1),
		validUser(t, 2),
		validUser(t, 3),
		validUser(t, 4),
		validUser(t, 5),
		validUser(t, 6),
	}
	s.userRepository.On("UsersListTotal", ctx).Return(len(usersToReturn), nil)
	s.userRepository.On("UserList", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(usersToReturn[:limit], nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	firstPage := &models.GetListUsers{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), firstPage)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(usersToReturn), int(*firstPage.Total))
	require.Equal(t, int(limit), len(firstPage.Items))

	offset = limit
	data = users.GetAllUsersParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderColumn: &orderColumn,
		OrderBy:     &orderBy,
	}
	s.userRepository.On("UsersListTotal", ctx).Return(len(usersToReturn), nil)
	s.userRepository.On("UserList", ctx, int(limit), int(offset), orderBy, orderColumn).
		Return(usersToReturn[offset:], nil)

	resp = handlerFunc(data, nil)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	secondPage := &models.GetListUsers{}
	err = json.Unmarshal(responseRecorder.Body.Bytes(), secondPage)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(usersToReturn), int(*secondPage.Total))
	require.Greater(t, len(usersToReturn), len(secondPage.Items))
	require.GreaterOrEqual(t, int(limit), len(secondPage.Items))
	require.Equal(t, len(usersToReturn)-int(offset), len(secondPage.Items))

	require.False(t, usersDuplicated(t, firstPage.Items, secondPage.Items))
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUserById_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.user.GetUserById(s.userRepository)
	userID := 1
	data := users.GetUserParams{
		HTTPRequest: &request,
		UserID:      int64(userID),
	}
	err := errors.New("some err")
	s.userRepository.On("GetUserByID", ctx, userID).Return(nil, err)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUserById_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.user.GetUserById(s.userRepository)
	userID := 1
	data := users.GetUserParams{
		HTTPRequest: &request,
		UserID:      int64(userID),
	}
	user := &ent.User{
		ID: 1,
	}
	s.userRepository.On("GetUserByID", ctx, userID).Return(user, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUserById_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.user.GetUserById(s.userRepository)
	userID := 1
	data := users.GetUserParams{
		HTTPRequest: &request,
		UserID:      int64(userID),
	}
	user := &ent.User{
		ID: 1,
		Edges: ent.UserEdges{
			Role: &ent.Role{},
		},
	}

	s.userRepository.On("GetUserByID", ctx, userID).Return(user, nil)

	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	responseUsers := &models.UserInfo{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), responseUsers)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, user.ID, int(*responseUsers.ID))

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_DeleteCurrentUserFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 3
	principal := &models.Principal{
		ID:   int64(userID),
		Role: roles.User,
	}

	handlerFunc := s.user.DeleteCurrentUser(s.userRepository)
	data := users.DeleteCurrentUserParams{
		HTTPRequest: &request,
	}

	s.userRepository.On("Delete", ctx, userID).Return(nil)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusOK, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_DeleteUser_OK() {
	t := s.T()

	ctx := context.Background()
	userID := 1232
	principal := &models.Principal{
		ID: int64(userID),
	}
	data := users.DeleteUserParams{
		HTTPRequest: &http.Request{},
		UserID:      int64(userID),
	}
	expectedUser := &ent.User{IsReadonly: true}

	s.userRepository.On("GetUserByID", ctx, userID).Return(expectedUser, nil)
	s.userRepository.On("Delete", ctx, userID).Return(nil)

	handlerFunc := s.user.DeleteUser(s.userRepository)
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusNoContent, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_DeleteUser_DeleteNonReadonlyUserError() {
	t := s.T()

	ctx := context.Background()
	userID := 1232
	principal := &models.Principal{
		ID: int64(userID),
	}
	data := users.DeleteUserParams{
		HTTPRequest: &http.Request{},
		UserID:      int64(userID),
	}
	expectedUser := &ent.User{IsReadonly: false}

	s.userRepository.On("GetUserByID", ctx, userID).Return(expectedUser, nil)

	handlerFunc := s.user.DeleteUser(s.userRepository)
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusForbidden, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_DeleteUser_GetUserByID_UserNotFound() {
	t := s.T()

	ctx := context.Background()
	userID := 1232
	principal := &models.Principal{}
	data := users.DeleteUserParams{
		HTTPRequest: &http.Request{},
		UserID:      int64(userID),
	}

	expectedError := &ent.NotFoundError{}
	s.userRepository.On("GetUserByID", ctx, userID).Return(nil, expectedError)

	handlerFunc := s.user.DeleteUser(s.userRepository)
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusNotFound, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_DeleteUser_GetUserByID_InternalError() {
	t := s.T()

	ctx := context.Background()
	userID := 1232
	principal := &models.Principal{}
	data := users.DeleteUserParams{
		HTTPRequest: &http.Request{},
		UserID:      int64(userID),
	}

	expectedError := fmt.Errorf("internal error")
	s.userRepository.On("GetUserByID", ctx, userID).Return(nil, expectedError)

	handlerFunc := s.user.DeleteUser(s.userRepository)
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_DeleteUser_Delete_UserNotFound() {
	t := s.T()

	ctx := context.Background()
	userID := 1232
	principal := &models.Principal{}
	data := users.DeleteUserParams{
		HTTPRequest: &http.Request{},
		UserID:      int64(userID),
	}

	expectedUser := &ent.User{IsReadonly: true}
	expectedError := &ent.NotFoundError{}

	s.userRepository.On("GetUserByID", ctx, userID).Return(expectedUser, nil)
	s.userRepository.On("Delete", ctx, userID).Return(expectedError)

	handlerFunc := s.user.DeleteUser(s.userRepository)
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusNotFound, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_DeleteUser_Delete_InternalError() {
	t := s.T()

	ctx := context.Background()
	userID := 1232
	principal := &models.Principal{}
	data := users.DeleteUserParams{
		HTTPRequest: &http.Request{},
		UserID:      int64(userID),
	}

	expectedUser := &ent.User{IsReadonly: true}
	expectedError := fmt.Errorf("internal error")

	s.userRepository.On("GetUserByID", ctx, userID).Return(expectedUser, nil)
	s.userRepository.On("Delete", ctx, userID).Return(expectedError)

	handlerFunc := s.user.DeleteUser(s.userRepository)
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_ChangePasswordFunc_GetUserErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	handlerFunc := s.user.ChangePassword(s.userRepository)

	id := 1
	user := validUser(t, id)
	password := "password"
	passwordHash, err := utils.PasswordHash(password)
	if err != nil {
		t.Fatal(err)
	}
	user.Password = passwordHash
	newPassword := "newPassword"

	data := users.ChangePasswordParams{
		HTTPRequest: &request,
		PasswordPatch: &models.PatchPasswordRequest{
			OldPassword: password,
			NewPassword: newPassword,
		},
	}
	principal := &models.Principal{
		ID:   int64(id),
		Role: roles.Admin,
	}

	err = errors.New("failed to get user")
	s.userRepository.On("GetUserByID", ctx, user.ID).Return(nil, err)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_ChangePasswordFunc_ComparePasswordErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	handlerFunc := s.user.ChangePassword(s.userRepository)

	id := 1
	user := validUser(t, id)
	password := "password"
	passwordHash, err := utils.PasswordHash(password)
	if err != nil {
		t.Fatal(err)
	}
	user.Password = passwordHash
	newPassword := "newPassword"
	nonValidPassword := "nonValidPassword"

	data := users.ChangePasswordParams{
		HTTPRequest: &request,
		PasswordPatch: &models.PatchPasswordRequest{
			OldPassword: nonValidPassword,
			NewPassword: newPassword,
		},
	}
	principal := &models.Principal{
		ID:   int64(id),
		Role: roles.Admin,
	}

	s.userRepository.On("GetUserByID", ctx, user.ID).Return(user, nil)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusForbidden, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_ChangePasswordFunc_ChangePasswordError() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	handlerFunc := s.user.ChangePassword(s.userRepository)

	id := 1
	user := validUser(t, id)
	password := "password"
	passwordHash, err := utils.PasswordHash(password)
	if err != nil {
		t.Fatal(err)
	}
	user.Password = passwordHash
	newPassword := "newPassword"

	data := users.ChangePasswordParams{
		HTTPRequest: &request,
		PasswordPatch: &models.PatchPasswordRequest{
			OldPassword: password,
			NewPassword: newPassword,
		},
	}
	principal := &models.Principal{
		ID:   int64(id),
		Role: roles.Admin,
	}

	err = errors.New("failed to change password")
	s.userRepository.On("GetUserByID", ctx, user.ID).Return(user, nil)
	s.userRepository.On("ChangePasswordByLogin", ctx, user.Login, newPassword).Return(err)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_ChangePasswordFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	handlerFunc := s.user.ChangePassword(s.userRepository)

	id := 1
	user := validUser(t, id)
	password := "password"
	passwordHash, err := utils.PasswordHash(password)
	if err != nil {
		t.Fatal(err)
	}
	user.Password = passwordHash
	newPassword := "newPassword"

	data := users.ChangePasswordParams{
		HTTPRequest: &request,
		PasswordPatch: &models.PatchPasswordRequest{
			OldPassword: password,
			NewPassword: newPassword,
		},
	}
	principal := &models.Principal{
		ID:   int64(id),
		Role: roles.Admin,
	}

	s.userRepository.On("GetUserByID", ctx, user.ID).Return(user, nil)
	s.userRepository.On("ChangePasswordByLogin", ctx, user.Login, newPassword).Return(nil)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusNoContent, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_ChangeEmailFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	handlerFunc := s.user.ChangeEmail(s.userRepository, s.changeEmailService)

	id := 1
	user := validUser(t, id)

	testEmail := "test@email1"

	data := users.ChangeEmailParams{
		HTTPRequest: &request,
		EmailPatch: &models.PatchEmailRequest{
			NewEmail: testEmail,
		},
	}

	principal := &models.Principal{
		ID:   int64(id),
		Role: roles.User,
	}

	s.userRepository.On("GetUserByID", ctx, user.ID).Return(user, nil)
	s.userRepository.On("UnConfirmRegistration", ctx, user.Login).Return(nil)
	s.changeEmailService.On("SendEmailConfirmationLink", ctx, user.Login, testEmail).Return(nil)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusNoContent, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_ChangeEmailFunc_Err() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	handlerFunc := s.user.ChangeEmail(s.userRepository, s.changeEmailService)

	id := 1
	user := validUser(t, id)

	testEmail := "test@email1"

	data := users.ChangeEmailParams{
		HTTPRequest: &request,
		EmailPatch: &models.PatchEmailRequest{
			NewEmail: testEmail,
		},
	}

	principal := &models.Principal{
		ID:   int64(id),
		Role: roles.User,
	}

	s.userRepository.On("GetUserByID", ctx, user.ID).Return(user, nil)
	s.userRepository.On("UnConfirmRegistration", ctx, user.Login).Return(nil)
	err := errors.New("unable to send email confirmation link")
	s.changeEmailService.On("SendEmailConfirmationLink", ctx, user.Login, testEmail).Return(err)

	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_UpdateReadonlyAccess_Grant() {
	t := s.T()

	ctx := context.Background()
	userID := 1232
	isReadonly := true
	data := users.UpdateReadonlyAccessParams{
		HTTPRequest: &http.Request{},
		UserID:      int64(userID),
		Body:        users.UpdateReadonlyAccessBody{IsReadonly: isReadonly},
	}

	s.userRepository.On("SetIsReadonly", ctx, userID, isReadonly).Return(nil)

	handlerFunc := s.user.UpdateReadonlyAccess(s.userRepository)

	principal := &models.Principal{}
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusNoContent, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_UpdateReadonlyAccess_Revoke() {
	t := s.T()

	ctx := context.Background()
	userID := 1232
	isReadonly := false
	data := users.UpdateReadonlyAccessParams{
		HTTPRequest: &http.Request{},
		UserID:      int64(userID),
		Body:        users.UpdateReadonlyAccessBody{IsReadonly: isReadonly},
	}

	s.userRepository.On("SetIsReadonly", ctx, userID, isReadonly).Return(nil)

	handlerFunc := s.user.UpdateReadonlyAccess(s.userRepository)
	principal := &models.Principal{}
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusNoContent, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_UpdateReadonlyAccess_UserNotFound() {
	t := s.T()

	ctx := context.Background()
	userID := 1232
	isReadonly := false
	data := users.UpdateReadonlyAccessParams{
		HTTPRequest: &http.Request{},
		UserID:      int64(userID),
		Body:        users.UpdateReadonlyAccessBody{IsReadonly: isReadonly},
	}

	expectedError := &ent.NotFoundError{}
	s.userRepository.On("SetIsReadonly", ctx, userID, isReadonly).Return(expectedError)

	handlerFunc := s.user.UpdateReadonlyAccess(s.userRepository)
	principal := &models.Principal{}
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusNotFound, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_UpdateReadonlyAccess_InternalError() {
	t := s.T()

	ctx := context.Background()
	userID := 1232
	isReadonly := false
	data := users.UpdateReadonlyAccessParams{
		HTTPRequest: &http.Request{},
		UserID:      int64(userID),
		Body:        users.UpdateReadonlyAccessBody{IsReadonly: isReadonly},
	}

	expectedError := fmt.Errorf("internal error")
	s.userRepository.On("SetIsReadonly", ctx, userID, isReadonly).Return(expectedError)

	handlerFunc := s.user.UpdateReadonlyAccess(s.userRepository)
	principal := &models.Principal{}
	resp := handlerFunc(data, principal)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func validUser(t *testing.T, id int) *ent.User {
	t.Helper()
	return &ent.User{
		ID:    id,
		Name:  fmt.Sprintf("User%d", id),
		Login: fmt.Sprintf("user_%d", id),
		Email: fmt.Sprintf("user_%d@mail.com", id),
		Edges: ent.UserEdges{
			Role: &ent.Role{
				ID:   1,
				Name: "user",
				Slug: "user",
			},
		},
	}
}

func usersDuplicated(t *testing.T, array1, array2 []*models.UserInfo) bool {
	t.Helper()
	diff := make(map[int64]int, len(array1))
	for _, v := range array1 {
		diff[*v.ID] = 1
	}
	for _, v := range array2 {
		if _, ok := diff[*v.ID]; ok {
			return true
		}
	}
	return false
}
