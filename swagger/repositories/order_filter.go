package repositories

import (
	"context"
	"errors"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/orderstatus"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/statusname"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
)

type OrderRepositoryWithFilter interface {
	OrdersByStatus(ctx context.Context, status string, limit, offset int,
		orderBy, orderColumn string) ([]*ent.Order, error)
	OrdersByStatusTotal(ctx context.Context, status string) (int, error)
	OrdersByPeriodAndStatus(ctx context.Context, from, to time.Time, status string, limit, offset int,
		orderBy, orderColumn string) ([]*ent.Order, error)
	OrdersByPeriodAndStatusTotal(ctx context.Context, from, to time.Time, status string) (int, error)
}
type orderFilterRepository struct {
	client *ent.Client
}

func NewOrderFilter(client *ent.Client) *orderFilterRepository {
	return &orderFilterRepository{client: client}
}

var fieldsToOrderOrdersByStatus = []string{
	order.FieldID,
	order.FieldCreatedAt,
	order.FieldRentStart,
	order.FieldRentEnd,
}

func (r *orderFilterRepository) OrdersByStatusTotal(ctx context.Context, status string) (int, error) {
	return r.client.OrderStatus.Query().
		QueryStatusName().Where(statusname.StatusEQ(status)).QueryOrderStatus().Count(ctx)
}

func (r *orderFilterRepository) OrdersByPeriodAndStatusTotal(ctx context.Context,
	from, to time.Time, status string) (int, error) {
	return r.client.OrderStatus.Query().
		QueryStatusName().Where(statusname.StatusEQ(status)).QueryOrderStatus().
		Where(orderstatus.CurrentDateGT(from)).
		Where(orderstatus.CurrentDateLTE(to)).
		Count(ctx)
}

func (r *orderFilterRepository) OrdersByPeriodAndStatus(ctx context.Context, from, to time.Time, status string,
	limit, offset int, orderBy, orderColumn string) ([]*ent.Order, error) {
	if !utils.IsOrderField(orderColumn, fieldsToOrderOrdersByStatus) {
		return nil, errors.New("wrong field to order by")
	}
	orderFunc, err := utils.GetOrderFunc(orderBy, orderColumn)
	if err != nil {
		return nil, err
	}

	items, err := r.client.Order.Query().
		QueryOrderStatus().
		QueryStatusName().Where(statusname.StatusEQ(status)).
		QueryOrderStatus().
		Where(orderstatus.CurrentDateGT(from)).
		Where(orderstatus.CurrentDateLTE(to)).
		QueryOrder().
		WithOrderStatus().
		Order(orderFunc).Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *orderFilterRepository) OrdersByStatus(ctx context.Context, status string,
	limit, offset int, orderBy, orderColumn string) ([]*ent.Order, error) {
	if !utils.IsOrderField(orderColumn, fieldsToOrderOrdersByStatus) {
		return nil, errors.New("wrong field to order by")
	}
	orderFunc, err := utils.GetOrderFunc(orderBy, orderColumn)
	if err != nil {
		return nil, err
	}

	items, err := r.client.Order.Query().
		QueryOrderStatus().
		QueryStatusName().Where(statusname.StatusEQ(status)).
		QueryOrderStatus().QueryOrder().
		WithOrderStatus().
		Order(orderFunc).Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, err
	}
	return items, nil
}
