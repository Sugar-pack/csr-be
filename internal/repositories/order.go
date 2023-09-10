package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-openapi/strfmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/orderstatus"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/orderstatusname"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type OrderAccessDenied struct {
	Err error
}

func (r OrderAccessDenied) Error() string {
	return r.Err.Error()
}

var fieldsToOrderOrders = []string{
	order.FieldID,
	order.FieldRentStart,
}

type orderRepository struct {
}

func NewOrderRepository() domain.OrderRepository {
	return &orderRepository{}
}

func getDates(start *strfmt.DateTime, end *strfmt.DateTime, maxSeconds int) (*time.Time, *time.Time, error) {
	rentStart := time.Time(*start)
	rentEnd := time.Time(*end)

	if rentStart.After(rentEnd) {
		return nil, nil, errors.New("start date should be before end date")
	}

	diff := rentEnd.Sub(rentStart)
	days := diff.Hours() / 24
	if days < 1 {
		return nil, nil, errors.New("small rent period")
	}

	if int(diff.Seconds()) > maxSeconds {
		return nil, nil, errors.New("too big reservation period")
	}

	return &rentStart, &rentEnd, nil
}

func getQuantity(quantity int, maxQuantity int) (*int, error) {
	if quantity > maxQuantity {
		return nil, fmt.Errorf("quantity limit exceeded: %d allowed", maxQuantity)
	}

	return &quantity, nil
}

func (r *orderRepository) List(ctx context.Context, ownerId int, filter domain.OrderFilter) ([]*ent.Order, error) {
	if !utils.IsValueInList(filter.OrderColumn, fieldsToOrderOrders) {
		return nil, errors.New("wrong column to order by")
	}
	orderFunc, err := utils.GetOrderFunc(filter.OrderBy, filter.OrderColumn)
	if err != nil {
		return nil, err
	}
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	query := tx.Order.Query().
		Where(order.HasUsersWith(user.ID(ownerId))).
		Order(orderFunc).Limit(filter.Limit).Offset(filter.Offset)

	query = r.applyListFilters(query, filter)

	items, err := query.WithUsers().WithOrderStatus().WithEquipments().All(ctx)
	if err != nil {
		return nil, err
	}

	return items, err
}

func (r *orderRepository) OrdersTotal(ctx context.Context, ownerId int) (int, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return tx.Order.Query().Where(order.HasUsersWith(user.ID(ownerId))).Count(ctx)
}

