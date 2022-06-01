package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gotest.tools/assert"

	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/users"
)

type BlockerTestSuite struct {
	suite.Suite
	logger            *zap.Logger
	blockerRepository *repomock.BlockerRepository
	blocker           *Blocker
}

func TestBlockerSuite(t *testing.T) {
	suite.Run(t, new(BlockerTestSuite))
}

func (s *BlockerTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.blockerRepository = &repomock.BlockerRepository{}
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
	resp := handlerFunc.Handle(data)

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
	resp := handlerFunc.Handle(data)

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
	resp := handlerFunc.Handle(data)

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
	resp := handlerFunc.Handle(data)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.blockerRepository.AssertExpectations(t)
}
