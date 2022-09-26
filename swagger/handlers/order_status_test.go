package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/orders"
)

func TestSetOrderStatusHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:orderstatushandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	SetOrderStatusHandler(logger, api)
	assert.NotEmpty(t, api.OrdersGetOrdersByStatusHandler)
	assert.NotEmpty(t, api.OrdersGetOrdersByDateAndStatusHandler)
	assert.NotEmpty(t, api.OrdersAddNewOrderStatusHandler)
	assert.NotEmpty(t, api.OrdersGetFullOrderHistoryHandler)
	assert.NotEmpty(t, api.OrdersGetAllStatusNamesHandler)
}

type OrderStatusTestSuite struct {
	suite.Suite
	logger                *zap.Logger
	statusNameRepository  *repomock.StatusNameRepository
	orderStatusRepository *repomock.OrderStatusRepository
	orderFilterRepository *repomock.OrderRepositoryWithFilter
	orderStatus           *OrderStatus
}

func orderWithEdges(t *testing.T, id int) *ent.Order {
	t.Helper()
	return &ent.Order{
		ID:          id,
		Description: fmt.Sprintf("test description %d", id),
		Quantity:    id%2 + 1,
		RentStart:   time.Now().Add(time.Duration(-id*24) * time.Hour),
		RentEnd:     time.Now().Add(time.Duration(id*24) * time.Hour),
		CreatedAt:   time.Now().Add(time.Duration(-id) * time.Hour),
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
					ID: id,
					Edges: ent.OrderStatusEdges{
						StatusName: &ent.StatusName{
							ID: id%2 + 1,
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
	s := new(OrderStatusTestSuite)
	suite.Run(t, s)
}

func (s *OrderStatusTestSuite) SetupTest() {
	s.logger = zap.NewExample()
	s.statusNameRepository = &repomock.StatusNameRepository{}
	s.orderStatusRepository = &repomock.OrderStatusRepository{}
	s.orderFilterRepository = &repomock.OrderRepositoryWithFilter{}
	s.orderStatus = NewOrderStatus(s.logger)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetAllStatusNames_OK() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_GetAllStatusNames_RepoErr() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_GetAllStatusNames_MapError() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_OrderStatusesHistory_RepoErr() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_OrderStatusesHistory_CantAccess() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_OrderStatusesHistory_EmptyHistory() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_OrderStatusesHistory_MapError() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_OrderStatusesHistory_OK() {
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
		t.Fatal(err)
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

func (s *OrderStatusTestSuite) TestOrderStatus_AddNewStatusToOrder_EmptyData() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_AddNewStatusToOrder_NoAccess() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_AddNewStatusToOrder_RepoError() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_AddNewStatusToOrder_OK() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByStatus_NoAccess() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByStatus_RepoErr() {
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
	s.orderFilterRepository.On("OrdersByStatusTotal", ctx, statusName).Return(0, err)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByStatus_EmptyResult() {
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

	s.orderFilterRepository.On("OrdersByStatusTotal", ctx, statusName).
		Return(0, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseOrders := models.OrderList{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseOrders)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, int(*responseOrders.Total))
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByStatus_MapErr() {
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
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	params := orders.GetOrdersByStatusParams{
		HTTPRequest: &request,
		Status:      statusName,
	}

	ordersToReturn := []*ent.Order{{}}

	s.orderFilterRepository.On("OrdersByStatusTotal", ctx, statusName).
		Return(1, nil)
	s.orderFilterRepository.On("OrdersByStatus", ctx, statusName, limit, offset, orderBy, orderColumn).
		Return(ordersToReturn, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByStatus_EmptyPaginationParams() {
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
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	params := orders.GetOrdersByStatusParams{
		HTTPRequest: &request,
		Status:      statusName,
	}

	ordersToReturn := []*ent.Order{
		orderWithEdges(t, 1),
	}
	s.orderFilterRepository.On("OrdersByStatusTotal", ctx, statusName).
		Return(1, nil)
	s.orderFilterRepository.On("OrdersByStatus",
		ctx, statusName, limit, offset, orderBy, orderColumn).
		Return(ordersToReturn, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	responseOrders := models.OrderList{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseOrders)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(ordersToReturn), int(*responseOrders.Total))
	assert.Equal(t, len(ordersToReturn), len(responseOrders.Items))
	for _, o := range responseOrders.Items {
		assert.True(t, containsOrder(t, ordersToReturn, o))
	}
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByStatus_LimitGreaterThanTotal() {
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
	limit := int64(5)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	params := orders.GetOrdersByStatusParams{
		HTTPRequest: &request,
		Status:      statusName,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	allOrders := []*ent.Order{
		orderWithEdges(t, 1),
		orderWithEdges(t, 2),
		orderWithEdges(t, 3),
		orderWithEdges(t, 4),
		orderWithEdges(t, 5),
	}
	ordersByStatus := []*ent.Order{
		allOrders[0],
		allOrders[2],
		allOrders[4],
	}
	s.orderFilterRepository.On("OrdersByStatusTotal", ctx, statusName).
		Return(len(ordersByStatus), nil)
	s.orderFilterRepository.On("OrdersByStatus",
		ctx, statusName, int(limit), int(offset), orderBy, orderColumn).
		Return(ordersByStatus, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	responseOrders := models.OrderList{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseOrders)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(ordersByStatus), int(*responseOrders.Total))
	assert.GreaterOrEqual(t, int(limit), len(responseOrders.Items))
	for _, o := range responseOrders.Items {
		assert.True(t, containsOrder(t, ordersByStatus, o))
	}
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByStatus_LimitLessThanTotal() {
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
	limit := int64(2)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	params := orders.GetOrdersByStatusParams{
		HTTPRequest: &request,
		Status:      statusName,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	allOrders := []*ent.Order{
		orderWithEdges(t, 1),
		orderWithEdges(t, 2),
		orderWithEdges(t, 3),
		orderWithEdges(t, 4),
		orderWithEdges(t, 5),
	}
	ordersByStatus := []*ent.Order{
		allOrders[0],
		allOrders[2],
		allOrders[4],
	}
	s.orderFilterRepository.On("OrdersByStatusTotal", ctx, statusName).
		Return(len(ordersByStatus), nil)
	s.orderFilterRepository.On("OrdersByStatus",
		ctx, statusName, int(limit), int(offset), orderBy, orderColumn).
		Return(ordersByStatus[:limit], nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	responseOrders := models.OrderList{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseOrders)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(ordersByStatus), int(*responseOrders.Total))
	assert.GreaterOrEqual(t, int(limit), len(responseOrders.Items))
	assert.Greater(t, len(ordersByStatus), len(responseOrders.Items))
	for _, o := range responseOrders.Items {
		assert.True(t, containsOrder(t, ordersByStatus, o))
	}
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByStatus_SecondPage() {
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
	limit := int64(2)
	offset := int64(2)
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	params := orders.GetOrdersByStatusParams{
		HTTPRequest: &request,
		Status:      statusName,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	allOrders := []*ent.Order{
		orderWithEdges(t, 1),
		orderWithEdges(t, 2),
		orderWithEdges(t, 3),
		orderWithEdges(t, 4),
		orderWithEdges(t, 5),
	}
	ordersByStatus := []*ent.Order{
		allOrders[0],
		allOrders[2],
		allOrders[4],
	}
	s.orderFilterRepository.On("OrdersByStatusTotal", ctx, statusName).
		Return(len(ordersByStatus), nil)
	s.orderFilterRepository.On("OrdersByStatus",
		ctx, statusName, int(limit), int(offset), orderBy, orderColumn).
		Return(ordersByStatus[offset:], nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	responseOrders := models.OrderList{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseOrders)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(ordersByStatus), int(*responseOrders.Total))
	assert.GreaterOrEqual(t, int(limit), len(responseOrders.Items))
	assert.Equal(t, len(ordersByStatus)-int(offset), len(responseOrders.Items))
	assert.Greater(t, len(ordersByStatus), len(responseOrders.Items))
	for _, o := range responseOrders.Items {
		assert.True(t, containsOrder(t, ordersByStatus, o))
	}
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByStatus_SeveralPages() {
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
	limit := int64(2)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	params := orders.GetOrdersByStatusParams{
		HTTPRequest: &request,
		Status:      statusName,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}

	allOrders := []*ent.Order{
		orderWithEdges(t, 1),
		orderWithEdges(t, 2),
		orderWithEdges(t, 3),
		orderWithEdges(t, 4),
		orderWithEdges(t, 5),
	}
	ordersByStatus := []*ent.Order{
		allOrders[0],
		allOrders[2],
		allOrders[4],
	}
	s.orderFilterRepository.On("OrdersByStatusTotal", ctx, statusName).
		Return(len(ordersByStatus), nil)
	s.orderFilterRepository.On("OrdersByStatus",
		ctx, statusName, int(limit), int(offset), orderBy, orderColumn).
		Return(ordersByStatus[:limit], nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	firstPage := models.OrderList{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &firstPage)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(ordersByStatus), int(*firstPage.Total))
	assert.Equal(t, int(limit), len(firstPage.Items))
	for _, o := range firstPage.Items {
		assert.True(t, containsOrder(t, ordersByStatus, o))
	}

	offset = limit
	params = orders.GetOrdersByStatusParams{
		HTTPRequest: &request,
		Status:      statusName,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}
	s.orderFilterRepository.On("OrdersByStatusTotal", ctx, statusName).
		Return(len(ordersByStatus), nil)
	s.orderFilterRepository.On("OrdersByStatus",
		ctx, statusName, int(limit), int(offset), orderBy, orderColumn).
		Return(ordersByStatus[offset:], nil)

	resp = handlerFunc(params, access)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	secondPage := models.OrderList{}
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &secondPage)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(ordersByStatus), int(*secondPage.Total))
	assert.GreaterOrEqual(t, int(limit), len(secondPage.Items))
	assert.Equal(t, len(ordersByStatus)-int(offset), len(secondPage.Items))
	assert.Greater(t, len(ordersByStatus), len(secondPage.Items))
	for _, o := range secondPage.Items {
		assert.True(t, containsOrder(t, ordersByStatus, o))
	}

	assert.False(t, ordersDuplicated(t, firstPage.Items, secondPage.Items))
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_NoAccess() {
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

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_RepoErr() {
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
	s.orderFilterRepository.On("OrdersByPeriodAndStatusTotal", ctx, fromTime, toTime, status).
		Return(0, errors.New("repo error"))

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_EmptyResult() {
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
	s.orderFilterRepository.On("OrdersByPeriodAndStatusTotal", ctx, fromTime, toTime, status).
		Return(0, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	responseOrders := models.OrderList{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseOrders)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, int(*responseOrders.Total))
	assert.Equal(t, 0, len(responseOrders.Items))
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_MapErr() {
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
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	params := orders.GetOrdersByDateAndStatusParams{
		HTTPRequest: &request,
		FromDate:    strfmt.Date(fromTime),
		StatusName:  status,
		ToDate:      strfmt.Date(toTime),
	}
	ordersToReturn := []*ent.Order{{}}

	s.orderFilterRepository.On("OrdersByPeriodAndStatusTotal", ctx, fromTime, toTime, status).
		Return(1, nil)
	s.orderFilterRepository.On("OrdersByPeriodAndStatus",
		ctx, fromTime, toTime, status, limit, offset, orderBy, orderColumn).
		Return(ordersToReturn, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_EmptyPaginationParams() {
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
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	params := orders.GetOrdersByDateAndStatusParams{
		HTTPRequest: &request,
		FromDate:    strfmt.Date(fromTime),
		StatusName:  status,
		ToDate:      strfmt.Date(toTime),
	}
	ordersToReturn := []*ent.Order{
		orderWithEdges(t, 1),
	}

	s.orderFilterRepository.On("OrdersByPeriodAndStatusTotal", ctx, fromTime, toTime, status).
		Return(1, nil)
	s.orderFilterRepository.On("OrdersByPeriodAndStatus",
		ctx, fromTime, toTime, status, limit, offset, orderBy, orderColumn).
		Return(ordersToReturn, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_LimitGreaterThanTotal() {
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
	statusName := "status"
	limit := int64(5)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	from := time.Now().Add(-6 * time.Hour)
	to := time.Now().Add(6 * time.Hour)
	params := orders.GetOrdersByDateAndStatusParams{
		HTTPRequest: &request,
		StatusName:  statusName,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
		FromDate:    strfmt.Date(from),
		ToDate:      strfmt.Date(to),
	}

	allOrders := []*ent.Order{
		orderWithEdges(t, 1),
		orderWithEdges(t, 3),
		orderWithEdges(t, 5),
		orderWithEdges(t, 7),
		orderWithEdges(t, 9),
	}
	ordersByStatus := []*ent.Order{
		allOrders[0],
		allOrders[1],
		allOrders[2],
	}
	s.orderFilterRepository.On("OrdersByPeriodAndStatusTotal", ctx, from, to, statusName).
		Return(len(ordersByStatus), nil)
	s.orderFilterRepository.On("OrdersByPeriodAndStatus",
		ctx, from, to, statusName, int(limit), int(offset), orderBy, orderColumn).
		Return(ordersByStatus, nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	responseOrders := models.OrderList{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseOrders)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(ordersByStatus), int(*responseOrders.Total))
	assert.GreaterOrEqual(t, int(limit), len(responseOrders.Items))
	for _, o := range responseOrders.Items {
		assert.True(t, containsOrder(t, ordersByStatus, o))
	}
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_LimitLessThanTotal() {
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
	statusName := "status"
	limit := int64(2)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	from := time.Now().Add(-6 * time.Hour)
	to := time.Now().Add(6 * time.Hour)
	params := orders.GetOrdersByDateAndStatusParams{
		HTTPRequest: &request,
		StatusName:  statusName,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
		FromDate:    strfmt.Date(from),
		ToDate:      strfmt.Date(to),
	}

	allOrders := []*ent.Order{
		orderWithEdges(t, 1),
		orderWithEdges(t, 3),
		orderWithEdges(t, 5),
		orderWithEdges(t, 7),
		orderWithEdges(t, 9),
	}
	ordersByStatus := []*ent.Order{
		allOrders[0],
		allOrders[1],
		allOrders[2],
	}
	s.orderFilterRepository.On("OrdersByPeriodAndStatusTotal", ctx, from, to, statusName).
		Return(len(ordersByStatus), nil)
	s.orderFilterRepository.On("OrdersByPeriodAndStatus",
		ctx, from, to, statusName, int(limit), int(offset), orderBy, orderColumn).
		Return(ordersByStatus[:limit], nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	responseOrders := models.OrderList{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseOrders)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(ordersByStatus), int(*responseOrders.Total))
	assert.GreaterOrEqual(t, int(limit), len(responseOrders.Items))
	assert.Greater(t, len(ordersByStatus), len(responseOrders.Items))
	for _, o := range responseOrders.Items {
		assert.True(t, containsOrder(t, ordersByStatus, o))
	}
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_SecondPage() {
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
	statusName := "status"
	limit := int64(2)
	offset := int64(2)
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	from := time.Now().Add(-6 * time.Hour)
	to := time.Now().Add(6 * time.Hour)
	params := orders.GetOrdersByDateAndStatusParams{
		HTTPRequest: &request,
		StatusName:  statusName,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
		FromDate:    strfmt.Date(from),
		ToDate:      strfmt.Date(to),
	}

	allOrders := []*ent.Order{
		orderWithEdges(t, 1),
		orderWithEdges(t, 3),
		orderWithEdges(t, 5),
		orderWithEdges(t, 7),
		orderWithEdges(t, 9),
	}
	ordersByStatus := []*ent.Order{
		allOrders[0],
		allOrders[1],
		allOrders[2],
	}
	s.orderFilterRepository.On("OrdersByPeriodAndStatusTotal", ctx, from, to, statusName).
		Return(len(ordersByStatus), nil)
	s.orderFilterRepository.On("OrdersByPeriodAndStatus",
		ctx, from, to, statusName, int(limit), int(offset), orderBy, orderColumn).
		Return(ordersByStatus[offset:], nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	responseOrders := models.OrderList{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseOrders)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(ordersByStatus), int(*responseOrders.Total))
	assert.GreaterOrEqual(t, int(limit), len(responseOrders.Items))
	assert.Equal(t, len(ordersByStatus)-int(offset), len(responseOrders.Items))
	assert.Greater(t, len(ordersByStatus), len(responseOrders.Items))
	for _, o := range responseOrders.Items {
		assert.True(t, containsOrder(t, ordersByStatus, o))
	}
	s.orderFilterRepository.AssertExpectations(t)
}

func (s *OrderStatusTestSuite) TestOrderStatus_GetOrdersByPeriodAndStatus_SeveralPages() {
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
	statusName := "status"
	limit := int64(2)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	from := time.Now().Add(-6 * time.Hour)
	to := time.Now().Add(6 * time.Hour)
	params := orders.GetOrdersByDateAndStatusParams{
		HTTPRequest: &request,
		StatusName:  statusName,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
		FromDate:    strfmt.Date(from),
		ToDate:      strfmt.Date(to),
	}

	allOrders := []*ent.Order{
		orderWithEdges(t, 1),
		orderWithEdges(t, 2),
		orderWithEdges(t, 3),
		orderWithEdges(t, 4),
		orderWithEdges(t, 5),
	}
	ordersByStatus := []*ent.Order{
		allOrders[0],
		allOrders[2],
		allOrders[4],
	}
	s.orderFilterRepository.On("OrdersByPeriodAndStatusTotal", ctx, from, to, statusName).
		Return(len(ordersByStatus), nil)
	s.orderFilterRepository.On("OrdersByPeriodAndStatus",
		ctx, from, to, statusName, int(limit), int(offset), orderBy, orderColumn).
		Return(ordersByStatus[:limit], nil)

	resp := handlerFunc(params, access)
	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	firstPage := models.OrderList{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &firstPage)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(ordersByStatus), int(*firstPage.Total))
	assert.Equal(t, int(limit), len(firstPage.Items))
	for _, o := range firstPage.Items {
		assert.True(t, containsOrder(t, ordersByStatus, o))
	}

	offset = limit
	params = orders.GetOrdersByDateAndStatusParams{
		HTTPRequest: &request,
		StatusName:  statusName,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
		FromDate:    strfmt.Date(from),
		ToDate:      strfmt.Date(to),
	}
	s.orderFilterRepository.On("OrdersByPeriodAndStatusTotal", ctx, from, to, statusName).
		Return(len(ordersByStatus), nil)
	s.orderFilterRepository.On("OrdersByPeriodAndStatus",
		ctx, from, to, statusName, int(limit), int(offset), orderBy, orderColumn).
		Return(ordersByStatus[offset:], nil)

	resp = handlerFunc(params, access)
	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	secondPage := models.OrderList{}
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &secondPage)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(ordersByStatus), int(*secondPage.Total))
	assert.GreaterOrEqual(t, int(limit), len(secondPage.Items))
	assert.Equal(t, len(ordersByStatus)-int(offset), len(secondPage.Items))
	assert.Greater(t, len(ordersByStatus), len(secondPage.Items))
	for _, o := range secondPage.Items {
		assert.True(t, containsOrder(t, ordersByStatus, o))
	}

	assert.False(t, ordersDuplicated(t, firstPage.Items, secondPage.Items))
	s.orderFilterRepository.AssertExpectations(t)
}
