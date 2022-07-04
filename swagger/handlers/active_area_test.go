package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/active_areas"
)

type ActiveAreaTestSuite struct {
	suite.Suite
	logger     *zap.Logger
	repository *repomock.ActiveAreaRepository
	handler    *ActiveArea
}

func TestActiveAreaSuite(t *testing.T) {
	suite.Run(t, new(ActiveAreaTestSuite))
}

func (s *ActiveAreaTestSuite) SetupTest() {
	s.logger, _ = zap.NewDevelopment()
	s.repository = &repomock.ActiveAreaRepository{}
	s.handler = NewActiveArea(s.logger)
}

func (s *ActiveAreaTestSuite) TestActiveArea_GetActiveAreasFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.handler.GetActiveAreasFunc(s.repository)
	data := active_areas.GetAllActiveAreasParams{
		HTTPRequest: &request,
	}

	err := errors.New("some error")
	s.repository.On("AllActiveAreas", ctx).Return(nil, err)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.repository.AssertExpectations(t)
}

func (s *ActiveAreaTestSuite) TestActiveArea_GetActiveAreasFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	handlerFunc := s.handler.GetActiveAreasFunc(s.repository)
	data := active_areas.GetAllActiveAreasParams{
		HTTPRequest: &request,
	}

	var areas []*ent.ActiveArea
	areas = append(areas, &ent.ActiveArea{
		ID:   1,
		Name: "test",
	},
	)
	s.repository.On("AllActiveAreas", ctx).Return(areas, nil)

	access := "dummy access"
	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var responseAreas []models.ActiveArea
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseAreas)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(areas), len(responseAreas))
	assert.Equal(t, areas[0].ID, int(*responseAreas[0].ID))
	assert.Equal(t, areas[0].Name, *responseAreas[0].Name)
	s.repository.AssertExpectations(t)
}
