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
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/orders"
)

type OrderStatusNameTestSuite struct {
	suite.Suite
	logger      *zap.Logger
	repo        *repomock.StatusNameRepository
	orderStatus *OrderStatus
}

func TestOrderStatusSuite(t *testing.T) {
	s := new(OrderStatusNameTestSuite)
	suite.Run(t, s)
}

func (s *OrderStatusNameTestSuite) SetupTest() {
	s.logger = zap.NewExample()
	s.repo = &repomock.StatusNameRepository{}
	s.orderStatus = NewOrderStatus(s.logger)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_GetAllStatusNames_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	count := 1
	statuses := make([]*ent.StatusName, count)
	id := 1
	statusName := "test status"
	statuses[0] = &ent.StatusName{
		ID:     id,
		Status: statusName,
	}

	s.repo.On("ListOfStatuses", ctx).Return(statuses, nil)
	handlerFunc := s.orderStatus.GetAllStatusNames(s.repo)
	data := orders.GetAllStatusNamesParams{
		HTTPRequest: &request,
	}
	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	expectedStatuses := make([]models.OrderStatusName, count)
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &expectedStatuses)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(id), *expectedStatuses[0].ID)
	assert.Equal(t, statusName, *expectedStatuses[0].Name)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_GetAllStatusNames_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	err := errors.New("test error")
	s.repo.On("ListOfStatuses", ctx).Return(nil, err)
	handlerFunc := s.orderStatus.GetAllStatusNames(s.repo)
	data := orders.GetAllStatusNamesParams{
		HTTPRequest: &request,
	}
	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_GetAllStatusNames_MapError() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	statuses := make([]*ent.StatusName, 1)
	statuses[0] = nil
	s.repo.On("ListOfStatuses", ctx).Return(statuses, nil)
	handlerFunc := s.orderStatus.GetAllStatusNames(s.repo)
	data := orders.GetAllStatusNamesParams{
		HTTPRequest: &request,
	}
	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
}
