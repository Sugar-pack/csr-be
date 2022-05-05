package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/orders"
)

type OrderStatusNameTestSuite struct {
	suite.Suite
	logger                *zap.Logger
	statusNameRepository  *repomock.StatusNameRepository
	orderStatusRepository *repomock.OrderStatusRepository
	orderFilterRepository *repomock.OrderRepositoryWithFilter
	orderStatus           *OrderStatus
}

func orderWithEdges(t *testing.T) ent.Order {
	t.Helper()
	return ent.Order{
		ID:          1,
		Description: "test description",
		Quantity:    1,
		RentStart:   time.Now().UTC(),
		RentEnd:     time.Now().UTC(),
		CreatedAt:   time.Now().UTC(),
		Edges: ent.OrderEdges{
			Users: []*ent.User{
				{
					ID:    1,
					Login: "login",
				},
			},
			Equipments: []*ent.Equipment{
				{
					Description: "description",
				},
			},
			OrderStatus: []*ent.OrderStatus{
				{
					ID: 1,
					Edges: ent.OrderStatusEdges{
						StatusName: &ent.StatusName{
							ID: 1,
						},
						Users: &ent.User{
							ID: 1,
						},
					},
				},
			},
		},
	}
}

func TestOrderStatusSuite(t *testing.T) {
	s := new(OrderStatusNameTestSuite)
	suite.Run(t, s)
}

