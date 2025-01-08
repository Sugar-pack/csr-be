package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type orderStatusTestSuite struct {
	suite.Suite
	ctx           context.Context
	client        *ent.Client
	statusNameMap map[int]string
	repository    domain.OrderStatusRepository
	adminUser     *ent.User
	order         *ent.Order
}

func TestOrderStatusSuite(t *testing.T) {
	s := new(orderStatusTestSuite)
	suite.Run(t, s)
}

func (s *orderStatusTestSuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:orderstatus?mode=memory&cache=shared&_fk=1")
	s.client = client

	s.statusNameMap = map[int]string{ // list of statuses. copy of sql migration
		1: domain.OrderStatusInReview,
		2: domain.OrderStatusApproved,
		3: domain.OrderStatusInProgress,
		4: domain.OrderStatusRejected,
		5: domain.OrderStatusClosed,
	}

	_, err := s.client.OrderStatusName.Delete().Exec(s.ctx) // clean up
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.client.User.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.client.Order.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	for _, statusName := range s.statusNameMap { // create statuses
		_, errCreation := s.client.OrderStatusName.Create().SetStatus(statusName).Save(s.ctx)
		if errCreation != nil {
			t.Fatal(errCreation)
		}
	}

	user, err := s.client.User.Create().SetLogin("admin").SetName("user"). // create user
										SetPassword("admin").SetEmail("test@example.com").Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	s.adminUser = user

	order, err := s.client.Order.Create().SetDescription("test order").SetQuantity(1). // create order
												SetRentStart(time.Now()).SetRentEnd(time.Now()).Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	s.order = order

	s.repository = NewOrderStatusRepository()
}

func (s *orderStatusTestSuite) TearDownSuite() {
	s.client.Close()
}

func (s *orderStatusTestSuite) TestOrderStatusRepository_UpdateStatus_OrderNotExists() {
	t := s.T()
	userID := s.adminUser.ID
	comment := "test comment"
	createdAt := strfmt.DateTime(time.Now().UTC())
	orderID := int64(s.order.ID + 10)
	status, ok := s.statusNameMap[1]
	if !ok {
		t.Error("cant find status with id 1")
	}
	data := models.NewOrderStatus{
		Comment:   &comment,
		CreatedAt: &createdAt,
		OrderID:   &orderID,
		Status:    &status,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = s.repository.UpdateStatus(ctx, userID, data)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
}

func (s *orderStatusTestSuite) TestOrderStatusRepository_UpdateStatus_StatusNameNotExists() {
	t := s.T()
	userID := s.adminUser.ID
	comment := "test comment"
	createdAt := strfmt.DateTime(time.Now().UTC())
	orderID := int64(s.order.ID)
	status := "test status"
	data := models.NewOrderStatus{
		Comment:   &comment,
		CreatedAt: &createdAt,
		OrderID:   &orderID,
		Status:    &status,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = s.repository.UpdateStatus(ctx, userID, data)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
}

func (s *orderStatusTestSuite) TestOrderStatusRepository_UpdateStatus_UserNotExists() {
	t := s.T()
	userID := s.adminUser.ID + 10
	comment := "test comment"
	createdAt := strfmt.DateTime(time.Now().UTC())
	orderID := int64(s.order.ID)
	status, ok := s.statusNameMap[1]
	if !ok {
		t.Error("cant find status with id 1")
	}
	data := models.NewOrderStatus{
		Comment:   &comment,
		CreatedAt: &createdAt,
		OrderID:   &orderID,
		Status:    &status,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = s.repository.UpdateStatus(ctx, userID, data)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
}

func (s *orderStatusTestSuite) TestOrderStatusRepository_UpdateStatus_OK() {
	t := s.T()
	userID := s.adminUser.ID
	comment := "test comment"
	createdAt := strfmt.DateTime(time.Now().UTC())
	orderID := int64(s.order.ID)
	status, ok := s.statusNameMap[1]
	if !ok {
		t.Error("cant find status with id 1")
	}
	data := models.NewOrderStatus{
		Comment:   &comment,
		CreatedAt: &createdAt,
		OrderID:   &orderID,
		Status:    &status,
	}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = s.repository.UpdateStatus(ctx, userID, data)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	_, err = s.client.OrderStatus.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func (s *orderStatusTestSuite) TestOrderStatusRepository_StatusHistory_Empty() {
	t := s.T()
	orderID := s.order.ID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	statuses, err := s.repository.StatusHistory(ctx, orderID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Empty(t, statuses)
}

func (s *orderStatusTestSuite) TestOrderStatusRepository_StatusHistory() {
	t := s.T()
	orderID := s.order.ID
	// create order status

	orderStatus, err := s.client.OrderStatus.Create().SetComment("test comment").SetCurrentDate(time.Now().UTC()).
		SetOrderID(s.order.ID).SetUsersID(s.adminUser.ID).Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	statuses, err := s.repository.StatusHistory(ctx, orderID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, 1, len(statuses))
	require.Equal(t, orderStatus.ID, statuses[0].ID)
	require.Equal(t, orderStatus.Comment, statuses[0].Comment)
	require.Equal(t, orderStatus.CurrentDate, statuses[0].CurrentDate)
	_, err = s.client.OrderStatus.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func (s *orderStatusTestSuite) TestOrderStatusRepository_GetOrderCurrentStatus() {
	t := s.T()
	orderID := s.order.ID
	// create order status
	orderStatus, err := s.client.OrderStatus.Create().SetComment("test comment").SetCurrentDate(time.Now().UTC()).
		SetOrderID(s.order.ID).SetUsersID(s.adminUser.ID).Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	status, err := s.repository.GetOrderCurrentStatus(ctx, orderID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, orderStatus.ID, status.ID)
	require.Equal(t, orderStatus.Comment, status.Comment)
	require.Equal(t, orderStatus.CurrentDate, status.CurrentDate)
	_, err = s.client.OrderStatus.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func (s *orderStatusTestSuite) TestOrderStatusRepository_GetUserStatusHistory() {
	t := s.T()
	userID := s.adminUser.ID
	// create order status
	orderStatus, err := s.client.OrderStatus.Create().SetComment("test comment").SetCurrentDate(time.Now().UTC()).
		SetOrderID(s.order.ID).SetUsersID(s.adminUser.ID).Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	statuses, err := s.repository.GetUserStatusHistory(ctx, userID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, 1, len(statuses))
	require.Equal(t, orderStatus.ID, statuses[0].ID)
	require.Equal(t, orderStatus.Comment, statuses[0].Comment)
	require.Equal(t, orderStatus.CurrentDate, statuses[0].CurrentDate)
	_, err = s.client.OrderStatus.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
}
