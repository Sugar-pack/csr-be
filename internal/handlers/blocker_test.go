package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/users"
)

func TestSetBlockerHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:blockerhandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	SetBlockerHandler(logger, api)
	assert.NotEmpty(t, api.UsersBlockUserHandler)
	assert.NotEmpty(t, api.UsersUnblockUserHandler)
}

type BlockerTestSuite struct {
	suite.Suite
	logger            *zap.Logger
	blockerRepository *mocks.BlockerRepository
	blocker           *Blocker
}

func TestBlockerSuite(t *testing.T) {
	suite.Run(t, new(BlockerTestSuite))
}

func (s *BlockerTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.blockerRepository = &mocks.BlockerRepository{}
	s.blocker = NewBlocker(s.logger)
}

func (s *BlockerTestSuite) TestBlocker_BlockUserFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	data := users.BlockUserParams{
		HTTPRequest: &request,
		UserID:      int64(userID),
	}
	err := errors.New("test")
	s.blockerRepository.On("SetIsBlockedUser", ctx, userID, true).Return(err)

	handlerFunc := s.blocker.BlockUserFunc(s.blockerRepository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.blockerRepository.AssertExpectations(t)
}

func (s *BlockerTestSuite) TestBlocker_BlockUserFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	data := users.BlockUserParams{
		HTTPRequest: &request,
		UserID:      int64(userID),
	}

	s.blockerRepository.On("SetIsBlockedUser", ctx, userID, true).Return(nil)

	handlerFunc := s.blocker.BlockUserFunc(s.blockerRepository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.blockerRepository.AssertExpectations(t)
}

func (s *BlockerTestSuite) TestBlocker_UnblockUserFunc_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	data := users.UnblockUserParams{
		HTTPRequest: &request,
		UserID:      int64(userID),
	}
	err := errors.New("test")
	s.blockerRepository.On("SetIsBlockedUser", ctx, userID, false).Return(err)

	handlerFunc := s.blocker.UnblockUserFunc(s.blockerRepository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.blockerRepository.AssertExpectations(t)
}

func (s *BlockerTestSuite) TestBlocker_UnblockUserFunc_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	data := users.UnblockUserParams{
		HTTPRequest: &request,
		UserID:      int64(userID),
	}
	s.blockerRepository.On("SetIsBlockedUser", ctx, userID, false).Return(nil)

	handlerFunc := s.blocker.UnblockUserFunc(s.blockerRepository)
	access := "dummy access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.blockerRepository.AssertExpectations(t)
}
