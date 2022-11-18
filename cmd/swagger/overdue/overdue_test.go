package overdue

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	repomock "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/mocks/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

type CheckupTestSuite struct {
	suite.Suite
	eqStatusRepo    *repomock.EquipmentStatusRepository
	orderStatusRepo *repomock.OrderStatusRepository
	orderFilterRepo *repomock.OrderRepositoryWithFilter
	setupApi        OverdueCheckup
}

func TestStatusSuite(t *testing.T) {
	suite.Run(t, new(CheckupTestSuite))
}

func (s *CheckupTestSuite) SetupTest() {
	s.eqStatusRepo = &repomock.EquipmentStatusRepository{}
	s.orderStatusRepo = &repomock.OrderStatusRepository{}
	s.orderFilterRepo = &repomock.OrderRepositoryWithFilter{}
	s.setupApi = NewOverdueCheckup(s.orderStatusRepo, s.orderFilterRepo, s.eqStatusRepo)
}

func (s *CheckupTestSuite) TestCheckup_Success_PartialUpdate() {
	t := s.T()
	ctx := context.Background()

	client := enttest.Open(t, "sqlite3", "file:overdue_success?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	observedZapCore, observedLogs := observer.New(zap.InfoLevel)
	observedLogger := zap.New(observedZapCore)

	var list []*ent.Order
	lastIDNoUpdate := 2

	for i := 0; i < 3; i++ {
		order := orderWithEdges(t, i)
		if i == lastIDNoUpdate {
			order.RentEnd = time.Now().Add(24 * time.Hour)
		}
		list = append(list, order)
	}

	s.orderFilterRepo.On("OrdersByStatus", mock.Anything, repositories.OrderStatusInProgress, math.MaxInt64, 0, utils.AscOrder, order.FieldID).Return(list, nil)

	var updated []int
	for i := range list {
		order := list[i]
		orderID := list[i].ID

		date := strfmt.DateTime(order.RentEnd.Add(24 * time.Hour))
		id := int64(orderID)
		eqID := int64(order.Edges.Equipments[0].ID)
		if i != lastIDNoUpdate {
			modelExpected := models.NewOrderStatus{
				CreatedAt: &date,
				OrderID:   &id,
				Status:    &repositories.OrderStatusOverdue,
			}
			s.orderStatusRepo.On("UpdateStatus", mock.Anything, order.Edges.Users.ID, modelExpected).Return(nil)
			s.eqStatusRepo.On("GetEquipmentsStatusesByOrder", mock.Anything, orderID).Return(order.Edges.EquipmentStatus, nil)
			eqModel := &models.EquipmentStatus{
				ID:        &eqID,
				StartDate: &date,
			}
			s.eqStatusRepo.On("Update", mock.Anything, eqModel).Return(nil, nil)
			updated = append(updated, orderID)
		}
	}

	s.setupApi.Checkup(ctx, client, observedLogger)

	require.Equal(t, 1, observedLogs.Len())
	firstLog := observedLogs.All()[0]
	assert.Equal(t, "Updated Statuses to Overdue", firstLog.Message)
	assert.Equal(t, firstLog.Context[0], zap.Ints("order id", updated))
	s.orderFilterRepo.AssertExpectations(t)
	s.orderStatusRepo.AssertExpectations(t)
	s.eqStatusRepo.AssertExpectations(t)
}

func (s *CheckupTestSuite) TestCheckup_EmptyListOfOrders() {
	t := s.T()
	ctx := context.Background()

	client := enttest.Open(t, "sqlite3", "file:overdue?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	observedZapCore, observedLogs := observer.New(zap.InfoLevel)
	observedLogger := zap.New(observedZapCore)

	s.orderFilterRepo.On("OrdersByStatus", mock.Anything, repositories.OrderStatusInProgress, math.MaxInt64, 0, utils.AscOrder, order.FieldID).Return(nil, nil)
	s.setupApi.Checkup(ctx, client, observedLogger)

	require.Equal(t, 1, observedLogs.Len())
	firstLog := observedLogs.All()[0]
	assert.Equal(t, "Order list with in progress status is empty", firstLog.Message)
	s.orderFilterRepo.AssertExpectations(t)
	s.orderStatusRepo.AssertExpectations(t)
	s.eqStatusRepo.AssertExpectations(t)
}

func (s *CheckupTestSuite) TestCheckup_OrderFilter_RepoErr() {
	t := s.T()
	ctx := context.Background()

	client := enttest.Open(t, "sqlite3", "file:overdue?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	observedZapCore, observedLogs := observer.New(zap.ErrorLevel)
	observedLogger := zap.New(observedZapCore) // zap.WithFatalHook(zapcore.WriteThenGoexit)

	err := errors.New("error")
	s.orderFilterRepo.On("OrdersByStatus", mock.Anything, repositories.OrderStatusInProgress, math.MaxInt64, 0, utils.AscOrder, order.FieldID).Return(nil, err)
	s.setupApi.Checkup(ctx, client, observedLogger)

	require.Equal(t, 1, observedLogs.Len())
	firstLog := observedLogs.All()[0]
	assert.Equal(t, "Error while ordering by status", firstLog.Message)
	s.orderFilterRepo.AssertExpectations(t)
	s.orderStatusRepo.AssertExpectations(t)
	s.eqStatusRepo.AssertExpectations(t)
}

func (s *CheckupTestSuite) TestCheckup_NoUpdate_NotOverdue() {
	t := s.T()
	ctx := context.Background()

	observedZapCore, observedLogs := observer.New(zap.ErrorLevel)
	observedLogger := zap.New(observedZapCore)

	client := enttest.Open(t, "sqlite3", "file:overdue?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	var list []*ent.Order

	for i := 0; i < 3; i++ {
		order := orderWithEdges(t, i)
		order.RentEnd = time.Now().Add(24 * time.Hour)
		list = append(list, order)
	}

	s.orderFilterRepo.On("OrdersByStatus", mock.Anything, repositories.OrderStatusInProgress, math.MaxInt64, 0, utils.AscOrder, order.FieldID).Return(list, nil)

	s.setupApi.Checkup(ctx, client, observedLogger)

	require.Equal(t, 0, observedLogs.Len())
	s.orderFilterRepo.AssertExpectations(t)
	s.orderStatusRepo.AssertExpectations(t)
	s.eqStatusRepo.AssertExpectations(t)
}

func (s *CheckupTestSuite) TestCheckup_OrderStatus_RepoErr() {
	t := s.T()
	ctx := context.Background()

	client := enttest.Open(t, "sqlite3", "file:overdue_success?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	observedZapCore, observedLogs := observer.New(zap.ErrorLevel)
	observedLogger := zap.New(observedZapCore)

	var list []*ent.Order
	lastIDWithErr := 2

	for i := 0; i < 3; i++ {
		order := orderWithEdges(t, i)
		list = append(list, order)
	}

	s.orderFilterRepo.On("OrdersByStatus", mock.Anything, repositories.OrderStatusInProgress, math.MaxInt64, 0, utils.AscOrder, order.FieldID).Return(list, nil)

	for i := range list {
		order := list[i]
		orderID := list[i].ID

		date := strfmt.DateTime(order.RentEnd.Add(24 * time.Hour))
		id := int64(orderID)
		eqID := int64(order.Edges.Equipments[0].ID)
		if i != lastIDWithErr {
			modelExpected := models.NewOrderStatus{
				CreatedAt: &date,
				OrderID:   &id,
				Status:    &repositories.OrderStatusOverdue,
			}
			s.orderStatusRepo.On("UpdateStatus", mock.Anything, order.Edges.Users.ID, modelExpected).Return(nil)
			s.eqStatusRepo.On("GetEquipmentsStatusesByOrder", mock.Anything, orderID).Return(order.Edges.EquipmentStatus, nil)
			eqModel := &models.EquipmentStatus{
				ID:        &eqID,
				StartDate: &date,
			}
			s.eqStatusRepo.On("Update", mock.Anything, eqModel).Return(nil, nil)
		} else {
			err := errors.New("error")
			s.orderStatusRepo.On("UpdateStatus", mock.Anything, order.Edges.Users.ID, mock.Anything).Return(err)
		}
	}

	s.setupApi.Checkup(ctx, client, observedLogger)

	require.Equal(t, 1, observedLogs.Len())
	firstLog := observedLogs.All()[0]
	assert.Equal(t, "Error while updating status to overdue", firstLog.Message)
	s.orderFilterRepo.AssertExpectations(t)
	s.orderStatusRepo.AssertExpectations(t)
	s.eqStatusRepo.AssertExpectations(t)
}

func (s *CheckupTestSuite) TestCheckup_EquipmentsStatuses_RepoErr() {
	t := s.T()
	ctx := context.Background()

	client := enttest.Open(t, "sqlite3", "file:overdue_success?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	observedZapCore, observedLogs := observer.New(zap.ErrorLevel)
	observedLogger := zap.New(observedZapCore)

	var list []*ent.Order
	lastIDWithErr := 2

	for i := 0; i < 3; i++ {
		order := orderWithEdges(t, i)
		list = append(list, order)
	}

	s.orderFilterRepo.On("OrdersByStatus", mock.Anything, repositories.OrderStatusInProgress, math.MaxInt64, 0, utils.AscOrder, order.FieldID).Return(list, nil)

	for i := range list {
		order := list[i]
		orderID := list[i].ID

		date := strfmt.DateTime(order.RentEnd.Add(24 * time.Hour))
		id := int64(orderID)
		eqID := int64(order.Edges.Equipments[0].ID)
		modelExpected := models.NewOrderStatus{
			CreatedAt: &date,
			OrderID:   &id,
			Status:    &repositories.OrderStatusOverdue,
		}
		s.orderStatusRepo.On("UpdateStatus", mock.Anything, order.Edges.Users.ID, modelExpected).Return(nil)

		if i != lastIDWithErr {
			s.eqStatusRepo.On("GetEquipmentsStatusesByOrder", mock.Anything, orderID).Return(order.Edges.EquipmentStatus, nil)
			eqModel := &models.EquipmentStatus{
				ID:        &eqID,
				StartDate: &date,
			}
			s.eqStatusRepo.On("Update", mock.Anything, eqModel).Return(nil, nil)
		} else {
			err := errors.New("error")
			s.eqStatusRepo.On("GetEquipmentsStatusesByOrder", mock.Anything, orderID).Return(nil, err)
		}
	}

	s.setupApi.Checkup(ctx, client, observedLogger)

	require.Equal(t, 1, observedLogs.Len())
	firstLog := observedLogs.All()[0]
	assert.Equal(t, "Error while updating status to overdue", firstLog.Message)
	s.orderFilterRepo.AssertExpectations(t)
	s.orderStatusRepo.AssertExpectations(t)
	s.eqStatusRepo.AssertExpectations(t)
}

func (s *CheckupTestSuite) TestCheckup_EquipmentsStatuses_UpdateRepoErr() {
	t := s.T()
	ctx := context.Background()

	client := enttest.Open(t, "sqlite3", "file:overdue_success?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	observedZapCore, observedLogs := observer.New(zap.ErrorLevel)
	observedLogger := zap.New(observedZapCore)

	var list []*ent.Order
	lastIDWithErr := 2

	for i := 0; i < 3; i++ {
		order := orderWithEdges(t, i)
		list = append(list, order)
	}

	s.orderFilterRepo.On("OrdersByStatus", mock.Anything, repositories.OrderStatusInProgress, math.MaxInt64, 0, utils.AscOrder, order.FieldID).Return(list, nil)

	for i := range list {
		order := list[i]
		orderID := list[i].ID

		date := strfmt.DateTime(order.RentEnd.Add(24 * time.Hour))
		id := int64(orderID)
		eqID := int64(order.Edges.Equipments[0].ID)
		modelExpected := models.NewOrderStatus{
			CreatedAt: &date,
			OrderID:   &id,
			Status:    &repositories.OrderStatusOverdue,
		}
		s.orderStatusRepo.On("UpdateStatus", mock.Anything, order.Edges.Users.ID, modelExpected).Return(nil)
		s.eqStatusRepo.On("GetEquipmentsStatusesByOrder", mock.Anything, orderID).Return(order.Edges.EquipmentStatus, nil)
		eqModel := &models.EquipmentStatus{
			ID:        &eqID,
			StartDate: &date,
		}
		if i != lastIDWithErr {
			s.eqStatusRepo.On("Update", mock.Anything, eqModel).Return(nil, nil)
		} else {
			err := errors.New("error")
			s.eqStatusRepo.On("Update", mock.Anything, eqModel).Return(nil, err)
		}
	}

	s.setupApi.Checkup(ctx, client, observedLogger)

	require.Equal(t, 1, observedLogs.Len())
	firstLog := observedLogs.All()[0]
	assert.Equal(t, "Error while updating status to overdue", firstLog.Message)
	s.orderFilterRepo.AssertExpectations(t)
	s.orderStatusRepo.AssertExpectations(t)
	s.eqStatusRepo.AssertExpectations(t)
}

func orderWithEdges(t *testing.T, id int) *ent.Order {
	t.Helper()
	equipment := &ent.Equipment{
		ID:          id,
		Description: "description",
	}
	return &ent.Order{
		ID:          id,
		Description: fmt.Sprintf("test description %d", id),
		Quantity:    id%2 + 1,
		RentStart:   time.Now().Add(time.Duration(-id*2*24) * time.Hour),
		RentEnd:     time.Now().Add(time.Duration(-id*24) * time.Hour),
		CreatedAt:   time.Now().Add(time.Duration(-id*2*24) * time.Hour),
		Edges: ent.OrderEdges{
			Users: &ent.User{
				ID:    1,
				Login: "login",
			},
			Equipments: []*ent.Equipment{equipment},
			OrderStatus: []*ent.OrderStatus{
				{
					ID: id,
					Edges: ent.OrderStatusEdges{
						OrderStatusName: &ent.OrderStatusName{
							ID:     id%2 + 1,
							Status: repositories.OrderStatusInProgress,
						},
						Users: &ent.User{
							ID: id,
						},
						Order: &ent.Order{
							ID: id,
							Edges: ent.OrderEdges{
								Equipments: []*ent.Equipment{equipment},
							},
						},
					},
				},
			},
			EquipmentStatus: []*ent.EquipmentStatus{
				{
					ID: id,
					Edges: ent.EquipmentStatusEdges{
						EquipmentStatusName: &ent.EquipmentStatusName{
							Name: repositories.EquipmentStatusInUse,
						},
					},
				},
			},
		},
	}
}
