package repositories

import (
	"context"
	"fmt"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/orderstatus"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/orderstatusname"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type orderStatusRepository struct {
}

func NewOrderStatusRepository() *orderStatusRepository {
	return &orderStatusRepository{}
}

func (r *orderStatusRepository) ApproveOrRejectOrder(ctx context.Context, userID int, status models.NewOrderStatus) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	order, err := tx.Order.Get(ctx, int(*status.OrderID))
	if err != nil {
		return fmt.Errorf("status history error, failed to get order: %s", err)
	}

	statusName, err := tx.OrderStatusName.Query().Where(orderstatusname.Status(*status.Status)).Only(ctx)
	if err != nil {
		return fmt.Errorf("status history error, failed to get status name: %s", err)
	}
	user, err := tx.User.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("status history error, failed to get user: %s", err)
	}
	_, err = tx.OrderStatus.Create().
		SetComment(*status.Comment).
		SetCurrentDate(time.Time(*status.CreatedAt)).
		SetOrder(order).
		SetOrderStatusName(statusName).
		SetUsers(user).Save(ctx)

	if err != nil {
		return fmt.Errorf("status history error, failed to create order status: %s", err)
	}
	return nil
}

func (r *orderStatusRepository) StatusHistory(ctx context.Context, orderId int) ([]*ent.OrderStatus, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	statuses, err := tx.OrderStatus.Query().
		QueryOrder().Where(order.IDEQ(orderId)).QueryOrderStatus().
		WithOrder().WithOrderStatusName().WithUsers().All(ctx)

	return statuses, err
}

func (r *orderStatusRepository) UpdateStatus(ctx context.Context, userID int, status models.NewOrderStatus) error {
	if status.OrderID == nil {
		return fmt.Errorf("order id is required")
	}
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	receivedOrder, err := tx.Order.Get(ctx, int(*status.OrderID))
	if err != nil {
		return fmt.Errorf("status history error, failed to get order: %s", err)
	}

	statusName, err := tx.OrderStatusName.Query().
		Where(orderstatusname.Status(*status.Status)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("status history error, failed to get status name: %s", err)
	}
	receivedUser, err := tx.User.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("status history error, failed to get user: %s", err)
	}
	_, err = tx.OrderStatus.Create().
		SetComment(*status.Comment).
		SetCurrentDate(time.Time(*status.CreatedAt)).
		SetOrder(receivedOrder).
		SetOrderStatusName(statusName).
		SetUsers(receivedUser).Save(ctx)

	if err != nil {
		return fmt.Errorf("status history error, failed to create order status: %s", err)
	}

	if *status.Status == domain.OrderStatusApproved {
		_, err = tx.Order.Update().Where(order.IsFirstEQ(true)).
			Where(order.HasUsersWith(user.ID(userID))).
			SetIsFirst(false).
			Save(ctx)

		if err != nil {
			return fmt.Errorf("unable to update is_first field for orders: %s", err)
		}
	}

	return nil
}

func (r *orderStatusRepository) GetOrderCurrentStatus(ctx context.Context, orderId int) (*ent.OrderStatus, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	order, err := tx.Order.Get(ctx, orderId)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get order: %s", err)
	}

	status, err := order.QueryOrderStatus().
		WithOrderStatusName().
		WithOrder().
		Order(ent.Desc(orderstatus.FieldCurrentDate)).First(ctx)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get statuses: %s", err)
	}
	equipment, err := status.Edges.Order.QueryEquipments().All(ctx)
	if err != nil {
		return nil, err
	}
	status.Edges.Order.Edges.Equipments = equipment
	return status, nil
}

func (r *orderStatusRepository) GetUserStatusHistory(ctx context.Context, userId int) ([]*ent.OrderStatus, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, err := tx.User.Get(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get user: %s", err)
	}

	pointersStatuses, err := user.QueryOrderStatus().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get order statuses: %s", err)
	}
	return pointersStatuses, nil
}
