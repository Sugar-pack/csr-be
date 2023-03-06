package repositories

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type OrderSuite struct {
	suite.Suite
	ctx                   context.Context
	orderRepository       domain.OrderRepository
	orderStatusRepository *orderStatusRepository
	client                *ent.Client
	orders                []*ent.Order
	user                  *ent.User
	equipments            []*ent.Equipment
}

func TestOrdersSuite(t *testing.T) {
	s := new(OrderSuite)
	suite.Run(t, s)
}

func (s *OrderSuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:orders?mode=memory&cache=shared&_fk=1")
	s.client = client
	s.orderRepository = NewOrderRepository()
	s.orderStatusRepository = NewOrderStatusRepository()

	s.user = &ent.User{
		Login: "user1", Email: "user1@email.com", Password: "1234", Name: "user1",
	}
	_, err := s.client.User.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	u, err := s.client.User.Create().
		SetLogin(s.user.Login).SetEmail(s.user.Email).
		SetPassword(s.user.Password).SetName(s.user.Name).
		Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	s.user = u

	category := &ent.Category{
		Name:                "Клетка",
		MaxReservationTime:  10 * 60 * 60 * 24,
		MaxReservationUnits: 10,
		HasSubcategory:      false,
	}
	_, err = s.client.Category.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	cat, err := s.client.Category.Create().SetName(category.Name).
		SetMaxReservationTime(category.MaxReservationTime).
		SetMaxReservationUnits(category.MaxReservationUnits).
		SetHasSubcategory(category.HasSubcategory).
		Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	s.equipments = []*ent.Equipment{
		{
			TermsOfUse:  "http://localhost",
			Name:        "equipment 1",
			Title:       "equipment1",
			TechIssue:   false,
			Description: "test equipment",
		},
	}
	_, err = s.client.Equipment.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i, eq := range s.equipments {
		e, err := s.client.Equipment.Create().SetName(eq.Name).
			SetTitle(eq.Title).SetTermsOfUse(eq.TermsOfUse).
			SetTechIssue(eq.TechIssue).SetDescription(eq.Description).
			SetCategory(cat).
			Save(s.ctx)
		if err != nil {
			t.Fatal(err)
		}
		s.equipments[i] = e
	}

	s.orders = []*ent.Order{
		{
			Quantity:  1,
			RentStart: time.Date(2022, time.January, 1, 12, 0, 0, 0, time.Local),
			RentEnd:   time.Date(2022, time.January, 10, 12, 0, 0, 0, time.Local),
			Edges: ent.OrderEdges{
				Users: s.user,
			},
		},
		{
			Quantity:  1,
			RentStart: time.Date(2022, time.January, 1, 12, 0, 0, 0, time.Local),
			RentEnd:   time.Date(2022, time.February, 10, 12, 0, 0, 0, time.Local),
			Edges: ent.OrderEdges{
				Users: s.user,
			},
		},
		{
			Quantity:  2,
			RentStart: time.Date(2022, time.February, 1, 12, 0, 0, 0, time.Local),
			RentEnd:   time.Date(2022, time.February, 10, 12, 0, 0, 0, time.Local),
			Edges: ent.OrderEdges{
				Users: s.user,
			},
		},
		{
			Quantity:  1,
			RentStart: time.Date(2022, time.March, 1, 12, 0, 0, 0, time.Local),
			RentEnd:   time.Date(2022, time.March, 10, 12, 0, 0, 0, time.Local),
			Edges: ent.OrderEdges{
				Users: s.user,
			},
		},
	}

	_, err = s.client.Order.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i, order := range s.orders {
		o, err := s.client.Order.Create().
			SetDescription(order.Description).
			SetQuantity(order.Quantity).
			SetRentStart(order.RentStart).
			SetRentEnd(order.RentEnd).
			SetUsers(order.Edges.Users).
			Save(s.ctx)
		if err != nil {
			t.Fatal(err)
		}
		s.orders[i].ID = o.ID
		s.orders[i].CreatedAt = o.CreatedAt
	}

	_, err = s.client.OrderStatusName.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.client.OrderStatusName.Create().SetStatus(domain.OrderStatusInReview).Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func (s *OrderSuite) TearDownSuite() {
	s.client.Close()
}

func (s *OrderSuite) TestOrderRepository_Create_EmptyEquipments() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)

	description := "test"
	startDate := strfmt.DateTime(time.Now().UTC())
	endDate := strfmt.DateTime(time.Now().UTC().Add(time.Hour * 24 * 5))
	data := &models.OrderCreateRequest{
		Description: description,
		RentEnd:     &endDate,
		RentStart:   &startDate,
	}
	createdOrder, err := s.orderRepository.Create(ctx, data, s.user.ID, []int{})
	require.Error(t, err)
	require.Nil(t, createdOrder)
}

