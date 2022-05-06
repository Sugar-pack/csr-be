package repositories

import (
	"context"
	"fmt"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/orderstatus"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/statusname"
)

type OrderRepositoryWithFilter interface {
	OrdersByStatus(ctx context.Context, status string) ([]ent.Order, error)
	OrdersByPeriodAndStatus(ctx context.Context, from, to time.Time, status string) ([]ent.Order, error)
}
type orderFilterRepository struct {
	client *ent.Client
}

func NewOrderFilter(client *ent.Client) *orderFilterRepository {
	return &orderFilterRepository{client: client}
}

func (r *orderFilterRepository) OrdersByPeriodAndStatus(ctx context.Context, from, to time.Time, status string) ([]ent.Order, error) {
	statusID, err := r.client.StatusName.Query().Where(statusname.StatusEQ(status)).OnlyID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get status id: %w", err)
	}

	orderStatusByStatus, err := r.client.OrderStatus.Query().
		Where(orderstatus.CurrentDateGT(from)).
		Where(orderstatus.CurrentDateLTE(to)).
		QueryStatusName().Where(statusname.IDEQ(statusID)).QueryOrderStatus().All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get order status by status: %w", err)
	}
	if len(orderStatusByStatus) == 0 {
		return nil, fmt.Errorf("no orders with status %s", status)
	}
	var orders []ent.Order
	for _, orderStatus := range orderStatusByStatus {
		order, errOrder := r.client.Order.Query().WithOrderStatus(func(query *ent.OrderStatusQuery) {
			query.Where(orderstatus.IDEQ(orderStatus.ID))
		}).Only(ctx)
		if errOrder != nil {
			return nil, errOrder
		}
		if order != nil {
			orders = append(orders, *order)
		}
	}
	return orders, nil

}

func (r *orderFilterRepository) OrdersByStatus(ctx context.Context, status string) ([]ent.Order, error) {
	statusID, err := r.client.StatusName.Query().Where(statusname.StatusEQ(status)).OnlyID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get status id: %w", err)
	}

	orderStatusByStatus, err := r.client.OrderStatus.Query().
		QueryStatusName().Where(statusname.IDEQ(statusID)).QueryOrderStatus().All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get order status by status: %w", err)
	}
	if len(orderStatusByStatus) == 0 {
		return nil, fmt.Errorf("no orders with status %s", status)
	}
	var orders []ent.Order
	for _, orderStatus := range orderStatusByStatus {
		order, errOrder := r.client.Order.Query().WithOrderStatus(func(query *ent.OrderStatusQuery) {
			query.Where(orderstatus.IDEQ(orderStatus.ID))
		}).Only(ctx)
		if errOrder != nil {
			return nil, errOrder
		}
		if order != nil {
			orders = append(orders, *order)
		}
	}
	return orders, nil

}
