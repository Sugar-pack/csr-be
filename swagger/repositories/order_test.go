package repositories

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type OrderSuite struct {
	suite.Suite
	ctx        context.Context
	repository OrderRepository
	client     *ent.Client
	orders     []*ent.Order
	users      []*ent.User
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
	s.repository = NewOrderRepository()

	s.users = []*ent.User{
		{Login: "user1", Email: "user1@email.com", Password: "1234", Name: "user1"},
	}
	_, err := s.client.User.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i, user := range s.users {
		u, err := s.client.User.Create().
			SetLogin(user.Login).SetEmail(user.Email).
			SetPassword(user.Password).SetName(user.Name).
			Save(s.ctx)
		if err != nil {
			t.Fatal(err)
		}
		s.users[i] = u
	}

	s.orders = []*ent.Order{
		{
			Quantity:  1,
			RentStart: time.Date(2022, time.January, 1, 12, 0, 0, 0, time.Local),
			RentEnd:   time.Date(2022, time.January, 10, 12, 0, 0, 0, time.Local),
			Edges: ent.OrderEdges{
				Users: s.users,
			},
		},
		{
			Quantity:  1,
			RentStart: time.Date(2022, time.January, 1, 12, 0, 0, 0, time.Local),
			RentEnd:   time.Date(2022, time.February, 10, 12, 0, 0, 0, time.Local),
			Edges: ent.OrderEdges{
				Users: s.users,
			},
		},
		{
			Quantity:  2,
			RentStart: time.Date(2022, time.February, 1, 12, 0, 0, 0, time.Local),
			RentEnd:   time.Date(2022, time.February, 10, 12, 0, 0, 0, time.Local),
			Edges: ent.OrderEdges{
				Users: s.users,
			},
		},
		{
			Quantity:  1,
			RentStart: time.Date(2022, time.March, 1, 12, 0, 0, 0, time.Local),
			RentEnd:   time.Date(2022, time.March, 10, 12, 0, 0, 0, time.Local),
			Edges: ent.OrderEdges{
				Users: s.users,
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
			AddUsers(order.Edges.Users[0]).
			Save(s.ctx)
		if err != nil {
			t.Fatal(err)
		}
		s.orders[i].ID = o.ID
		s.orders[i].CreatedAt = o.CreatedAt
	}
}

func (s *OrderSuite) TearDownSuite() {
	s.client.Close()
}

func (s *OrderSuite) TestOrderRepository_OrdersTotal() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	totalOrders, err := s.repository.OrdersTotal(ctx, s.users[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.orders), totalOrders)
}

func (s *OrderSuite) TestOrderRepository_List_EmptyOrderBy() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := ""
	orderColumn := order.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.List(ctx, s.users[0].ID, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Nil(t, orders)
}

func (s *OrderSuite) TestOrderRepository_List_WrongOrderColumn() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldDescription
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.List(ctx, s.users[0].ID, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Nil(t, orders)
}

func (s *OrderSuite) TestOrderRepository_List_OrderByIDDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := order.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.List(ctx, s.users[0].ID, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.orders), len(orders))
	prevOrderID := math.MaxInt
	for _, value := range orders {
		assert.True(t, containsOrder(t, value, s.orders))
		assert.Less(t, value.ID, prevOrderID)
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
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.List(ctx, s.users[0].ID, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.orders), len(orders))
	prevOrderRentStart := time.Unix(1<<63-62135596801, 999999999)
	for _, value := range orders {
		assert.True(t, containsOrder(t, value, s.orders))
		assert.True(t, value.RentStart.Before(prevOrderRentStart) || value.RentStart.Equal(prevOrderRentStart))
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
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.List(ctx, s.users[0].ID, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.orders), len(orders))
	prevOrderID := 0
	for _, value := range orders {
		assert.True(t, containsOrder(t, value, s.orders))
		assert.Greater(t, value.ID, prevOrderID)
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
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.List(ctx, s.users[0].ID, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.orders), len(orders))
	prevOrderRentStart := time.Unix(0, 0)
	for _, value := range orders {
		assert.True(t, containsOrder(t, value, s.orders))
		assert.True(t, value.RentStart.After(prevOrderRentStart) || value.RentStart.Equal(prevOrderRentStart))
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
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.List(ctx, s.users[0].ID, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, limit, len(orders))
	assert.Greater(t, len(s.orders), len(orders))
}

func (s *OrderSuite) TestOrderRepository_List_Offset() {
	t := s.T()
	limit := 0
	offset := 3
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.List(ctx, s.users[0].ID, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.orders)-offset, len(orders))
	assert.Greater(t, len(s.orders), len(orders))
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
