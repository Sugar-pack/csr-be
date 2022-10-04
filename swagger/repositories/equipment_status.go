package repositories

import (
	"context"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/equipmentstatus"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/equipmentstatusname"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

var (
	EquipmentStatusAvailable    = "available"
	EquipmentStatusBooked       = "booked"
	EquipmentStatusInUse        = "in use"
	EquipmentStatusNotAvailable = "not available"
)

type EquipmentStatusRepository interface {
	Create(ctx context.Context, data *models.NewEquipmentStatus) (*ent.EquipmentStatus, error)
	GetEquipmentsStatusesByOrder(ctx context.Context, orderID int) ([]*ent.EquipmentStatus, error)
	IsAvailableByPeriod(ctx context.Context, eqID int, startDate, endDate time.Time) (bool, error)
	Update(ctx context.Context, data *models.EquipmentStatus) (*ent.EquipmentStatus, error)
}

type equipmentStatusRepository struct {
}

func NewEquipmentStatusRepository() EquipmentStatusRepository {
	return &equipmentStatusRepository{}
}

func (r *equipmentStatusRepository) Create(ctx context.Context, data *models.NewEquipmentStatus) (*ent.EquipmentStatus, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	eq, err := tx.Equipment.Query().Where(equipment.IDEQ(int(*data.EquipmentID))).
		WithCategory().Only(ctx)
	if err != nil {
		return nil, err
	}

	category, err := eq.QueryCategory().First(ctx)
	if err != nil {
		return nil, err
	}

	startDate, endDate, err := getDates(data.StartDate, data.EndDate, int(category.MaxReservationTime))
	if err != nil {
		return nil, err
	}

	statusName, err := tx.EquipmentStatusName.Query().
		Where(equipmentstatusname.NameEQ(*data.StatusName)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	var equipmentOrder *ent.Order
	if data.OrderID != 0 {
		equipmentOrder, err = tx.Order.Query().Where(order.IDEQ(int(data.OrderID))).Only(ctx)
		if err != nil {
			return nil, err
		}
	}

	return tx.EquipmentStatus.Create().
		SetCreatedAt(time.Now()).
		SetComment(data.Comment).
		SetEndDate(*endDate).
		SetStartDate(*startDate).
		SetEquipments(eq).
		SetEquipmentStatusName(statusName).
		SetOrder(equipmentOrder).
		SetUpdatedAt(time.Now()).
		Save(ctx)
}

func (r *equipmentStatusRepository) GetEquipmentsStatusesByOrder(ctx context.Context, orderID int) ([]*ent.EquipmentStatus, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.EquipmentStatus.Query().
		QueryOrder().Where(order.IDEQ(orderID)).QueryEquipmentStatus().
		Where(equipmentstatus.ClosedAtIsNil()).
		WithEquipmentStatusName().
		All(ctx)
}

func (r *equipmentStatusRepository) IsAvailableByPeriod(ctx context.Context, eqID int,
	startDate, endDate time.Time) (bool, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return false, err
	}
	statuses, err := tx.EquipmentStatus.Query().
		QueryEquipments().Where(equipment.IDEQ(eqID)).QueryEquipmentStatus().
		Where(equipmentstatus.And(
			equipmentstatus.StartDateLTE(endDate.Add(time.Hour*24))),
			equipmentstatus.EndDateGTE(startDate),
			equipmentstatus.ClosedAtIsNil()).
		WithEquipmentStatusName().
		All(ctx)
	if err != nil {
		return false, err
	}
	for _, s := range statuses {
		if s.Edges.EquipmentStatusName.Name != EquipmentStatusAvailable {
			return false, nil
		}
	}
	return true, nil
}

func (r *equipmentStatusRepository) Update(ctx context.Context, data *models.EquipmentStatus) (*ent.EquipmentStatus, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	equipmentStatus := tx.EquipmentStatus.UpdateOneID(int(*data.ID))
	if data.StatusName != nil {
		eqStatusName, err := tx.EquipmentStatusName.Query().
			Where(equipmentstatusname.NameEQ(*data.StatusName)).
			Only(ctx)
		if err != nil {
			return nil, err
		}
		equipmentStatus.SetEquipmentStatusName(eqStatusName)
	}
	closedAt := time.Time(data.ClosedAt)
	if !closedAt.IsZero() {
		equipmentStatus.SetClosedAt(closedAt)
	}
	equipmentStatus.SetUpdatedAt(time.Now())
	return equipmentStatus.Save(ctx)
}
