package overdue

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-openapi/strfmt"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/handlers"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

var ErrNoChangesToCommit = fmt.Errorf("no changes to commit")

type overdueCheckup struct {
	orderStatusRepo domain.OrderStatusRepository
	orderFilterRepo domain.OrderRepositoryWithFilter
	eqStatusRepo    domain.EquipmentStatusRepository
	lg              *zap.Logger
}

func NewOverdueCheckup(
	orderStatusRepo domain.OrderStatusRepository,
	orderFilterRepo domain.OrderRepositoryWithFilter,
	eqStatusRepo domain.EquipmentStatusRepository,
	lg *zap.Logger,
) domain.OrderOverdueCheckup {
	return &overdueCheckup{
		orderStatusRepo: orderStatusRepo,
		orderFilterRepo: orderFilterRepo,
		eqStatusRepo:    eqStatusRepo,
		lg:              lg,
	}
}
func (ov *overdueCheckup) PeriodicalCheckup(ctx context.Context, overdueTimeCheckDuration time.Duration, cln *ent.Client, logger *zap.Logger) {
	ov.Checkup(ctx, cln, logger)
	tm := time.NewTimer(overdueTimeCheckDuration)
	for {
		select {
		case <-ctx.Done():
			ov.lg.Info("overdue checkup signal stop")
			tm.Stop()
			return

		case <-tm.C:
			ov.Checkup(ctx, cln, logger)
		}
	}
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

	list, err := ov.orderFilterRepo.OrdersByStatus(ctx, domain.OrderStatusInProgress, math.MaxInt64, 0, utils.AscOrder, order.FieldID)
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
	orderStatusRepo domain.OrderStatusRepository, equipmentStatusRepo domain.EquipmentStatusRepository,
) error {
	nextDay := strfmt.DateTime(rentEnd.Add(24 * time.Hour))
	model := models.NewOrderStatus{
		CreatedAt: &nextDay,
		OrderID:   &orderID,
		Comment:   &domain.OrderStatusOverdue,
		Status:    &domain.OrderStatusOverdue,
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

func equipmentCheck(ctx context.Context, orderID int, logger *zap.Logger, equipmentStatusRepo domain.EquipmentStatusRepository) ([]*ent.EquipmentStatus, error) {
	orderEquipmentStatuses, err := equipmentStatusRepo.GetEquipmentsStatusesByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	err = handlers.CheckEqStatuses(logger, orderEquipmentStatuses, domain.EquipmentStatusInUse)
	if err != nil {
		return nil, err
	}
	return orderEquipmentStatuses, nil
}