func (s *OrderStatusNameTestSuite) SetupTest() {
	s.logger = zap.NewExample()
	s.statusNameRepository = &repomock.StatusNameRepository{}
	s.orderStatusRepository = &repomock.OrderStatusRepository{}
	s.orderFilterRepository = &repomock.OrderRepositoryWithFilter{}
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

	s.statusNameRepository.On("ListOfStatuses", ctx).Return(statuses, nil)
	handlerFunc := s.orderStatus.GetAllStatusNames(s.statusNameRepository)
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
	s.statusNameRepository.On("ListOfStatuses", ctx).Return(nil, err)
	handlerFunc := s.orderStatus.GetAllStatusNames(s.statusNameRepository)
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
	s.statusNameRepository.On("ListOfStatuses", ctx).Return(statuses, nil)
	handlerFunc := s.orderStatus.GetAllStatusNames(s.statusNameRepository)
	data := orders.GetAllStatusNamesParams{
		HTTPRequest: &request,
	}
	resp := handlerFunc(data, nil)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_OrderStatusesHistory_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	access := "definitely not access"
	handlerFunc := s.orderStatus.OrderStatusesHistory(s.orderStatusRepository)
	orderID := int64(1)
	data := orders.GetFullOrderHistoryParams{
		HTTPRequest: &request,
		OrderID:     orderID,
	}
	err := errors.New("test error")
	s.orderStatusRepository.On("StatusHistory", ctx, int(orderID)).Return(nil, err)

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderStatusRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_OrderStatusesHistory_CantAccess() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	access := "definitely not access"
	handlerFunc := s.orderStatus.OrderStatusesHistory(s.orderStatusRepository)
	orderID := int64(1)
	data := orders.GetFullOrderHistoryParams{
		HTTPRequest: &request,
		OrderID:     orderID,
	}

	s.orderStatusRepository.On("StatusHistory", ctx, int(orderID)).Return(nil, nil)

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusForbidden, responseRecorder.Code)
	s.orderStatusRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_OrderStatusesHistory_EmptyHistory() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.OrderStatusesHistory(s.orderStatusRepository)
	orderID := int64(1)
	data := orders.GetFullOrderHistoryParams{
		HTTPRequest: &request,
		OrderID:     orderID,
	}

	s.orderStatusRepository.On("StatusHistory", ctx, int(orderID)).Return(nil, nil)

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
	s.orderStatusRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_OrderStatusesHistory_MapError() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.OrderStatusesHistory(s.orderStatusRepository)
	orderID := int64(1)
	data := orders.GetFullOrderHistoryParams{
		HTTPRequest: &request,
		OrderID:     orderID,
	}

	count := 1
	history := make([]*ent.OrderStatus, count)
	history[0] = nil
	s.orderStatusRepository.On("StatusHistory", ctx, int(orderID)).Return(history, nil)

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderStatusRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_OrderStatusesHistory_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.OrderStatusesHistory(s.orderStatusRepository)
	orderID := int64(1)
	data := orders.GetFullOrderHistoryParams{
		HTTPRequest: &request,
		OrderID:     orderID,
	}

	count := 1
	history := make([]*ent.OrderStatus, count)
	history[0] = &ent.OrderStatus{
		ID:          1,
		Comment:     "comment",
		CurrentDate: time.Now().UTC(),
		Edges: ent.OrderStatusEdges{
			StatusName: &ent.StatusName{
				ID:     0,
				Status: "test status",
			},
			Users: &ent.User{
				ID:    0,
				Login: "test user",
			},
		},
	}
	s.orderStatusRepository.On("StatusHistory", ctx, int(orderID)).Return(history, nil)

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	response := &models.OrderStatuses{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), response)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, count, len(*response))
	assert.Equal(t, history[0].ID, int(*(*response)[0].ID))
	assert.Equal(t, history[0].Comment, *(*response)[0].Comment)
	assert.Equal(t, strfmt.DateTime(history[0].CurrentDate).String(), (*response)[0].CreatedAt.String())
	assert.Equal(t, history[0].Edges.StatusName.Status, *(*response)[0].Status)
	assert.Equal(t, history[0].Edges.Users.Login, *(*response)[0].ChangedBy.Name)
	assert.Equal(t, history[0].Edges.Users.ID, int(*(*response)[0].ChangedBy.ID))
	s.orderStatusRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_AddNewStatusToOrder_EmptyData() {
	t := s.T()
	request := http.Request{}
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.AddNewStatusToOrder(s.orderStatusRepository)
	data := orders.AddNewOrderStatusParams{
		HTTPRequest: &request,
		Data:        nil,
	}

	resp := handlerFunc(data, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_AddNewStatusToOrder_NoAccess() {
	t := s.T()
	request := http.Request{}
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: "not admin",
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.AddNewStatusToOrder(s.orderStatusRepository)
	data := &models.NewOrderStatus{}
	params := orders.AddNewOrderStatusParams{
		HTTPRequest: &request,
		Data:        data,
	}

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusForbidden, responseRecorder.Code)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_AddNewStatusToOrder_RepoError() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.AddNewStatusToOrder(s.orderStatusRepository)
	data := &models.NewOrderStatus{}
	params := orders.AddNewOrderStatusParams{
		HTTPRequest: &request,
		Data:        data,
	}

	err := errors.New("error")
	s.orderStatusRepository.On("UpdateStatus", ctx, userID, *data).Return(err)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderStatusRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_AddNewStatusToOrder_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.AddNewStatusToOrder(s.orderStatusRepository)
	data := &models.NewOrderStatus{}
	params := orders.AddNewOrderStatusParams{
		HTTPRequest: &request,
		Data:        data,
	}

	s.orderStatusRepository.On("UpdateStatus", ctx, userID, *data).Return(nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.orderStatusRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_GetOrdersByStatus_NoAccess() {
	t := s.T()
	request := http.Request{}
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: "not admin",
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.GetOrdersByStatus(s.orderFilterRepository)
	statusName := "status"
	params := orders.GetOrdersByStatusParams{
		HTTPRequest: &request,
		Status:      statusName,
	}

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusForbidden, responseRecorder.Code)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_GetOrdersByStatus_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.GetOrdersByStatus(s.orderFilterRepository)
	statusName := "status"
	params := orders.GetOrdersByStatusParams{
		HTTPRequest: &request,
		Status:      statusName,
	}

	err := errors.New("error")
	s.orderFilterRepository.On("OrdersByStatus", ctx, statusName).Return(nil, err)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_GetOrdersByStatus_EmptyResult() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.GetOrdersByStatus(s.orderFilterRepository)
	statusName := "status"
	params := orders.GetOrdersByStatusParams{
		HTTPRequest: &request,
		Status:      statusName,
	}

	s.orderFilterRepository.On("OrdersByStatus", ctx, statusName).Return(nil, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_GetOrdersByStatus_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.GetOrdersByStatus(s.orderFilterRepository)
	statusName := "status"
	params := orders.GetOrdersByStatusParams{
		HTTPRequest: &request,
		Status:      statusName,
	}

	count := 1
	ordersToReturn := make([]ent.Order, count)
	ordersToReturn[0] = ent.Order{}
	s.orderFilterRepository.On("OrdersByStatus", ctx, statusName).Return(ordersToReturn, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_GetOrdersByStatus_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.GetOrdersByStatus(s.orderFilterRepository)
	statusName := "status"
	params := orders.GetOrdersByStatusParams{
		HTTPRequest: &request,
		Status:      statusName,
	}

	count := 1
	ordersToReturn := make([]ent.Order, count)
	ordersToReturn[0] = orderWithEdges(t)
	s.orderFilterRepository.On("OrdersByStatus", ctx, statusName).Return(ordersToReturn, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	responseOrders := &[]models.Order{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), responseOrders)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, count, len(*responseOrders))
	assert.Equal(t, ordersToReturn[0].ID, int(*(*responseOrders)[0].ID))
	assert.Equal(t, ordersToReturn[0].Description, *(*responseOrders)[0].Description)
	assert.Equal(t, ordersToReturn[0].Quantity, int(*(*responseOrders)[0].Quantity))
	assert.Equal(t, strfmt.DateTime(ordersToReturn[0].RentStart).String(), (*responseOrders)[0].RentStart.String())
	assert.Equal(t, strfmt.DateTime(ordersToReturn[0].RentEnd).String(), (*responseOrders)[0].RentEnd.String())
	assert.Equal(t, ordersToReturn[0].Edges.Users[0].ID, int(*(*responseOrders)[0].User.ID))
	assert.Equal(t, ordersToReturn[0].Edges.Users[0].Login, *(*responseOrders)[0].User.Name)
	assert.Equal(t, ordersToReturn[0].Edges.Equipments[0].Description, *(*responseOrders)[0].Equipment.Description)
	assert.Equal(t, ordersToReturn[0].Edges.OrderStatus[0].Edges.StatusName.ID, int(*(*responseOrders)[0].LastStatus.ID))
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_NoAccess() {
	t := s.T()
	request := http.Request{}
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: "definitely not admin",
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.GetOrdersByPeriodAndStatus(s.orderFilterRepository)
	params := orders.GetOrdersByDateAndStatusParams{
		HTTPRequest: &request,
	}

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusForbidden, responseRecorder.Code)
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.GetOrdersByPeriodAndStatus(s.orderFilterRepository)
	status := "status"
	fromTime := time.Now().UTC()
	toTime := time.Now().UTC()
	params := orders.GetOrdersByDateAndStatusParams{
		HTTPRequest: &request,
		FromDate:    strfmt.Date(fromTime),
		StatusName:  status,
		ToDate:      strfmt.Date(toTime),
	}

	s.orderFilterRepository.On("OrdersByPeriodAndStatus", ctx, fromTime, toTime, status).Return(nil, errors.New("repo error"))

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.GetOrdersByPeriodAndStatus(s.orderFilterRepository)
	status := "status"
	fromTime := time.Now().UTC()
	toTime := time.Now().UTC()
	params := orders.GetOrdersByDateAndStatusParams{
		HTTPRequest: &request,
		FromDate:    strfmt.Date(fromTime),
		StatusName:  status,
		ToDate:      strfmt.Date(toTime),
	}
	count := 1
	ordersToReturn := make([]ent.Order, count)
	ordersToReturn[0] = ent.Order{}

	s.orderFilterRepository.On("OrdersByPeriodAndStatus", ctx, fromTime, toTime, status).Return(ordersToReturn, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusNameTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()
	userID := 1
	login := "login"
	role := &authentication.Role{
		Id:   userID,
		Slug: authentication.AdminSlug,
	}
	access := authentication.Auth{
		Id:    userID,
		Login: login,
		Role:  role,
	}
	handlerFunc := s.orderStatus.GetOrdersByPeriodAndStatus(s.orderFilterRepository)
	status := "status"
	fromTime := time.Now().UTC()
	toTime := time.Now().UTC()
	params := orders.GetOrdersByDateAndStatusParams{
		HTTPRequest: &request,
		FromDate:    strfmt.Date(fromTime),
		StatusName:  status,
		ToDate:      strfmt.Date(toTime),
	}
	count := 1
	ordersToReturn := make([]ent.Order, count)
	ordersToReturn[0] = orderWithEdges(t)

	s.orderFilterRepository.On("OrdersByPeriodAndStatus", ctx, fromTime, toTime, status).Return(ordersToReturn, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.orderFilterRepository.AssertExpectations(t)
}
