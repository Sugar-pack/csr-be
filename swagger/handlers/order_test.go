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

func orderWithNoEdges() *ent.Order {
	return &ent.Order{
		ID: 1,
	}
}

func orderWithAllEdges() *ent.Order {
	return &ent.Order{
		ID: 100,
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

type orderTestSuite struct {
	suite.Suite
	logger          *zap.Logger
	orderRepository *repomock.OrderRepository
	orderHandler    *Order
}

func TestOrderSuite(t *testing.T) {
	suite.Run(t, new(orderTestSuite))
}

func (s *orderTestSuite) SetupTest() {
	s.logger = zap.NewExample()
	s.orderRepository = &repomock.OrderRepository{}
	s.orderHandler = NewOrder(s.logger)
}

func (s *orderTestSuite) TestOrder_ListOrder_AccessErr() {
	t := s.T()
	request := http.Request{}

	handlerFunc := s.orderHandler.ListOrderFunc(s.orderRepository)
	data := orders.GetAllOrdersParams{HTTPRequest: &request}
	access := "definitely not an access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_ListOrder_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	err := errors.New("error")
	s.orderRepository.On("List", ctx, userID).Return(nil, 0, err)

	handlerFunc := s.orderHandler.ListOrderFunc(s.orderRepository)
	data := orders.GetAllOrdersParams{HTTPRequest: &request}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_ListOrder_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	var orderList []*ent.Order
	orderList = append(orderList, orderWithNoEdges())
	s.orderRepository.On("List", ctx, userID).Return(orderList, 1, nil)

	handlerFunc := s.orderHandler.ListOrderFunc(s.orderRepository)
	data := orders.GetAllOrdersParams{HTTPRequest: &request}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_ListOrder_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	var orderList []*ent.Order
	orderList = append(orderList, orderWithAllEdges())
	s.orderRepository.On("List", ctx, userID).Return(orderList, 1, nil)

	handlerFunc := s.orderHandler.ListOrderFunc(s.orderRepository)
	data := orders.GetAllOrdersParams{HTTPRequest: &request}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var response orders.GetAllOrdersOKBody
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(orderList), int(response.Data.Total))
	assert.Equal(t, orderList[0].ID, int(*response.Data.Items[0].ID))

	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_CreateOrder_AccessErr() {
	t := s.T()
	request := http.Request{}

	handlerFunc := s.orderHandler.CreateOrderFunc(s.orderRepository)
	data := orders.CreateOrderParams{
		HTTPRequest: &request,
	}
	access := "definitely not an access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_CreateOrder_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	description := "description"
	equipment := int64(1)
	quantity := int64(10)
	rentStart := strfmt.DateTime(time.Now())
	rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	createOrder := &models.OrderCreateRequest{
		Description: &description,
		Equipment:   &equipment,
		Quantity:    &quantity,
		RentStart:   &rentStart,
		RentEnd:     &rentEnd,
	}
	userID := 1
	err := errors.New("error")
	s.orderRepository.On("Create", ctx, createOrder, userID).Return(nil, err)

	handlerFunc := s.orderHandler.CreateOrderFunc(s.orderRepository)
	data := orders.CreateOrderParams{
		HTTPRequest: &request,
		Data:        createOrder,
	}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_CreateOrder_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	description := "description"
	equipment := int64(1)
	quantity := int64(10)
	rentStart := strfmt.DateTime(time.Now())
	rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	createOrder := &models.OrderCreateRequest{
		Description: &description,
		Equipment:   &equipment,
		Quantity:    &quantity,
		RentStart:   &rentStart,
		RentEnd:     &rentEnd,
	}
	userID := 1

	orderToReturn := orderWithNoEdges()
	s.orderRepository.On("Create", ctx, createOrder, userID).Return(orderToReturn, nil)

	handlerFunc := s.orderHandler.CreateOrderFunc(s.orderRepository)
	data := orders.CreateOrderParams{
		HTTPRequest: &request,
		Data:        createOrder,
	}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_CreateOrder_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	description := "description"
	equipment := int64(1)
	quantity := int64(10)
	rentStart := strfmt.DateTime(time.Now())
	rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	createOrder := &models.OrderCreateRequest{
		Description: &description,
		Equipment:   &equipment,
		Quantity:    &quantity,
		RentStart:   &rentStart,
		RentEnd:     &rentEnd,
	}
	userID := 1

	orderToReturn := orderWithAllEdges()
	s.orderRepository.On("Create", ctx, createOrder, userID).Return(orderToReturn, nil)

	handlerFunc := s.orderHandler.CreateOrderFunc(s.orderRepository)
	data := orders.CreateOrderParams{
		HTTPRequest: &request,
		Data:        createOrder,
	}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)
	responseOrder := models.Order{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseOrder)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, orderToReturn.ID, int(*responseOrder.ID))

	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_UpdateOrder_AccessErr() {
	t := s.T()
	request := http.Request{}

	handlerFunc := s.orderHandler.UpdateOrderFunc(s.orderRepository)
	data := orders.UpdateOrderParams{
		HTTPRequest: &request,
	}
	access := "definitely not an access"
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_UpdateOrder_RepoErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	description := "description"
	quantity := int64(10)
	rentStart := strfmt.DateTime(time.Now())
	rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	createOrder := &models.OrderUpdateRequest{
		Description: &description,
		Quantity:    &quantity,
		RentStart:   &rentStart,
		RentEnd:     &rentEnd,
	}
	userID := 1
	orderID := 2
	err := errors.New("error")
	s.orderRepository.On("Update", ctx, orderID, createOrder, userID).Return(nil, err)

	handlerFunc := s.orderHandler.UpdateOrderFunc(s.orderRepository)
	data := orders.UpdateOrderParams{
		HTTPRequest: &request,
		Data:        createOrder,
		OrderID:     int64(orderID),
	}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_UpdateOrder_MapErr() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	description := "description"
	quantity := int64(10)
	rentStart := strfmt.DateTime(time.Now())
	rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	createOrder := &models.OrderUpdateRequest{
		Description: &description,
		Quantity:    &quantity,
		RentStart:   &rentStart,
		RentEnd:     &rentEnd,
	}
	userID := 1
	orderID := 2
	orderToReturn := orderWithNoEdges()
	s.orderRepository.On("Update", ctx, orderID, createOrder, userID).Return(orderToReturn, nil)

	handlerFunc := s.orderHandler.UpdateOrderFunc(s.orderRepository)
	data := orders.UpdateOrderParams{
		HTTPRequest: &request,
		Data:        createOrder,
		OrderID:     int64(orderID),
	}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_UpdateOrder_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	description := "description"
	quantity := int64(10)
	rentStart := strfmt.DateTime(time.Now())
	rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	createOrder := &models.OrderUpdateRequest{
		Description: &description,
		Quantity:    &quantity,
		RentStart:   &rentStart,
		RentEnd:     &rentEnd,
	}
	userID := 1
	orderID := 2
	orderToReturn := orderWithAllEdges()
	s.orderRepository.On("Update", ctx, orderID, createOrder, userID).Return(orderToReturn, nil)

	handlerFunc := s.orderHandler.UpdateOrderFunc(s.orderRepository)
	data := orders.UpdateOrderParams{
		HTTPRequest: &request,
		Data:        createOrder,
		OrderID:     int64(orderID),
	}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseOrder := models.Order{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseOrder)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, orderToReturn.ID, int(*responseOrder.ID))

	s.orderRepository.AssertExpectations(t)
}
