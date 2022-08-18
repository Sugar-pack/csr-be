package repositories

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
)

type orderFilterTestSuite struct {
	suite.Suite
	ctx           context.Context
	client        *ent.Client
	repository    OrderRepositoryWithFilter
	orders        []*ent.Order
	statusesNames []*ent.StatusName
}

func TestOrderFilterSuite(t *testing.T) {
	s := new(orderFilterTestSuite)
	suite.Run(t, s)
}

func (s *orderFilterTestSuite) validOrder(id int) *ent.Order {
	s.T().Helper()
	return &ent.Order{
		ID:          id,
		Description: fmt.Sprintf("description %d", id),
		Quantity:    id%2 + 1,
		RentStart:   time.Now().Add(time.Duration(-id*24) * time.Hour),
		RentEnd:     time.Now().Add(time.Duration(id*24) * time.Hour),
		CreatedAt:   time.Now().Add(time.Duration(-id) * time.Hour),
		Edges: ent.OrderEdges{
			OrderStatus: []*ent.OrderStatus{
				{
					Comment:     fmt.Sprintf("updated status %d", id),
					CurrentDate: time.Now().Add(time.Duration(-id) * time.Hour),
					Edges:       ent.OrderStatusEdges{StatusName: s.statusesNames[id%2]},
				},
			},
		},
	}
}

