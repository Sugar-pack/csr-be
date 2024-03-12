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
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/orderstatusname"
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
			Description: "test equipment 1",
		},
		{
			TermsOfUse:  "http://localhost",
			Name:        "equipment 2",
			Title:       "equipment2",
			TechIssue:   false,
			Description: "test equipment 2",
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

	// list of statuses with IDs. Amount of statuses is equal to amount of orders.
	statusNameMap := map[int]string{
		1: domain.OrderStatusInReview,   // active
		2: domain.OrderStatusInProgress, // active
		3: domain.OrderStatusRejected,   // finished
		4: domain.OrderStatusClosed,     // finished
	}
	_, err = s.client.OrderStatusName.Delete().Exec(s.ctx) // clean up
	if err != nil {
		t.Fatal(err)
	}
	for _, statusName := range statusNameMap { // create statuses
		_, errCreation := s.client.OrderStatusName.Create().SetStatus(statusName).Save(s.ctx)
		if errCreation != nil {
			t.Fatal(errCreation)
		}
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
		statusName, err := s.client.OrderStatusName.Query().
			Where(orderstatusname.StatusEQ(statusNameMap[i+1])).Only(s.ctx)
		if err != nil {
			t.Fatal(err)
		}

		o, err := s.client.Order.Create().
			SetDescription(order.Description).
			SetQuantity(order.Quantity).
			SetRentStart(order.RentStart).
			SetRentEnd(order.RentEnd).
			SetUsers(order.Edges.Users).
			SetCurrentStatus(statusName).
			AddEquipments(s.equipments[i%2]).
			AddEquipmentIDs(s.equipments[i%2].ID).
			Save(s.ctx)
		if err != nil {
			t.Fatal(err)
		}
		s.orders[i].ID = o.ID
		s.orders[i].CreatedAt = o.CreatedAt

		_, err = s.client.OrderStatus.Create().
			SetComment("Test order status").
			SetCurrentDate(time.Now()).
			SetOrder(o).
			SetOrderStatusName(statusName).
			SetUsers(s.user).
			SetUsersID(s.user.ID).
			Save(s.ctx)
		if err != nil {
			t.Fatal(err)
		}
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

	filter := domain.OrderFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			Offset:      0,
			OrderBy:     utils.AscOrder,
			OrderColumn: order.FieldID,
		},
	}
	orderList, err := s.orderRepository.List(ctx, &s.user.ID, filter)
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
	totalOrders, err := s.orderRepository.OrdersTotal(ctx, &s.user.ID)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, len(s.orders), totalOrders)
	// Check orders for all users (should be the same amount because of only user)
	totalOrders, err = s.orderRepository.OrdersTotal(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.orders), totalOrders)
}

func (s *OrderSuite) TestOrderRepository_List_EmptyOrderBy() {
	t := s.T()
	filter := domain.OrderFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			Offset:      0,
			OrderBy:     "",
			OrderColumn: order.FieldID,
		},
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)

	orders, err := s.orderRepository.List(ctx, &s.user.ID, filter)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Nil(t, orders)
}

func (s *OrderSuite) TestOrderRepository_List_WrongOrderColumn() {
	t := s.T()
	filter := domain.OrderFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			Offset:      0,
			OrderBy:     utils.AscOrder,
			OrderColumn: order.FieldDescription,
		},
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, &s.user.ID, filter)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Nil(t, orders)
}

func (s *OrderSuite) TestOrderRepository_List_OrderByIDDesc() {
	t := s.T()
	filter := domain.OrderFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			Offset:      0,
			OrderBy:     utils.DescOrder,
			OrderColumn: order.FieldID,
		},
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, &s.user.ID, filter)
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
	filter := domain.OrderFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			Offset:      0,
			OrderBy:     utils.DescOrder,
			OrderColumn: order.FieldRentStart,
		},
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, &s.user.ID, filter)
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
	filter := domain.OrderFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			Offset:      0,
			OrderBy:     utils.AscOrder,
			OrderColumn: order.FieldID,
		},
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, &s.user.ID, filter)
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
	filter := domain.OrderFilter{
		Filter: domain.Filter{
			Limit:       math.MaxInt,
			Offset:      0,
			OrderBy:     utils.AscOrder,
			OrderColumn: order.FieldRentStart,
		},
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, &s.user.ID, filter)
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
	filter := domain.OrderFilter{
		Filter: domain.Filter{
			Limit:       2,
			Offset:      0,
			OrderBy:     utils.AscOrder,
			OrderColumn: order.FieldID,
		},
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, &s.user.ID, filter)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, filter.Limit, len(orders))
	require.Greater(t, len(s.orders), len(orders))
}

