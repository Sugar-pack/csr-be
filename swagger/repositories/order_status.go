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
)

type OrderStatusRepository interface {
	StatusHistory(ctx context.Context, orderId int) ([]*ent.OrderStatus, error)
	UpdateStatus(ctx context.Context, userID int, status models.NewOrderStatus) error
}

type orderStatusRepository struct {
	client *ent.Client
}

func NewOrderStatusRepository(client *ent.Client) *orderStatusRepository {
	return &orderStatusRepository{client: client}
}

func (r *orderStatusRepository) ApproveOrRejectOrder(ctx context.Context, userID int, status models.NewOrderStatus) error {
	order, err := r.client.Order.Get(ctx, int(*status.OrderID))
	if err != nil {
		return fmt.Errorf("status history error, failed to get order: %s", err)
	}

	statusName, err := r.client.StatusName.Query().Where(statusname.Status(*status.Status)).Only(ctx)
	if err != nil {
		return fmt.Errorf("status history error, failed to get status name: %s", err)
	}
	user, err := r.client.User.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("status history error, failed to get user: %s", err)
	}
	_, err = r.client.OrderStatus.Create().
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
	statuses, err := r.client.OrderStatus.Query().
		QueryOrder().Where(order.IDEQ(orderId)).QueryOrderStatus().
		WithOrder().WithStatusName().WithUsers().All(ctx)

	return statuses, err
}

func (r orderStatusRepository) FilterOrdersStatusesByPeriodAndStatus(ctx context.Context, from, to time.Time, status string) ([]ent.OrderStatus, error) {
	statusID, err := r.client.StatusName.Query().Where(statusname.StatusEQ(status)).OnlyID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get status id: %w", err)
	}

	orderStatusesByStatusPointers, err := r.client.OrderStatus.Query().
		Where(orderstatus.CurrentDateGT(from)).
		Where(orderstatus.CurrentDateLTE(to)).
		WithStatusName(func(query *ent.StatusNameQuery) {
			query.Where(statusname.IDEQ(statusID))
		}).All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get order status by status: %w", err)
	}
	if len(orderStatusesByStatusPointers) == 0 {
		return nil, fmt.Errorf("no orders with status %s", status)
	}
	statuses := make([]ent.OrderStatus, 0, len(orderStatusesByStatusPointers))
	for _, element := range orderStatusesByStatusPointers {
		statuses = append(statuses, *element)
	}

	return statuses, nil
}
func (r *orderStatusRepository) UpdateStatus(ctx context.Context, userID int, status models.NewOrderStatus) error {
	if status.OrderID == nil {
		return fmt.Errorf("order id is required")
	}
	order, err := r.client.Order.Get(ctx, int(*status.OrderID))
	if err != nil {
		return fmt.Errorf("status history error, failed to get order: %s", err)
	}

	statusName, err := r.client.StatusName.Query().Where(statusname.Status(*status.Status)).Only(ctx)
	if err != nil {
		return fmt.Errorf("status history error, failed to get status name: %s", err)
	}
	user, err := r.client.User.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("status history error, failed to get user: %s", err)
	}
	_, err = r.client.OrderStatus.Create().
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

func (r *orderStatusRepository) FilterOrderStatusesByName(ctx context.Context, statusName string) ([]ent.OrderStatus, error) {
	status, err := r.client.StatusName.Query().Where(statusname.Status(statusName)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get status name: %s", err)
	}

	pointersStatuses, err := status.QueryOrderStatus().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get order statuses: %s", err)
	}
	statuses := make([]ent.OrderStatus, 0, len(pointersStatuses))
	for _, element := range pointersStatuses {
		statuses = append(statuses, *element)
	}
	return statuses, nil
}

func (r *orderStatusRepository) GetOrderCurrentStatus(ctx context.Context, orderId int) (*ent.OrderStatus, error) {
	order, err := r.client.Order.Get(ctx, orderId)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get order: %s", err)
	}

	pointersStatuses, err := order.QueryOrderStatus().Order(ent.Desc("current_date")).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get statuses: %s", err)
	}
	status := pointersStatuses[0]
	return status, nil
}

func (r *orderStatusRepository) GetUserStatusHistory(ctx context.Context, userId int) ([]ent.OrderStatus, error) {
	user, err := r.client.User.Get(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get user: %s", err)
	}

	pointersStatuses, err := user.QueryOrderStatus().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get order statuses: %s", err)
	}
	statuses := make([]ent.OrderStatus, 0, len(pointersStatuses))
	for _, element := range pointersStatuses {
		statuses = append(statuses, *element)
	}
	return statuses, nil
}

func (r *orderStatusRepository) GetAllOrderStatuses(ctx context.Context, orderId int) ([]ent.OrderStatus, error) {
	pointersStatuses, err := r.client.OrderStatus.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get order statuses: %s", err)
	}
	statuses := make([]ent.OrderStatus, 0, len(pointersStatuses))
	for _, element := range pointersStatuses {
		statuses = append(statuses, *element)
	}
	return statuses, nil

}
