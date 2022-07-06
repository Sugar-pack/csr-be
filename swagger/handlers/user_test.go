package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/loads"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	servicemock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/services"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/users"
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
	tokenManager := &servicemock.TokenManager{}
	registrationConfirm := &servicemock.RegistrationConfirm{}
	SetUserHandler(client, logger, api, tokenManager, registrationConfirm)

	assert.NotEmpty(t, api.UsersLoginHandler)
	assert.NotEmpty(t, api.UsersRefreshHandler)
	assert.NotEmpty(t, api.UsersPostUserHandler)
	assert.NotEmpty(t, api.UsersGetCurrentUserHandler)
	assert.NotEmpty(t, api.UsersPatchUserHandler)
	assert.NotEmpty(t, api.UsersGetUserHandler)
	assert.NotEmpty(t, api.UsersGetAllUsersHandler)
	assert.NotEmpty(t, api.UsersAssignRoleToUserHandler)
}

type UserTestSuite struct {
	suite.Suite
	logger              *zap.Logger
	service             *servicemock.TokenManager
	user                *User
	userRepository      *repomock.UserRepository
	registrationConfirm *servicemock.RegistrationConfirm
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

func (s *UserTestSuite) SetupTest() {
	s.logger, _ = zap.NewDevelopment()
	s.service = &servicemock.TokenManager{}
	s.registrationConfirm = &servicemock.RegistrationConfirm{}
	s.userRepository = &repomock.UserRepository{}
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
	s.service.On("GenerateAccessToken", ctx, login, password).Return("", true, err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
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
	s.service.On("GenerateAccessToken", ctx, login, password).Return("", false, err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)
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
	s.service.On("GenerateAccessToken", ctx, login, password).Return("", false, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
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
	assert.Equal(t, http.StatusExpectationFailed, responseRecorder.Code)
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
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
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
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
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
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
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
	s.service.On("RefreshToken", ctx, token).Return("", true, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
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
	s.service.On("RefreshToken", ctx, token).Return("", false, err)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
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
	s.service.On("RefreshToken", ctx, token).Return(newToken, false, nil)

	resp := handlerFunc(data)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	responseToken := &models.AccessToken{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), responseToken)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, newToken, *responseToken.AccessToken)

	s.service.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUserFunc_AccessErr() {
	t := s.T()
	request := http.Request{}

	access := "definitely not an access"
	handlerFunc := s.user.GetUserFunc(s.userRepository)
	data := users.GetCurrentUserParams{
		HTTPRequest: &request,
	}
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUserFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	access := authentication.Auth{
		Id: 1,
	}
	handlerFunc := s.user.GetUserFunc(s.userRepository)
	data := users.GetCurrentUserParams{
		HTTPRequest: &request,
	}
	err := errors.New("some error")
	s.userRepository.On("GetUserByID", ctx, access.Id).Return(nil, err)

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUserFunc_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	access := authentication.Auth{
		Id: 1,
	}
	handlerFunc := s.user.GetUserFunc(s.userRepository)
	data := users.GetCurrentUserParams{
		HTTPRequest: &request,
	}
	user := &ent.User{}
	s.userRepository.On("GetUserByID", ctx, access.Id).Return(user, nil)

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUserFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	access := authentication.Auth{
		Id: 1,
	}
	handlerFunc := s.user.GetUserFunc(s.userRepository)
	data := users.GetCurrentUserParams{
		HTTPRequest: &request,
	}
	user := &ent.User{
		ID: access.Id,
		Edges: ent.UserEdges{
			Role: &ent.Role{},
		},
	}
	s.userRepository.On("GetUserByID", ctx, access.Id).Return(user, nil)

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_PatchUser_AccessErr() {
	t := s.T()
	request := http.Request{}

	patch := &models.PatchUserRequest{
		Name: "new name",
	}

	access := "definitely not an access"
	handlerFunc := s.user.PatchUserFunc(s.userRepository)
	data := users.PatchUserParams{
		HTTPRequest: &request,
		UserPatch:   patch,
	}

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_PatchUser_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	patch := &models.PatchUserRequest{
		Name: "new name",
	}

	access := authentication.Auth{Id: 1}
	handlerFunc := s.user.PatchUserFunc(s.userRepository)
	data := users.PatchUserParams{
		HTTPRequest: &request,
		UserPatch:   patch,
	}

	err := errors.New("some error")
	s.userRepository.On("UpdateUserByID", ctx, access.Id, patch).Return(err)

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_PatchUser_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	patch := &models.PatchUserRequest{
		Name: "new name",
	}

	access := authentication.Auth{Id: 1}
	handlerFunc := s.user.PatchUserFunc(s.userRepository)
	data := users.PatchUserParams{
		HTTPRequest: &request,
		UserPatch:   patch,
	}

	s.userRepository.On("UpdateUserByID", ctx, access.Id, patch).Return(nil)

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusNoContent, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_AssignRoleToUserFunc_AccessErr() {
	t := s.T()
	request := http.Request{}

	access := "definitely not an access"
	handlerFunc := s.user.AssignRoleToUserFunc(s.userRepository)
	data := users.AssignRoleToUserParams{
		HTTPRequest: &request,
		Data:        &models.AssignRoleToUser{},
	}

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusForbidden, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_AssignRoleToUserFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	access := authentication.Auth{
		Id: 1,
		Role: &authentication.Role{
			Slug: authentication.AdminSlug,
		},
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

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_AssignRoleToUserFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	access := authentication.Auth{
		Id: 1,
		Role: &authentication.Role{
			Slug: authentication.AdminSlug,
		},
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

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

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
	s.userRepository.On("UserList", ctx).Return(nil, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUsersList_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.user.GetUsersList(s.userRepository)
	data := users.GetAllUsersParams{
		HTTPRequest: &request,
	}
	var userList []*ent.User
	user := &ent.User{
		ID: 1,
	}
	userList = append(userList, user)
	s.userRepository.On("UserList", ctx).Return(userList, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

	s.userRepository.AssertExpectations(t)
}

func (s *UserTestSuite) TestUser_GetUsersList_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.user.GetUsersList(s.userRepository)
	data := users.GetAllUsersParams{
		HTTPRequest: &request,
	}
	var userList []*ent.User
	user := &ent.User{
		ID: 1,
		Edges: ent.UserEdges{
			Role: &ent.Role{},
		},
	}
	userList = append(userList, user)
	s.userRepository.On("UserList", ctx).Return(userList, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseUsers := &models.GetListUsers{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), responseUsers)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(userList), len(*responseUsers))
	assert.Equal(t, user.ID, int(*(*responseUsers)[0].ID))

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

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

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

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)

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

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseUsers := &models.UserInfo{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), responseUsers)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, user.ID, int(*responseUsers.ID))

	s.userRepository.AssertExpectations(t)
}
