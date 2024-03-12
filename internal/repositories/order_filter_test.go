package repositories

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type orderFilterTestSuite struct {
	suite.Suite
	ctx           context.Context
	client        *ent.Client
	repository    domain.OrderRepositoryWithFilter
	orders        []*ent.Order
	statusesNames []*ent.OrderStatusName
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
					Edges:       ent.OrderStatusEdges{OrderStatusName: s.statusesNames[id%2]},
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
	s.repository = NewOrderFilter()

	if _, err := s.client.OrderStatusName.Delete().Exec(s.ctx); err != nil {
		t.Fatal(err)
	}
	s.statusesNames = []*ent.OrderStatusName{
		{Status: domain.OrderStatusInReview},
		{Status: domain.OrderStatusApproved},
		{Status: domain.OrderStatusInProgress},
		{Status: domain.OrderStatusRejected},
		{Status: domain.OrderStatusClosed},
	}
	for i, statusName := range s.statusesNames {
		sName, err := s.client.OrderStatusName.Create().SetStatus(statusName.Status).Save(s.ctx)
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
			SetOrderStatusName(order.Edges.OrderStatus[0].Edges.OrderStatusName).
			Save(s.ctx)
		if err != nil {
			t.Fatal(err)
		}
		s.orders[i].Edges.OrderStatus[0].ID = orderStatus.ID
	}

	s.repository = NewOrderFilter()
}

func (s *orderFilterTestSuite) TearDownSuite() {
	s.client.Close()
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatusTotal() {
	t := s.T()

	status := s.statusesNames[0].Status
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	totalOrders, err := s.repository.OrdersByStatusTotal(ctx, status)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 2, totalOrders)

	status = s.statusesNames[1].Status
	totalOrders, err = s.repository.OrdersByStatusTotal(ctx, status)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, 3, totalOrders)
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatus_EmptyOrderBy() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := ""
	orderColumn := order.FieldID
	status := s.statusesNames[0].Status
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByStatus(ctx, status, limit, offset, orderBy, orderColumn)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Nil(t, orders)
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatus_WrongOrderColumn() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := order.FieldQuantity
	status := s.statusesNames[0].Status
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByStatus(ctx, status, limit, offset, orderBy, orderColumn)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Nil(t, orders)
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatus_OrderByIDDesc() {
	t := s.T()

	limit := 10
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := order.FieldID
	status := s.statusesNames[1].Status
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByStatus(ctx, status, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, 3, len(orders))
	prevOrderID := math.MaxInt
	for _, o := range orders {
		require.True(t, containsOrder(t, o, s.orders))
		require.GreaterOrEqual(t, prevOrderID, o.ID)
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
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByStatus(ctx, status, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, 3, len(orders))
	prevOrderCreatedAt := time.Unix(1<<63-62135596801, 999999999)
	for _, o := range orders {
		require.True(t, containsOrder(t, o, s.orders))
		require.True(t, prevOrderCreatedAt.After(o.CreatedAt) || prevOrderCreatedAt.Equal(o.CreatedAt))
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
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByStatus(ctx, status, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, 3, len(orders))
	prevOrderID := 0
	for _, o := range orders {
		require.True(t, containsOrder(t, o, s.orders))
		require.LessOrEqual(t, prevOrderID, o.ID)
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
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByStatus(ctx, status, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, 3, len(orders))
	prevOrderCreatedAt := time.Unix(0, 0)
	for _, o := range orders {
		require.True(t, containsOrder(t, o, s.orders))
		require.True(t, prevOrderCreatedAt.Before(o.CreatedAt) || prevOrderCreatedAt.Equal(o.CreatedAt))
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
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByStatus(ctx, status, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.GreaterOrEqual(t, limit, len(orders))
	require.Greater(t, len(s.orders), len(orders))
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByStatus_Offset() {
	t := s.T()

	limit := 10
	offset := 2
	orderBy := utils.AscOrder
	orderColumn := order.FieldCreatedAt
	status := s.statusesNames[1].Status
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByStatus(ctx, status, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.GreaterOrEqual(t, len(s.orders)-offset, len(orders))
	require.Greater(t, len(s.orders), len(orders))
}

func (s *orderFilterTestSuite) TestOrderRepositoryWithFilter_OrdersByPeriodAndStatusTotal() {
	t := s.T()

	status := s.statusesNames[0].Status
	from := time.Now().Add(-2*time.Hour - time.Minute)
	to := time.Now().Add(3 * time.Hour)
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	totalOrders, err := s.repository.OrdersByPeriodAndStatusTotal(ctx, from, to, status)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 1, totalOrders)

	status = s.statusesNames[1].Status
	from = time.Now().Add(-6 * time.Hour)
	to = time.Now().Add(6 * time.Hour)
	totalOrders, err = s.repository.OrdersByPeriodAndStatusTotal(ctx, from, to, status)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, 3, totalOrders)
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
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByPeriodAndStatus(ctx, from, to, status, limit, offset, orderBy, orderColumn)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Nil(t, orders)
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
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByPeriodAndStatus(ctx, from, to, status, limit, offset, orderBy, orderColumn)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Nil(t, orders)
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
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByPeriodAndStatus(ctx, from, to, status, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, 2, len(orders))
	prevOrderID := math.MaxInt
	for _, o := range orders {
		require.True(t, containsOrder(t, o, s.orders))
		require.GreaterOrEqual(t, prevOrderID, o.ID)
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
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByPeriodAndStatus(ctx, from, to, status, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, 2, len(orders))
	prevOrderCreatedAt := time.Unix(1<<63-62135596801, 999999999)
	for _, o := range orders {
		require.True(t, containsOrder(t, o, s.orders))
		require.True(t, prevOrderCreatedAt.After(o.CreatedAt) || prevOrderCreatedAt.Equal(o.CreatedAt))
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
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByPeriodAndStatus(ctx, from, to, status, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, 2, len(orders))
	prevOrderID := 0
	for _, o := range orders {
		require.True(t, containsOrder(t, o, s.orders))
		require.LessOrEqual(t, prevOrderID, o.ID)
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
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByPeriodAndStatus(ctx, from, to, status, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, 2, len(orders))
	prevOrderCreatedAt := time.Unix(0, 0)
	for _, o := range orders {
		require.True(t, containsOrder(t, o, s.orders))
		require.True(t, prevOrderCreatedAt.Before(o.CreatedAt) || prevOrderCreatedAt.Equal(o.CreatedAt))
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
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByPeriodAndStatus(ctx, from, to, status, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, limit, len(orders))
	require.Greater(t, len(s.orders), len(orders))
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
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	orders, err := s.repository.OrdersByPeriodAndStatus(ctx, from, to, status, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.GreaterOrEqual(t, len(s.orders)-offset, len(orders))
	require.Greater(t, len(s.orders), len(orders))
}