func (s *orderFilterTestSuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:orderfilter?mode=memory&cache=shared&_fk=1")
	s.client = client
	s.repository = NewOrderFilter(s.client)

	if _, err := s.client.StatusName.Delete().Exec(s.ctx); err != nil {
		t.Fatal(err)
	}
	s.statusesNames = []*ent.StatusName{
		{
			Status: "review",
		},
		{
			Status: "approved",
		},
		{
			Status: "in progress",
		},
		{
			Status: "rejected",
		},
		{
			Status: "closed",
		},
	}
	for i, statusName := range s.statusesNames {
		sName, err := s.client.StatusName.Create().SetStatus(statusName.Status).Save(s.ctx)
		if err != nil {
			t.Fatal(err)
		}
		s.statusesNames[i].ID = sName.ID
	}

	if _, err := s.client.Order.Delete().Exec(s.ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := s.client.OrderStatus.Delete().Exec(s.ctx); err != nil {
		t.Fatal(err)
	}
	s.orders = []*ent.Order{
		s.validOrder(1),
		s.validOrder(2),
		s.validOrder(3),
		s.validOrder(4),
		s.validOrder(5),
	}
	for i, order := range s.orders {
		o, err := s.client.Order.Create().
			SetDescription(order.Description).
			SetQuantity(order.Quantity).
			SetRentStart(order.RentStart).
			SetRentEnd(order.RentEnd).
			SetCreatedAt(order.CreatedAt).
			Save(s.ctx)
		if err != nil {
			t.Fatal(err)
		}
		s.orders[i].ID = o.ID
		s.orders[i].CreatedAt = o.CreatedAt

		orderStatus, err := s.client.OrderStatus.Create().
			SetComment(order.Edges.OrderStatus[0].Comment).
			SetCurrentDate(order.Edges.OrderStatus[0].CurrentDate).
			SetOrder(o).
			SetStatusName(order.Edges.OrderStatus[0].Edges.StatusName).
			Save(s.ctx)
		if err != nil {
			t.Fatal(err)
		}
		s.orders[i].Edges.OrderStatus[0].ID = orderStatus.ID
	}

	s.repository = NewOrderFilter(client)
}

func (s *orderFilterTestSuite) TearDownSuite() {
	s.client.Close()
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatusTotal() {
	t := s.T()

	status := s.statusesNames[0].Status
	totalOrders, err := s.repository.OrdersByStatusTotal(s.ctx, status)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, totalOrders)

	status = s.statusesNames[1].Status
	totalOrders, err = s.repository.OrdersByStatusTotal(s.ctx, status)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, totalOrders)
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatus_EmptyOrderBy() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := ""
	orderColumn := order.FieldID
	status := s.statusesNames[0].Status
	orders, err := s.repository.OrdersByStatus(s.ctx, status, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.Nil(t, orders)
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatus_WrongOrderColumn() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldQuantity
	status := s.statusesNames[0].Status
	orders, err := s.repository.OrdersByStatus(s.ctx, status, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.Nil(t, orders)
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatus_OrderByIDDesc() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := order.FieldID
	status := s.statusesNames[1].Status
	orders, err := s.repository.OrdersByStatus(s.ctx, status, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(orders))
	prevOrderID := math.MaxInt
	for _, o := range orders {
		assert.True(t, containsOrder(t, o, s.orders))
		assert.GreaterOrEqual(t, prevOrderID, o.ID)
		prevOrderID = o.ID
	}
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatus_OrderByCreatedAtDesc() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := order.FieldCreatedAt
	status := s.statusesNames[1].Status
	orders, err := s.repository.OrdersByStatus(s.ctx, status, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(orders))
	prevOrderCreatedAt := time.Unix(1<<63-62135596801, 999999999)
	for _, o := range orders {
		assert.True(t, containsOrder(t, o, s.orders))
		assert.True(t, prevOrderCreatedAt.After(o.CreatedAt) || prevOrderCreatedAt.Equal(o.CreatedAt))
		prevOrderCreatedAt = o.CreatedAt
	}
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatus_OrderByIDAsc() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	status := s.statusesNames[1].Status
	orders, err := s.repository.OrdersByStatus(s.ctx, status, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(orders))
	prevOrderID := 0
	for _, o := range orders {
		assert.True(t, containsOrder(t, o, s.orders))
		assert.LessOrEqual(t, prevOrderID, o.ID)
		prevOrderID = o.ID
	}
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatus_OrderByCreatedAtAsc() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldCreatedAt
	status := s.statusesNames[1].Status
	orders, err := s.repository.OrdersByStatus(s.ctx, status, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(orders))
	prevOrderCreatedAt := time.Unix(0, 0)
	for _, o := range orders {
		assert.True(t, containsOrder(t, o, s.orders))
		assert.True(t, prevOrderCreatedAt.Before(o.CreatedAt) || prevOrderCreatedAt.Equal(o.CreatedAt))
		prevOrderCreatedAt = o.CreatedAt
	}
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatus_Limit() {
	t := s.T()

	limit := 2
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldCreatedAt
	status := s.statusesNames[1].Status
	orders, err := s.repository.OrdersByStatus(s.ctx, status, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, limit, len(orders))
	assert.Greater(t, len(s.orders), len(orders))
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatus_Offset() {
	t := s.T()

	limit := 10
	offset := 2
	orderBy := utils.AscOrder
	orderColumn := order.FieldCreatedAt
	status := s.statusesNames[1].Status
	orders, err := s.repository.OrdersByStatus(s.ctx, status, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(s.orders)-offset, len(orders))
	assert.Greater(t, len(s.orders), len(orders))
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByPeriodAndStatusTotal() {
	t := s.T()

	status := s.statusesNames[0].Status
	from := time.Now().Add(-2*time.Hour - time.Minute)
	to := time.Now().Add(3 * time.Hour)
	totalOrders, err := s.repository.OrdersByPeriodAndStatusTotal(s.ctx, from, to, status)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, totalOrders)

	status = s.statusesNames[1].Status
	from = time.Now().Add(-6 * time.Hour)
	to = time.Now().Add(6 * time.Hour)
	totalOrders, err = s.repository.OrdersByPeriodAndStatusTotal(s.ctx, from, to, status)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, totalOrders)
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByPeriodAndStatus_EmptyOrderBy() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := ""
	orderColumn := order.FieldID
	status := s.statusesNames[0].Status
	from := time.Now().Add(-2*time.Hour - time.Minute)
	to := time.Now().Add(3 * time.Hour)
	orders, err := s.repository.OrdersByPeriodAndStatus(s.ctx, from, to, status, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.Nil(t, orders)
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByPeriodAndStatus_WrongOrderColumn() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldQuantity
	status := s.statusesNames[0].Status
	from := time.Now().Add(-2*time.Hour - time.Minute)
	to := time.Now().Add(3 * time.Hour)
	orders, err := s.repository.OrdersByPeriodAndStatus(s.ctx, from, to, status, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.Nil(t, orders)
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByPeriodAndStatus_OrderByIDDesc() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := order.FieldID
	status := s.statusesNames[1].Status
	from := time.Now().Add(-3*time.Hour - time.Minute)
	to := time.Now().Add(3 * time.Hour)
	orders, err := s.repository.OrdersByPeriodAndStatus(s.ctx, from, to, status, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(orders))
	prevOrderID := math.MaxInt
	for _, o := range orders {
		assert.True(t, containsOrder(t, o, s.orders))
		assert.GreaterOrEqual(t, prevOrderID, o.ID)
		prevOrderID = o.ID
	}
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByPeriodAndStatus_OrderByCreatedAtDesc() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := order.FieldCreatedAt
	status := s.statusesNames[1].Status
	from := time.Now().Add(-3*time.Hour - time.Minute)
	to := time.Now().Add(3 * time.Hour)
	orders, err := s.repository.OrdersByPeriodAndStatus(s.ctx, from, to, status, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(orders))
	prevOrderCreatedAt := time.Unix(1<<63-62135596801, 999999999)
	for _, o := range orders {
		assert.True(t, containsOrder(t, o, s.orders))
		assert.True(t, prevOrderCreatedAt.After(o.CreatedAt) || prevOrderCreatedAt.Equal(o.CreatedAt))
		prevOrderCreatedAt = o.CreatedAt
	}
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByPeriodAndStatus_OrderByIDAsc() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldID
	status := s.statusesNames[1].Status
	from := time.Now().Add(-4*time.Hour - time.Minute)
	to := time.Now().Add(3 * time.Hour)
	orders, err := s.repository.OrdersByPeriodAndStatus(s.ctx, from, to, status, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(orders))
	prevOrderID := 0
	for _, o := range orders {
		assert.True(t, containsOrder(t, o, s.orders))
		assert.LessOrEqual(t, prevOrderID, o.ID)
		prevOrderID = o.ID
	}
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByPeriodAndStatus_OrderByCreatedAtAsc() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldCreatedAt
	status := s.statusesNames[1].Status
	from := time.Now().Add(-4*time.Hour - time.Minute)
	to := time.Now().Add(3 * time.Hour)
	orders, err := s.repository.OrdersByPeriodAndStatus(s.ctx, from, to, status, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(orders))
	prevOrderCreatedAt := time.Unix(0, 0)
	for _, o := range orders {
		assert.True(t, containsOrder(t, o, s.orders))
		assert.True(t, prevOrderCreatedAt.Before(o.CreatedAt) || prevOrderCreatedAt.Equal(o.CreatedAt))
		prevOrderCreatedAt = o.CreatedAt
	}
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByPeriodAndStatus_Limit() {
	t := s.T()

	limit := 1
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldCreatedAt
	status := s.statusesNames[1].Status
	from := time.Now().Add(-4*time.Hour - time.Minute)
	to := time.Now().Add(3 * time.Hour)
	orders, err := s.repository.OrdersByPeriodAndStatus(s.ctx, from, to, status, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, limit, len(orders))
	assert.Greater(t, len(s.orders), len(orders))
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByPeriodAndStatus_Offset() {
	t := s.T()

	limit := 10
	offset := 1
	orderBy := utils.AscOrder
	orderColumn := order.FieldCreatedAt
	status := s.statusesNames[1].Status
	from := time.Now().Add(-4*time.Hour - time.Minute)
	to := time.Now().Add(3 * time.Hour)
	orders, err := s.repository.OrdersByPeriodAndStatus(s.ctx, from, to, status, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(s.orders)-offset, len(orders))
	assert.Greater(t, len(s.orders), len(orders))
}