func (r *orderRepository) Create(ctx context.Context, data *models.OrderCreateRequest, ownerId int, equipmentIDs []int) (*ent.Order, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if len(equipmentIDs) == 0 {
		return nil, errors.New("no equipments for order")
	}
	equipments := make([]*ent.Equipment, len(equipmentIDs))
	for i, eqID := range equipmentIDs {
		eq, err := tx.Equipment.Query().Where(equipment.ID(eqID)).
			WithCategory().WithCurrentStatus().WithPetKinds().WithPetSize().WithPhoto().Only(ctx)
		if err != nil {
			return nil, err
		}
		equipments[i] = eq
	}

	category, err := equipments[0].QueryCategory().First(ctx)
	if err != nil {
		return nil, err
	}

	rentStart, rentEnd, err := getDates(data.RentStart, data.RentEnd, int(category.MaxReservationTime))
	if err != nil {
		return nil, err
	}

	owner, err := tx.User.Get(ctx, ownerId)
	if err != nil {
		return nil, err
	}

	ordersWithApprovedStatus, err := tx.Order.Query().
		Where(order.HasOrderStatusWith(orderstatus.
			HasOrderStatusNameWith(orderstatusname.StatusEQ(domain.OrderStatusApproved)))).
		Where(order.HasUsersWith(user.ID(ownerId))).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	var isFirst bool

	if ordersWithApprovedStatus == 0 {
		isFirst = true
	}

	statusName, err := tx.OrderStatusName.Query().Where(orderstatusname.StatusEQ(domain.OrderStatusInReview)).Only(ctx)
	if err != nil {
		return nil, err
	}

	createdOrder, err := tx.Order.
		Create().
		SetDescription(data.Description).
		SetQuantity(1).
		SetRentStart(*rentStart).
		SetRentEnd(*rentEnd).
		SetUsers(owner).
		SetUsersID(owner.ID).
		SetIsFirst(isFirst).
		SetCurrentStatus(statusName).
		AddEquipments(equipments...).
		AddEquipmentIDs(equipmentIDs...).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = tx.OrderStatus.Create().
		SetComment("Order created").
		SetCurrentDate(time.Now()).
		SetOrder(createdOrder).
		SetOrderStatusName(statusName).
		SetUsers(owner).
		SetUsersID(owner.ID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	newOrder, err := tx.Order.Query().Where(order.IDEQ(createdOrder.ID)). // get order with relations
										WithUsers().WithOrderStatus().Only(ctx)
	if err != nil {
		return nil, err
	}
	newOrder, err = r.getFullOrder(ctx, newOrder)
	if err != nil {
		return nil, err
	}
	return newOrder, nil
}

func (r *orderRepository) Update(ctx context.Context, id int, data *models.OrderUpdateRequest, userId int) (*ent.Order, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	foundOrder, err := tx.Order.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	owner, err := foundOrder.QueryUsers().First(ctx)
	if err != nil {
		return nil, err
	}

	if owner.ID != userId {
		return nil, OrderAccessDenied{Err: errors.New("permission denied")}
	}

	equipment, err := foundOrder.QueryEquipments().First(ctx)
	if err != nil {
		return nil, err
	}

	category, err := equipment.QueryCategory().First(ctx)
	if err != nil {
		return nil, err
	}

	rentStart, rentEnd, err := getDates(data.RentStart, data.RentEnd, int(category.MaxReservationTime))
	if err != nil {
		return nil, err
	}

	quantity, err := getQuantity(int(*data.Quantity), int(category.MaxReservationUnits))
	if err != nil {
		return nil, err
	}

	createdOrder, err := tx.Order.UpdateOne(foundOrder).
		SetRentStart(*rentStart).
		SetRentEnd(*rentEnd).
		SetDescription(*data.Description).
		SetQuantity(*quantity).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	returnOrder, err := tx.Order.Query().Where(order.IDEQ(createdOrder.ID)). // get order with relations
											WithUsers().WithOrderStatus().Only(ctx)
	if err != nil {
		return nil, err
	}

	return r.getFullOrder(ctx, returnOrder)
}

func (r *orderRepository) applyListFilters(q *ent.OrderQuery, filter domain.OrderFilter) *ent.OrderQuery {
	if filter.Status != nil && *filter.Status != domain.OrderStatusAll {
		statuses, isAggregated := domain.OrderStatusAggregation[*filter.Status]
		if !isAggregated {
			statuses = []string{*filter.Status}
		}
		statusValues := make([]driver.Value, len(statuses))
		for i, s := range statuses {
			statusValues[i] = s
		}
		q = q.Where(order.HasCurrentStatusWith(func(s *sql.Selector) {
			s.Where(sql.InValues(s.C(orderstatusname.FieldStatus), statusValues...))
		}))
	}
	return q
}
func (r *orderRepository) getFullOrder(ctx context.Context, order *ent.Order) (*ent.Order, error) {
	for i, orderStatus := range order.Edges.OrderStatus { // get order status relations
		statusName, err := orderStatus.QueryOrderStatusName().Only(ctx)
		if err != nil {
			return nil, err
		}
		order.Edges.OrderStatus[i].Edges.OrderStatusName = statusName
		statusUser, err := orderStatus.QueryUsers().Only(ctx)
		if err != nil {
			return nil, err
		}
		order.Edges.OrderStatus[i].Edges.Users = statusUser
	}
	eq, err := order.QueryEquipments().
		WithCategory().WithSubcategory().WithCurrentStatus().
		WithPhoto().WithPetSize().WithPetKinds().
		All(ctx)
	if err != nil {
		return nil, err
	}
	order.Edges.Equipments = eq

	return order, nil
}
