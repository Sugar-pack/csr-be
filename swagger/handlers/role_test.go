package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/roles"
)

func TestSetRoleHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:rolehandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	SetRoleHandler(client, logger, api)
	assert.NotEmpty(t, api.RolesGetRolesHandler)
}

type RoleTestSuite struct {
	suite.Suite
	logger     *zap.Logger
	repository *repomock.RoleRepository
	handler    *Role
}

func TestRoleSuite(t *testing.T) {
	suite.Run(t, new(RoleTestSuite))
}

func (s *RoleTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.repository = &repomock.RoleRepository{}
	s.handler = NewRole(s.logger)
}

func (s *RoleTestSuite) TestRole_GetRoles_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := roles.GetRolesParams{
		HTTPRequest: &request,
	}
	err := errors.New("test")
	s.repository.On("GetRoles", ctx).Return(nil, err)

	handlerFunc := s.handler.GetRolesFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *RoleTestSuite) TestRole_GetRoles_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	data := roles.GetRolesParams{
		HTTPRequest: &request,
	}
	var rolesToReturn []*ent.Role
	roleToReturn := &ent.Role{
		ID: 2,
	}
	rolesToReturn = append(rolesToReturn, roleToReturn)
	s.repository.On("GetRoles", ctx).Return(rolesToReturn, nil)

	handlerFunc := s.handler.GetRolesFunc(s.repository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var responseRoles []models.Role
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseRoles)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(rolesToReturn), len(responseRoles))
	assert.Equal(t, roleToReturn.ID, int(*responseRoles[0].ID))
	s.repository.AssertExpectations(t)
}