func (s *OrderSuite) TestOrderRepository_Create_OK() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)

	description := "test"
	equipmentID := int64(s.equipments[0].ID)
	eqID := int64(1)
	startDate := strfmt.DateTime(time.Now().UTC())
	endDate := strfmt.DateTime(time.Now().UTC().Add(time.Hour * 24 * 5))
	data := &models.OrderCreateRequest{
		Description: description,
		EquipmentID: &eqID,
		RentEnd:     &endDate,
		RentStart:   &startDate,
	}
	createdOrder, err := s.orderRepository.Create(ctx, data, s.user.ID, []int{s.equipments[0].ID})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NotEmpty(t, createdOrder)
	require.Equal(t, description, createdOrder.Description)
	require.NotEmpty(t, createdOrder.Edges.Equipments)
	require.Equal(t, int(equipmentID), createdOrder.Edges.Equipments[0].ID)
	require.NotEmpty(t, createdOrder.Edges.Users)
	require.Equal(t, s.user.ID, createdOrder.Edges.Users.ID)
	require.NotEmpty(t, createdOrder.Edges.OrderStatus)
	require.Equal(t, domain.OrderStatusInReview, createdOrder.Edges.OrderStatus[0].Edges.OrderStatusName.Status)
	require.Equal(t, true, createdOrder.IsFirst)
}

// isFirst field should be false, if one of orders has approved status
func (s *OrderSuite) TestOrderRepository_Create_isFirstCreatedOrderIsFalseIfOneOfOrdersHasApprovedStatus() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)

	description := "test"
	equipmentID := int64(s.equipments[0].ID)
	startDate := strfmt.DateTime(time.Now().UTC())
	endDate := strfmt.DateTime(time.Now().UTC().Add(time.Hour * 24 * 5))
	data := &models.OrderCreateRequest{
		Description: description,
		EquipmentID: &equipmentID,
		RentEnd:     &endDate,
		RentStart:   &startDate,
	}

	err = s.client.OrderStatusName.Create().
		SetStatus(domain.OrderStatusApproved).
		Exec(ctx)
	require.NoError(t, err)
	orderId := int64(s.orders[0].ID)
	testComment := "testComment"
	model := models.NewOrderStatus{
		CreatedAt: &startDate,
		OrderID:   &orderId,
		Status:    &domain.OrderStatusApproved,
		Comment:   &testComment,
	}
	err = s.orderStatusRepository.UpdateStatus(ctx, s.user.ID, model)
	require.NoError(t, err)

	createdOrder, err := s.orderRepository.Create(ctx, data, s.user.ID, []int{s.equipments[0].ID})
	require.NoError(t, err)

	require.NotEmpty(t, createdOrder)
	require.Equal(t, description, createdOrder.Description)
	require.NotEmpty(t, createdOrder.Edges.Equipments)
	require.Equal(t, int(equipmentID), createdOrder.Edges.Equipments[0].ID)
	require.NotEmpty(t, createdOrder.Edges.Users)
	require.Equal(t, s.user.ID, createdOrder.Edges.Users.ID)
	require.NotEmpty(t, createdOrder.Edges.OrderStatus)
	require.Equal(t, domain.OrderStatusInReview, createdOrder.Edges.OrderStatus[0].Edges.OrderStatusName.Status)
	require.Equal(t, false, createdOrder.IsFirst)

	require.NoError(t, tx.Commit())
}

// isFirst field should be true, if one of orders has rejected status
func (s *OrderSuite) TestOrderRepository_Create_isFirstCreatedOrderIsTrueIfOneOfOrdersHasRejectedStatus() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)

	description := "test"
	equipmentID := int64(s.equipments[0].ID)
	startDate := strfmt.DateTime(time.Now().UTC())
	endDate := strfmt.DateTime(time.Now().UTC().Add(time.Hour * 24 * 5))
	data := &models.OrderCreateRequest{
		Description: description,
		EquipmentID: &equipmentID,
		RentEnd:     &endDate,
		RentStart:   &startDate,
	}

	err = s.client.OrderStatusName.Create().
		SetStatus(domain.OrderStatusRejected).
		Exec(ctx)
	require.NoError(t, err)
	orderId := int64(s.orders[0].ID)
	testComment := "testComment"
	model := models.NewOrderStatus{
		CreatedAt: &startDate,
		OrderID:   &orderId,
		Status:    &domain.OrderStatusRejected,
		Comment:   &testComment,
	}
	err = s.orderStatusRepository.UpdateStatus(ctx, s.user.ID, model)
	require.NoError(t, err)

	createdOrder, err := s.orderRepository.Create(ctx, data, s.user.ID, []int{s.equipments[0].ID})
	require.NoError(t, err)

	require.NotEmpty(t, createdOrder)
	require.Equal(t, description, createdOrder.Description)
	require.NotEmpty(t, createdOrder.Edges.Equipments)
	require.Equal(t, int(equipmentID), createdOrder.Edges.Equipments[0].ID)
	require.NotEmpty(t, createdOrder.Edges.Users)
	require.Equal(t, s.user.ID, createdOrder.Edges.Users.ID)
	require.NotEmpty(t, createdOrder.Edges.OrderStatus)
	require.Equal(t, domain.OrderStatusInReview, createdOrder.Edges.OrderStatus[0].Edges.OrderStatusName.Status)
	require.Equal(t, true, createdOrder.IsFirst)

	require.NoError(t, tx.Commit())
}

