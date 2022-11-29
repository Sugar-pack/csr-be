package handlers

import (
	"encoding/json"
	"errors"
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

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	equipmentEnt "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func TestSetOrderHandler(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:orderhandler?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	logger := zap.NewNop()

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		t.Fatal(err)
	}
	api := operations.NewBeAPI(swaggerSpec)
	SetOrderHandler(logger, api)
	assert.NotEmpty(t, api.OrdersGetAllOrdersHandler)
	assert.NotEmpty(t, api.OrdersCreateOrderHandler)
	assert.NotEmpty(t, api.OrdersUpdateOrderHandler)
}

func orderWithNoEdges() *ent.Order {
	return &ent.Order{
		ID: 1,
	}
}

func orderWithAllEdges(t *testing.T, orderID int) *ent.Order {
	t.Helper()
	return &ent.Order{
		ID: orderID,
		Edges: ent.OrderEdges{
			Users: &ent.User{
				ID:    1,
				Login: "login",
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
						OrderStatusName: &ent.OrderStatusName{
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
	logger              *zap.Logger
	orderRepository     *mocks.OrderRepository
	eqStatusRepository  *mocks.EquipmentStatusRepository
	equipmentRepository *mocks.EquipmentRepository
	orderHandler        *Order
}

func TestOrderSuite(t *testing.T) {
	suite.Run(t, new(orderTestSuite))
}

func (s *orderTestSuite) SetupTest() {
	s.logger = zap.NewExample()
	s.orderRepository = &mocks.OrderRepository{}
	s.eqStatusRepository = &mocks.EquipmentStatusRepository{}
	s.equipmentRepository = &mocks.EquipmentRepository{}
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
	s.orderRepository.On("OrdersTotal", ctx, userID).Return(0, err)

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
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	var orderList []*ent.Order
	orderList = append(orderList, orderWithNoEdges())
	s.orderRepository.On("OrdersTotal", ctx, userID).Return(1, nil)
	s.orderRepository.On("List", ctx, userID, limit, offset, orderBy, orderColumn).
		Return(orderList, nil)

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

func (s *orderTestSuite) TestOrder_ListOrder_NotFound() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	s.orderRepository.On("OrdersTotal", ctx, userID).Return(0, nil)

	handlerFunc := s.orderHandler.ListOrderFunc(s.orderRepository)
	data := orders.GetAllOrdersParams{HTTPRequest: &request}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var response models.OrderList
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, int(*response.Total))
	assert.Equal(t, 0, len(response.Items))
	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_ListOrder_EmptyParams() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	orderList := []*ent.Order{
		orderWithAllEdges(t, 1),
	}
	s.orderRepository.On("OrdersTotal", ctx, userID).Return(1, nil)
	s.orderRepository.On("List", ctx, userID, limit, offset, orderBy, orderColumn).
		Return(orderList, nil)

	handlerFunc := s.orderHandler.ListOrderFunc(s.orderRepository)
	data := orders.GetAllOrdersParams{HTTPRequest: &request}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var response models.OrderList
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(orderList), int(*response.Total))
	assert.GreaterOrEqual(t, limit, len(response.Items))
	for _, item := range response.Items {
		assert.True(t, containsOrder(t, orderList, item))
	}
	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_ListOrder_LimitGreaterThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	limit := int64(10)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	orderList := []*ent.Order{
		orderWithAllEdges(t, 1),
		orderWithAllEdges(t, 2),
	}
	s.orderRepository.On("OrdersTotal", ctx, userID).Return(2, nil)
	s.orderRepository.On("List", ctx, userID, int(limit), int(offset), orderBy, orderColumn).
		Return(orderList, nil)

	handlerFunc := s.orderHandler.ListOrderFunc(s.orderRepository)
	data := orders.GetAllOrdersParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var response models.OrderList
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(orderList), int(*response.Total))
	assert.GreaterOrEqual(t, int(limit), len(response.Items))
	for _, item := range response.Items {
		assert.True(t, containsOrder(t, orderList, item))
	}

	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_ListOrder_LimitLessThanTotal() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	limit := int64(3)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	orderList := []*ent.Order{
		orderWithAllEdges(t, 1),
		orderWithAllEdges(t, 2),
		orderWithAllEdges(t, 3),
		orderWithAllEdges(t, 4),
	}
	s.orderRepository.On("OrdersTotal", ctx, userID).Return(4, nil)
	s.orderRepository.On("List", ctx, userID, int(limit), int(offset), orderBy, orderColumn).
		Return(orderList[:limit], nil)

	handlerFunc := s.orderHandler.ListOrderFunc(s.orderRepository)
	data := orders.GetAllOrdersParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var response models.OrderList
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(orderList), int(*response.Total))
	assert.GreaterOrEqual(t, int(limit), len(response.Items))
	for _, item := range response.Items {
		assert.True(t, containsOrder(t, orderList, item))
	}

	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_ListOrder_SecondPage() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	limit := int64(10)
	offset := int64(2)
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	orderList := []*ent.Order{
		orderWithAllEdges(t, 1),
		orderWithAllEdges(t, 2),
		orderWithAllEdges(t, 3),
		orderWithAllEdges(t, 4),
	}
	s.orderRepository.On("OrdersTotal", ctx, userID).Return(4, nil)
	s.orderRepository.On("List", ctx, userID, int(limit), int(offset), orderBy, orderColumn).
		Return(orderList[offset:], nil)

	handlerFunc := s.orderHandler.ListOrderFunc(s.orderRepository)
	data := orders.GetAllOrdersParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var response models.OrderList
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(orderList), int(*response.Total))
	assert.Equal(t, len(orderList)-int(offset), len(response.Items))
	for _, item := range response.Items {
		assert.True(t, containsOrder(t, orderList, item))
	}

	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_ListOrder_SeveralPages() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	userID := 1
	limit := int64(3)
	offset := int64(0)
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	orderList := []*ent.Order{
		orderWithAllEdges(t, 1),
		orderWithAllEdges(t, 2),
		orderWithAllEdges(t, 3),
		orderWithAllEdges(t, 4),
	}
	s.orderRepository.On("OrdersTotal", ctx, userID).Return(4, nil)
	s.orderRepository.On("List", ctx, userID, int(limit), int(offset), orderBy, orderColumn).
		Return(orderList[:limit], nil)

	handlerFunc := s.orderHandler.ListOrderFunc(s.orderRepository)
	data := orders.GetAllOrdersParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}
	access := authentication.Auth{Id: userID}
	resp := handlerFunc.Handle(data, access)

	responseRecorder := httptest.NewRecorder()
	producer := runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var firstPage models.OrderList
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &firstPage)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(orderList), int(*firstPage.Total))
	assert.Equal(t, int(limit), len(firstPage.Items))
	for _, item := range firstPage.Items {
		assert.True(t, containsOrder(t, orderList, item))
	}

	offset = limit
	s.orderRepository.On("OrdersTotal", ctx, userID).Return(4, nil)
	s.orderRepository.On("List", ctx, userID, int(limit), int(offset), orderBy, orderColumn).
		Return(orderList[offset:], nil)

	data = orders.GetAllOrdersParams{
		HTTPRequest: &request,
		Limit:       &limit,
		Offset:      &offset,
		OrderBy:     &orderBy,
		OrderColumn: &orderColumn,
	}
	resp = handlerFunc.Handle(data, access)

	responseRecorder = httptest.NewRecorder()
	producer = runtime.JSONProducer()
	resp.WriteResponse(responseRecorder, producer)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var secondPage models.OrderList
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &secondPage)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(orderList), int(*secondPage.Total))
	assert.Equal(t, len(orderList)-int(offset), len(secondPage.Items))
	for _, item := range secondPage.Items {
		assert.True(t, containsOrder(t, orderList, item))
	}

	assert.False(t, ordersDuplicated(t, firstPage.Items, secondPage.Items))
	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_CreateOrder_AccessErr() {
	t := s.T()
	request := http.Request{}

	handlerFunc := s.orderHandler.CreateOrderFunc(s.orderRepository, s.eqStatusRepository, s.equipmentRepository)
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
	categoryID := int64(3)
	quantity := int64(10)
	rentStart := strfmt.DateTime(time.Now())
	rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	createOrder := &models.OrderCreateRequest{
		Description: &description,
		Category:    &categoryID,
		Quantity:    &quantity,
		RentStart:   &rentStart,
		RentEnd:     &rentEnd,
	}
	userID := 1
	err := errors.New("error")
	s.equipmentRepository.On("EquipmentsByFilter", ctx, models.EquipmentFilter{
		Category: categoryID,
	}, math.MaxInt, 0, utils.DescOrder, equipmentEnt.FieldID).Return(nil, err)

	handlerFunc := s.orderHandler.CreateOrderFunc(s.orderRepository, s.eqStatusRepository, s.equipmentRepository)
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
	categoryID := int64(1)
	quantity := int64(1)
	rentStart := strfmt.DateTime(time.Now())
	rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	createOrder := &models.OrderCreateRequest{
		Description: &description,
		Category:    &categoryID,
		Quantity:    &quantity,
		RentStart:   &rentStart,
		RentEnd:     &rentEnd,
	}
	userID := 1

	orderToReturn := orderWithNoEdges()
	equipment := orderWithEdges(t, 1).Edges.Equipments[0]
	equipmentID := int64(equipment.ID)
	endDate := time.Time(rentEnd).AddDate(0, 0, 1)
	equipmentBookedEndDate := strfmt.DateTime(endDate)
	s.equipmentRepository.On("EquipmentsByFilter", ctx, models.EquipmentFilter{
		Category: categoryID,
	}, math.MaxInt, 0, utils.DescOrder, equipmentEnt.FieldID).Return([]*ent.Equipment{equipment}, nil)
	s.eqStatusRepository.On("HasStatusByPeriod", ctx, domain.EquipmentStatusAvailable, equipment.ID,
		time.Time(rentStart), time.Time(rentEnd)).Return(true, nil)
	s.orderRepository.On("Create", ctx, createOrder, userID, []int{equipment.ID}).Return(orderToReturn, nil)
	s.eqStatusRepository.On("Create", ctx, &models.NewEquipmentStatus{
		EndDate:     &equipmentBookedEndDate,
		EquipmentID: &equipmentID,
		OrderID:     int64(orderToReturn.ID),
		StartDate:   createOrder.RentStart,
		StatusName:  &domain.EquipmentStatusBooked,
	}).Return(nil, nil)

	handlerFunc := s.orderHandler.CreateOrderFunc(s.orderRepository, s.eqStatusRepository, s.equipmentRepository)
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

func (s *orderTestSuite) TestOrder_CreateOrder_NoAvailableEquipments() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	description := "description"
	categoryID := int64(3)
	quantity := int64(1)
	rentStart := strfmt.DateTime(time.Now())
	rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	createOrder := &models.OrderCreateRequest{
		Description: &description,
		Category:    &categoryID,
		Quantity:    &quantity,
		RentStart:   &rentStart,
		RentEnd:     &rentEnd,
	}
	userID := 1

	equipment := orderWithEdges(t, 1).Edges.Equipments[0]
	s.equipmentRepository.On("EquipmentsByFilter", ctx, models.EquipmentFilter{
		Category: categoryID,
	}, math.MaxInt, 0, utils.DescOrder, equipmentEnt.FieldID).Return([]*ent.Equipment{equipment}, nil)
	s.eqStatusRepository.On("HasStatusByPeriod", ctx, domain.EquipmentStatusAvailable, equipment.ID,
		time.Time(rentStart), time.Time(rentEnd)).Return(false, nil)

	handlerFunc := s.orderHandler.CreateOrderFunc(s.orderRepository, s.eqStatusRepository, s.equipmentRepository)
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
	responseOrder := models.Order{}
	err := json.Unmarshal(responseRecorder.Body.Bytes(), &responseOrder)
	if err != nil {
		t.Fatal(err)
	}
	assert.Empty(t, responseOrder)

	s.orderRepository.AssertExpectations(t)
}

