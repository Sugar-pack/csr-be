package repositories

import (
	"context"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/equipmentstatus"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/equipmentstatusname"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

const maintenanceTime = time.Hour * 24 // Equipment must be maintained and sanitized after return

type equipmentStatusRepository struct {
}

func NewEquipmentStatusRepository() domain.EquipmentStatusRepository {
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
		WithEquipmentStatusName().
		All(ctx)
}

func (r *equipmentStatusRepository) GetEquipmentStatusByID(
	ctx context.Context, equipmentStatusID int) (*ent.EquipmentStatus, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return tx.EquipmentStatus.Query().Where(equipmentstatus.ID(equipmentStatusID)).
		WithEquipmentStatusName().WithEquipments().
		Only(ctx)
}

func (r *equipmentStatusRepository) GetOrderAndUserByEquipmentStatusID(
	ctx context.Context,
	equipmentStatusID int) (*ent.Order, *ent.User, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	orderResult, err := tx.Order.Query().
		Where(order.HasEquipmentStatusWith(equipmentstatus.IDEQ(equipmentStatusID))).
		Only(ctx)
	if err != nil {
		return nil, nil, err
	}

	userResult, err := tx.User.Query().Where(user.HasOrderWith(order.IDEQ(equipmentStatusID))).
		Only(ctx)
	if err != nil {
		return nil, nil, err
	}

	return orderResult, userResult, nil
}

func (r *equipmentStatusRepository) GetUnavailableEquipmentStatusByEquipmentID(
	ctx context.Context, equipmentID int) ([]*ent.EquipmentStatus, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return tx.EquipmentStatus.Query().
		Where(equipmentstatus.HasEquipmentsWith(equipment.IDEQ(equipmentID))).
		Where(
			equipmentstatus.Or(
				equipmentstatus.HasEquipmentStatusNameWith(equipmentstatusname.NameEQ(domain.EquipmentStatusBooked)),
				equipmentstatus.HasEquipmentStatusNameWith(equipmentstatusname.NameEQ(domain.EquipmentStatusInUse)),
				equipmentstatus.HasEquipmentStatusNameWith(equipmentstatusname.NameEQ(domain.EquipmentStatusNotAvailable)),
			)).
		Where(equipmentstatus.EndDateGTE(time.Now())).
		All(ctx)
}

func (r *equipmentStatusRepository) HasStatusByPeriod(ctx context.Context, status string, eqID int,
	startDate, endDate time.Time) (bool, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return false, err
	}
	statuses, err := tx.EquipmentStatus.Query().
		QueryEquipments().
		Where(equipment.IDEQ(eqID)).
		QueryEquipmentStatus().
		Where(equipmentstatus.And(
			equipmentstatus.StartDateLTE(endDate.Add(maintenanceTime))),
			equipmentstatus.EndDateGTE(startDate)).
		WithEquipmentStatusName().
		All(ctx)
	if err != nil {
		return false, err
	}
	for _, s := range statuses {
		if s.Edges.EquipmentStatusName.Name != status {
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
	if data.StartDate != nil {
		equipmentStatus.SetStartDate(time.Time(*data.StartDate))
	}
	if data.EndDate != nil {
		equipmentStatus.SetEndDate(time.Time(*data.EndDate))
	}
	equipmentStatus.SetUpdatedAt(time.Now())
	return equipmentStatus.Save(ctx)
}

// GetLastEquipmentStatusByEquipmentID - returns last equipment status for equpment by column EndDate.
func (r *equipmentStatusRepository) GetLastEquipmentStatusByEquipmentID(ctx context.Context, equipmentID int) (*ent.EquipmentStatus, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.EquipmentStatus.
		Query().
		QueryEquipments().
		Where(equipment.IDEQ(equipmentID)).
		QueryEquipmentStatus().
		WithEquipmentStatusName().
		Order(ent.Asc(equipmentstatus.FieldEndDate)).
		First(ctx)
}