// isFirst field should be true for all new created orders
func (s *OrderSuite) TestOrderRepository_Create_isFirstFieldIsTrueForSeveralNewCreatedOrders() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)

	description := "test"
	equipmentID := int64(s.equipments[0].ID)
	startDate := strfmt.DateTime(time.Now().UTC())
	endDate := strfmt.DateTime(time.Now().UTC().Add(time.Hour * 24 * 5))
	data := &models.OrderCreateRequest{
		Description: description,
		EquipmentID: &equipmentID,
		RentEnd:     &endDate,
		RentStart:   &startDate,
	}

	createdFirstOrder, err := s.orderRepository.Create(ctx, data, s.user.ID, []int{s.equipments[0].ID})
	require.NoError(t, err)

	require.NotEmpty(t, createdFirstOrder)
	require.Equal(t, description, createdFirstOrder.Description)
	require.NotEmpty(t, createdFirstOrder.Edges.Equipments)
	require.Equal(t, int(equipmentID), createdFirstOrder.Edges.Equipments[0].ID)
	require.NotEmpty(t, createdFirstOrder.Edges.Users)
	require.Equal(t, s.user.ID, createdFirstOrder.Edges.Users.ID)
	require.NotEmpty(t, createdFirstOrder.Edges.OrderStatus)
	require.Equal(t, domain.OrderStatusInReview, createdFirstOrder.Edges.OrderStatus[0].Edges.OrderStatusName.Status)
	require.Equal(t, true, createdFirstOrder.IsFirst)

	createdSecondOrder, err := s.orderRepository.Create(ctx, data, s.user.ID, []int{s.equipments[0].ID})
	require.NoError(t, err)

	require.NotEmpty(t, createdSecondOrder)
	require.Equal(t, description, createdSecondOrder.Description)
	require.NotEmpty(t, createdSecondOrder.Edges.Equipments)
	require.Equal(t, int(equipmentID), createdSecondOrder.Edges.Equipments[0].ID)
	require.NotEmpty(t, createdSecondOrder.Edges.Users)
	require.Equal(t, s.user.ID, createdSecondOrder.Edges.Users.ID)
	require.NotEmpty(t, createdSecondOrder.Edges.OrderStatus)
	require.Equal(t, domain.OrderStatusInReview, createdSecondOrder.Edges.OrderStatus[0].Edges.OrderStatusName.Status)
	require.Equal(t, true, createdSecondOrder.IsFirst)

	require.NoError(t, tx.Commit())
}

