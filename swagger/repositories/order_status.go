package repositories

import (
	"context"
	"fmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
)

type OrderStatusRepository interface {
	StatusHistory(ctx context.Context, orderId int) ([]ent.OrderStatus, error)
	UpdateStatus(ctx context.Context, status ent.OrderStatus) error
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

func (r *orderStatusRepository) UpdateStatus(ctx context.Context, status ent.OrderStatus) error {
	_, err := r.client.OrderStatus.Create().
		SetComment(status.Comment).
		SetCurrentDate(status.CurrentDate).
		SetOrder(status.Edges.Order).
		SetStatusName(status.Edges.StatusName).
		SetUsers(status.Edges.Users).Save(ctx)

	if err != nil {
		return fmt.Errorf("status history error, failed to create order status: %s", err)
	}
	return nil

}
