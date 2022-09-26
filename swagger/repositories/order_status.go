package repositories

import (
	"context"
	"fmt"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/orderstatus"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/statusname"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type OrderStatusRepository interface {
	StatusHistory(ctx context.Context, orderId int) ([]*ent.OrderStatus, error)
	UpdateStatus(ctx context.Context, userID int, status models.NewOrderStatus) error
	GetOrderCurrentStatus(ctx context.Context, orderId int) (*ent.OrderStatus, error)
	GetUserStatusHistory(ctx context.Context, userId int) ([]*ent.OrderStatus, error)
}

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

	statusName, err := tx.StatusName.Query().Where(statusname.Status(*status.Status)).Only(ctx)
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
		SetStatusName(statusName).
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
		WithOrder().WithStatusName().WithUsers().All(ctx)

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
	order, err := tx.Order.Get(ctx, int(*status.OrderID))
	if err != nil {
		return fmt.Errorf("status history error, failed to get order: %s", err)
	}

	statusName, err := tx.StatusName.Query().Where(statusname.Status(*status.Status)).Only(ctx)
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
		SetStatusName(statusName).
		SetUsers(user).Save(ctx)

	if err != nil {
		return fmt.Errorf("status history error, failed to create order status: %s", err)
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

	status, err := order.QueryOrderStatus().Order(ent.Desc(orderstatus.FieldCurrentDate)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get statuses: %s", err)
	}
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