// isFirst field should be false for all previous created orders if one of them was approved
func (s *OrderSuite) TestOrderRepository_Create_isFirstFieldForPreviousCreatedOrdersIsFalseIfOneOfOrdersApproved() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)

	description := "test"
	equipmentID := int64(s.equipments[0].ID)
	startDate := strfmt.DateTime(time.Now().UTC())
	endDate := strfmt.DateTime(time.Now().UTC().Add(time.Hour * 24 * 5))
	data := &models.OrderCreateRequest{
		Description: description,
		EquipmentID: &equipmentID,
		RentEnd:     &endDate,
		RentStart:   &startDate,
	}

	orderId := int64(s.orders[0].ID)
	testComment := "testComment"
	model := models.NewOrderStatus{
		CreatedAt: &startDate,
		OrderID:   &orderId,
		Status:    &domain.OrderStatusApproved,
		Comment:   &testComment,
	}

	err = s.client.OrderStatusName.Create().
		SetStatus(domain.OrderStatusApproved).
		Exec(ctx)
	require.NoError(t, err)

	createdFirstOrder, err := s.orderRepository.Create(ctx, data, s.user.ID, []int{s.equipments[0].ID})
	require.NoError(t, err)

	require.NotEmpty(t, createdFirstOrder)
	require.Equal(t, description, createdFirstOrder.Description)
	require.NotEmpty(t, createdFirstOrder.Edges.Equipments)
	require.Equal(t, int(equipmentID), createdFirstOrder.Edges.Equipments[0].ID)
	require.NotEmpty(t, createdFirstOrder.Edges.Users)
	require.Equal(t, s.user.ID, createdFirstOrder.Edges.Users.ID)
	require.NotEmpty(t, createdFirstOrder.Edges.OrderStatus)
	require.Equal(t, domain.OrderStatusInReview, createdFirstOrder.Edges.OrderStatus[0].Edges.OrderStatusName.Status)
	require.Equal(t, true, createdFirstOrder.IsFirst)

	createdSecondOrder, err := s.orderRepository.Create(ctx, data, s.user.ID, []int{s.equipments[0].ID})
	require.NoError(t, err)

	require.NotEmpty(t, createdSecondOrder)
	require.Equal(t, description, createdSecondOrder.Description)
	require.NotEmpty(t, createdSecondOrder.Edges.Equipments)
	require.Equal(t, int(equipmentID), createdSecondOrder.Edges.Equipments[0].ID)
	require.NotEmpty(t, createdSecondOrder.Edges.Users)
	require.Equal(t, s.user.ID, createdSecondOrder.Edges.Users.ID)
	require.NotEmpty(t, createdSecondOrder.Edges.OrderStatus)
	require.Equal(t, domain.OrderStatusInReview, createdSecondOrder.Edges.OrderStatus[0].Edges.OrderStatusName.Status)
	require.Equal(t, true, createdSecondOrder.IsFirst)

	err = s.orderStatusRepository.UpdateStatus(ctx, s.user.ID, model)
	require.NoError(t, err)

	orderList, err := s.orderRepository.List(ctx, s.user.ID, math.MaxInt, 0, utils.AscOrder, order.FieldID)
	require.NoError(t, err)

	for _, order := range orderList {
		require.Equal(t, false, order.IsFirst)
	}

	require.NoError(t, tx.Commit())

}

func (s *OrderSuite) TestOrderRepository_OrdersTotal() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	totalOrders, err := s.orderRepository.OrdersTotal(ctx, s.user.ID)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.orders), totalOrders)
}

func (s *OrderSuite) TestOrderRepository_List_EmptyOrderBy() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := ""
	orderColumn := order.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, s.user.ID, limit, offset, orderBy, orderColumn)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Nil(t, orders)
}

func (s *OrderSuite) TestOrderRepository_List_WrongOrderColumn() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldDescription
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, s.user.ID, limit, offset, orderBy, orderColumn)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Nil(t, orders)
}

func (s *OrderSuite) TestOrderRepository_List_OrderByIDDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := order.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, s.user.ID, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.orders), len(orders))
	prevOrderID := math.MaxInt
	for _, value := range orders {
		require.True(t, containsOrder(t, value, s.orders))
		require.Less(t, value.ID, prevOrderID)
		prevOrderID = value.ID
	}
}

func (s *OrderSuite) TestOrderRepository_List_OrderByRentStartDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := order.FieldRentStart
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, s.user.ID, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.orders), len(orders))
	prevOrderRentStart := time.Unix(1<<63-62135596801, 999999999)
	for _, value := range orders {
		require.True(t, containsOrder(t, value, s.orders))
		require.True(t, value.RentStart.Before(prevOrderRentStart) || value.RentStart.Equal(prevOrderRentStart))
		prevOrderRentStart = value.RentStart
	}
}

func (s *OrderSuite) TestOrderRepository_List_OrderByIDAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, s.user.ID, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.orders), len(orders))
	prevOrderID := 0
	for _, value := range orders {
		require.True(t, containsOrder(t, value, s.orders))
		require.Greater(t, value.ID, prevOrderID)
		prevOrderID = value.ID
	}
}

func (s *OrderSuite) TestOrderRepository_List_OrderByRentStartAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldRentStart
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, s.user.ID, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.orders), len(orders))
	prevOrderRentStart := time.Unix(0, 0)
	for _, value := range orders {
		require.True(t, containsOrder(t, value, s.orders))
		require.True(t, value.RentStart.After(prevOrderRentStart) || value.RentStart.Equal(prevOrderRentStart))
		prevOrderRentStart = value.RentStart
	}
}

func (s *OrderSuite) TestOrderRepository_List_Limit() {
	t := s.T()
	limit := 2
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, s.user.ID, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, limit, len(orders))
	require.Greater(t, len(s.orders), len(orders))
}

func (s *OrderSuite) TestOrderRepository_List_Offset() {
	t := s.T()
	limit := 0
	offset := 3
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, s.user.ID, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.orders)-offset, len(orders))
	require.Greater(t, len(s.orders), len(orders))
}

func containsOrder(t *testing.T, order *ent.Order, orders []*ent.Order) bool {
	t.Helper()
	for _, o := range orders {
		if order.ID == o.ID && order.RentStart.Equal(o.RentStart) &&
			order.Quantity == o.Quantity && order.CreatedAt.Equal(o.CreatedAt) {
			return true
		}
	}
	return false
}