func (s *orderTestSuite) TestOrder_CreateOrder_OK() {
	t := s.T()
	request := http.Request{}
	ctx := request.Context()

	description := "description"
	categoryID := int64(3)
	quantity := int64(1)
	rentStart := strfmt.DateTime(time.Now())
	rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	createOrder := &models.OrderCreateRequest{
		Description: &description,
		Category:    &categoryID,
		Quantity:    &quantity,
		RentStart:   &rentStart,
		RentEnd:     &rentEnd,
	}
	userID := 1

	orderToReturn := orderWithAllEdges(t, 1)
	equipment := orderWithEdges(t, 1).Edges.Equipments[0]
	equipmentID := int64(equipment.ID)
	endDate := time.Time(rentEnd).AddDate(0, 0, 1)
	equipmentBookedEndDate := strfmt.DateTime(endDate)
	s.equipmentRepository.On("EquipmentsByFilter", ctx, models.EquipmentFilter{
		Category: categoryID,
	}, math.MaxInt, 0, utils.DescOrder, equipmentEnt.FieldID).Return([]*ent.Equipment{equipment}, nil)
	s.eqStatusRepository.On("HasStatusByPeriod", ctx, domain.EquipmentStatusAvailable, equipment.ID,
		time.Time(rentStart), time.Time(rentEnd)).Return(true, nil)
	s.orderRepository.On("Create", ctx, createOrder, userID, []int{equipment.ID}).Return(orderToReturn, nil)
	s.eqStatusRepository.On("Create", ctx, &models.NewEquipmentStatus{
		EndDate:     &equipmentBookedEndDate,
		EquipmentID: &equipmentID,
		OrderID:     int64(orderToReturn.ID),
		StartDate:   createOrder.RentStart,
		StatusName:  &domain.EquipmentStatusBooked,
	}).Return(nil, nil)

	handlerFunc := s.orderHandler.CreateOrderFunc(s.orderRepository, s.eqStatusRepository, s.equipmentRepository)
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
	orderToReturn := orderWithAllEdges(t, 1)
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

func containsOrder(t *testing.T, list []*ent.Order, order *models.Order) bool {
	t.Helper()
	for _, v := range list {
		if v.ID == int(*order.ID) && v.Description == *order.Description &&
			v.Quantity == int(*order.Quantity) && v.Edges.Users.ID == int(*order.User.ID) &&
			strfmt.DateTime(v.RentStart).String() == order.RentStart.String() &&
			strfmt.DateTime(v.RentEnd).String() == order.RentEnd.String() &&
			v.Edges.Equipments[0].Description == *order.Equipments[0].Description &&
			v.Edges.OrderStatus[0].ID == int(*order.LastStatus.ID) {
			return true
		}
	}
	return false
}

func ordersDuplicated(t *testing.T, array1, array2 []*models.Order) bool {
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
