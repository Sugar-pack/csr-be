package overdue

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-openapi/strfmt"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/handlers"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

var ErrNoChangesToCommit = fmt.Errorf("no changes to commit")

type overdueCheckup struct {
	orderStatusRepo repositories.OrderStatusRepository
	orderFilterRepo repositories.OrderRepositoryWithFilter
	eqStatusRepo    repositories.EquipmentStatusRepository
}

type OverdueCheckup interface {
	Checkup(ctx context.Context, cln *ent.Client, logger *zap.Logger)
}

func NewOverdueCheckup(orderStatusRepo repositories.OrderStatusRepository, orderFilterRepo repositories.OrderRepositoryWithFilter, eqStatusRepo repositories.EquipmentStatusRepository) OverdueCheckup {
	return &overdueCheckup{orderStatusRepo: orderStatusRepo, orderFilterRepo: orderFilterRepo, eqStatusRepo: eqStatusRepo}
}

// Checkup checks orders with status "in progress",
// equipments should be with status "in use".
// And if the day following the day of the end of the order has come,
// then such an order should move to the "Overdue" status.
func (ov *overdueCheckup) Checkup(ctx context.Context, cln *ent.Client, logger *zap.Logger) {
	tx, err := cln.Tx(ctx)
	if err != nil {
		logger.Fatal("Error while getting transaction")
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)

	list, err := ov.orderFilterRepo.OrdersByStatus(ctx, repositories.OrderStatusInProgress, math.MaxInt64, 0, utils.AscOrder, order.FieldID)
	if err != nil {
		logger.Error("Error while ordering by status", zap.Error(err))
		return
	}
	if len(list) == 0 {
		logger.Info("Order list with in progress status is empty")
		err = ErrNoChangesToCommit
		return
	}

	var (
		errArr   []error
		orderIDs []int
	)

	for _, order := range list {
		if time.Now().After(order.RentEnd) {
			err = updateStatusToOverdue(ctx, logger, order.RentEnd, int64(order.ID), order.Edges.Users.ID, ov.orderStatusRepo, ov.eqStatusRepo)
			if err != nil {
				logger.Error("Error while updating status to overdue", zap.Error(err))
				errArr = append(errArr, err)
			}
			orderIDs = append(orderIDs, order.ID)
		}
	}
	if errArr == nil && len(orderIDs) != 0 {
		err = tx.Commit()
		logger.Info("Updated Statuses to Overdue", zap.Ints("order id", orderIDs))
	} else if len(orderIDs) == 0 {
		err = ErrNoChangesToCommit
	}
}

func updateStatusToOverdue(ctx context.Context, logger *zap.Logger, rentEnd time.Time, orderID int64, userID int,
	orderStatusRepo repositories.OrderStatusRepository, equipmentStatusRepo repositories.EquipmentStatusRepository,
) error {
	nextDay := strfmt.DateTime(rentEnd.Add(24 * time.Hour))
	model := models.NewOrderStatus{
		CreatedAt: &nextDay,
		OrderID:   &orderID,
		Status:    &repositories.OrderStatusOverdue,
	}
	err := orderStatusRepo.UpdateStatus(ctx, userID, model)
	if err != nil {
		return err
	}
	orderEquipmentStatuses, err := equipmentCheck(ctx, int(orderID), logger, equipmentStatusRepo)
	if err != nil {
		return err
	}
	equipmentStatusModel := &models.EquipmentStatus{
		StartDate: &nextDay,
		// delete endDate?
	}
	if err = handlers.UpdateEqStatuses(ctx, equipmentStatusRepo, orderEquipmentStatuses, equipmentStatusModel); err != nil {
		return err
	}
	return nil
}

func equipmentCheck(ctx context.Context, orderID int, logger *zap.Logger, equipmentStatusRepo repositories.EquipmentStatusRepository) ([]*ent.EquipmentStatus, error) {
	orderEquipmentStatuses, err := equipmentStatusRepo.GetEquipmentsStatusesByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	err = handlers.CheckEqStatuses(logger, orderEquipmentStatuses, repositories.EquipmentStatusInUse)
	if err != nil {
		return nil, err
	}
	return orderEquipmentStatuses, nil
}