func (s *OrderSuite) TestOrderRepository_List_Offset() {
	t := s.T()
	filter := domain.OrderFilter{
		Filter: domain.Filter{
			Limit:       0,
			Offset:      3,
			OrderBy:     utils.AscOrder,
			OrderColumn: order.FieldID,
		},
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.orderRepository.List(ctx, &s.user.ID, filter)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.orders)-filter.Offset, len(orders))
	require.Greater(t, len(s.orders), len(orders))
}

func (s *OrderSuite) TestOrderRepository_List_StatusFilter() {
	t := s.T()
	filter := domain.Filter{
		Limit:       10,
		Offset:      0,
		OrderBy:     utils.AscOrder,
		OrderColumn: order.FieldID,
	}
	tests := map[string]struct {
		fl          domain.OrderFilter
		expectedErr string
		expectedIDs []int // in AscOrder
	}{
		domain.OrderStatusAll: {
			fl: domain.OrderFilter{
				Filter: filter,
				Status: &domain.OrderStatusAll,
			},
			expectedIDs: []int{s.orders[0].ID, s.orders[1].ID, s.orders[2].ID, s.orders[3].ID},
		},
		domain.OrderStatusActive: {
			fl: domain.OrderFilter{
				Filter: filter,
				Status: &domain.OrderStatusActive,
			},
			expectedIDs: []int{s.orders[0].ID, s.orders[1].ID},
		},
		domain.OrderStatusFinished: {
			fl: domain.OrderFilter{
				Filter: filter,
				Status: &domain.OrderStatusFinished,
			},
			expectedIDs: []int{s.orders[2].ID, s.orders[3].ID},
		},
		domain.OrderStatusInReview: {
			fl: domain.OrderFilter{
				Filter: filter,
				Status: &domain.OrderStatusInReview,
			},
			expectedIDs: []int{s.orders[0].ID},
		},
		domain.OrderStatusInProgress: {
			fl: domain.OrderFilter{
				Filter: filter,
				Status: &domain.OrderStatusInProgress,
			},
			expectedIDs: []int{s.orders[1].ID},
		},
		domain.OrderStatusRejected: {
			fl: domain.OrderFilter{
				Filter: filter,
				Status: &domain.OrderStatusRejected,
			},
			expectedIDs: []int{s.orders[2].ID},
		},
		domain.OrderStatusClosed: {
			fl: domain.OrderFilter{
				Filter: filter,
				Status: &domain.OrderStatusClosed,
			},
			expectedIDs: []int{s.orders[3].ID},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := s.ctx
			tx, err := s.client.Tx(ctx)
			require.NoError(t, err)
			ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
			orders, err := s.orderRepository.List(ctx, &s.user.ID, tc.fl)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				require.NoError(t, tx.Rollback())
			} else {
				require.NoError(t, err)
				require.NoError(t, tx.Commit())
				ids := make([]int, 0, len(orders))
				for _, o := range orders {
					ids = append(ids, o.ID)
				}
				require.Equal(t, tc.expectedIDs, ids)
			}
		})
	}
}

