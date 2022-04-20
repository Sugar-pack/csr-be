package repositories

import (
	"context"
	"fmt"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/statusname"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

type OrderStatusRepository interface {
	StatusHistory(ctx context.Context, orderId int) ([]ent.OrderStatus, error)
	UpdateStatus(ctx context.Context, userID int, status models.NewOrderStatus) error
}

type orderStatusRepository struct {
	client *ent.Client
}

func NewOrderStatusRepository(client *ent.Client) OrderStatusRepository {
	return &orderStatusRepository{client: client}
}

func (r *orderStatusRepository) StatusHistory(ctx context.Context, orderId int) ([]ent.OrderStatus, error) {
	order, err := r.client.Order.Get(ctx, orderId)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get order: %s", err)
	}

	pointersStatuses, err := order.QueryOrderStatus().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get order statuses: %s", err)
	}
	statuses := make([]ent.OrderStatus, 0, len(pointersStatuses))
	for _, element := range pointersStatuses {
		statuses = append(statuses, *element)
	}
	return statuses, nil

}

func (r *orderStatusRepository) UpdateStatus(ctx context.Context, userID int, status models.NewOrderStatus) error {
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
