package repositories

import (
	"context"

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
		return nil, err
	}

	pointers_statuses, err := order.QueryOrderStatus().All(ctx)
	if err != nil {
		return nil, err
	}
	statuses := []ent.OrderStatus{}
	for _, element := range pointers_statuses {
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
		return err
	}
	return nil

}