func (s *OrderSuite) TestOrderRepository_List_EquipmentFilter() {
	t := s.T()
	filter := domain.Filter{
		Limit:       10,
		Offset:      0,
		OrderBy:     utils.AscOrder,
		OrderColumn: order.FieldID,
	}
	tests := map[string]struct {
		fl          domain.OrderFilter
		expectedIDs []int // in AscOrder
	}{
		domain.OrderStatusAll: {
			fl: domain.OrderFilter{
				Filter: filter,
			},
			expectedIDs: []int{s.orders[0].ID, s.orders[1].ID, s.orders[2].ID, s.orders[3].ID},
		},
		"only Equipment ID 1": {
			fl: domain.OrderFilter{
				Filter:      filter,
				EquipmentID: &s.equipments[0].ID,
			},
			expectedIDs: []int{s.orders[0].ID, s.orders[2].ID},
		},
		"only Equipment ID 2": {
			fl: domain.OrderFilter{
				Filter:      filter,
				EquipmentID: &s.equipments[1].ID,
			},
			expectedIDs: []int{s.orders[1].ID, s.orders[3].ID},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := s.ctx
			tx, err := s.client.Tx(ctx)
			require.NoError(t, err)
			ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
			orders, err := s.orderRepository.List(ctx, &s.user.ID, tc.fl)
			require.NoError(t, err)
			ids := make([]int, 0, len(orders))
			for _, o := range orders {
				ids = append(ids, o.ID)
			}
			require.Equal(t, tc.expectedIDs, ids)
			require.NoError(t, tx.Rollback())
		})
	}
}

func (s *OrderSuite) TestOrderRepository_Update_OK() {
	t := s.T()
	ctx := s.ctx
	crtx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, crtx)

	description := "test"
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
	require.NoError(t, crtx.Commit())

	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	newDesc := "new desc"
	newStartDate := strfmt.DateTime(time.Now().UTC())
	newEndDate := strfmt.DateTime(time.Now().UTC().Add(time.Hour * 24 * 10))
	newQuantity := int64(1)
	req := &models.OrderUpdateRequest{
		Description: &newDesc,
		Quantity:    &newQuantity,
		RentStart:   &newStartDate,
		RentEnd:     &newEndDate,
	}
	updated, err := s.orderRepository.Update(ctx, createdOrder.ID, req, s.user.ID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, newDesc, updated.Description)
	require.Equal(t, newEndDate, strfmt.DateTime(updated.RentEnd))
	require.Equal(t, newStartDate, strfmt.DateTime(updated.RentStart))
}

func (s *OrderSuite) TestOrderRepository_Update_MissingOrder() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)

	newDesc := "new desc"
	newStartDate := strfmt.DateTime(time.Now().UTC())
	newEndDate := strfmt.DateTime(time.Now().UTC().Add(time.Hour * 24 * 10))
	newQuantity := int64(1)
	req := &models.OrderUpdateRequest{
		Description: &newDesc,
		Quantity:    &newQuantity,
		RentStart:   &newStartDate,
		RentEnd:     &newEndDate,
	}
	updated, err := s.orderRepository.Update(ctx, 123, req, s.user.ID)
	require.EqualError(t, err, "ent: order not found")
	require.NoError(t, tx.Rollback())
	require.Nil(t, updated)
}

func (s *OrderSuite) TestOrderRepository_Update_WrongOwner() {
	t := s.T()
	ctx := s.ctx
	crtx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, crtx)

	description := "test"
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
	require.NoError(t, crtx.Commit())

	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	newDesc := "new desc"
	newStartDate := strfmt.DateTime(time.Now().UTC())
	newEndDate := strfmt.DateTime(time.Now().UTC().Add(time.Hour * 24 * 10))
	newQuantity := int64(1)
	req := &models.OrderUpdateRequest{
		Description: &newDesc,
		Quantity:    &newQuantity,
		RentStart:   &newStartDate,
		RentEnd:     &newEndDate,
	}
	updated, err := s.orderRepository.Update(ctx, createdOrder.ID, req, s.user.ID+1)
	require.EqualError(t, err, "permission denied")
	require.NoError(t, tx.Rollback())
	require.Nil(t, updated)
}

func (s *OrderSuite) TestOrderRepository_getQuantity() {
	t := s.T()
	valid := int(1)
	tests := map[string]struct {
		q         int
		maxQ      int
		res       *int
		errString string
	}{
		"ok": {
			q:         1,
			maxQ:      2,
			res:       &valid,
			errString: "",
		},
		"nok": {
			q:         2,
			maxQ:      1,
			res:       nil,
			errString: "quantity limit exceeded: 1 allowed",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := getQuantity(tc.q, tc.maxQ)
			if tc.errString != "" {
				require.EqualError(t, err, tc.errString)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.res, res)
			}
		})
	}
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
