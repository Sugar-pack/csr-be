package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/orderstatusname"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type OrderAccessDenied struct {
	Err error
}

func (r OrderAccessDenied) Error() string {
	return r.Err.Error()
}

type OrderRepository interface {
	List(ctx context.Context, ownerId, limit, offset int, orderBy, orderColumn string) ([]*ent.Order, error)
	OrdersTotal(ctx context.Context, ownerId int) (int, error)
	Create(ctx context.Context, data *models.OrderCreateRequest, ownerId int, equipmentIDs []int) (*ent.Order, error)
	Update(ctx context.Context, id int, data *models.OrderUpdateRequest, ownerId int) (*ent.Order, error)
}

var fieldsToOrderOrders = []string{
	order.FieldID,
	order.FieldRentStart,
}

var (
	OrderStatusInReview   = "in review"
	OrderStatusApproved   = "approved"
	OrderStatusInProgress = "in progress"
	OrderStatusRejected   = "rejected"
	OrderStatusClosed     = "closed"
	OrderStatusPrepared   = "prepared"
)

type orderRepository struct {
}

func NewOrderRepository() OrderRepository {
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

func (r *orderRepository) List(ctx context.Context, ownerId, limit, offset int, orderBy, orderColumn string) ([]*ent.Order, error) {
	if !utils.IsOrderField(orderColumn, fieldsToOrderOrders) {
		return nil, errors.New("wrong column to order by")
	}
	orderFunc, err := utils.GetOrderFunc(orderBy, orderColumn)
	if err != nil {
		return nil, err
	}
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	items, err := tx.Order.Query().
		Where(order.HasUsersWith(user.ID(ownerId))).
		Order(orderFunc).
		Limit(limit).Offset(offset).
		WithUsers().WithOrderStatus().
		All(ctx)
	if err != nil {
		return nil, err
	}
	for i, item := range items { // get order status relations
		items[i], err = r.getFullOrder(ctx, item)
		if err != nil {
			return nil, err
		}
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

	quantity, err := getQuantity(int(*data.Quantity), int(category.MaxReservationUnits))
	if err != nil {
		return nil, err
	}

	createdOrder, err := tx.Order.
		Create().
		SetDescription(*data.Description).
		SetQuantity(*quantity).
		SetRentStart(*rentStart).
		SetRentEnd(*rentEnd).
		SetUsers(owner).
		SetUsersID(owner.ID).
		AddEquipments(equipments...).
		AddEquipmentIDs(equipmentIDs...).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	statusName, err := tx.OrderStatusName.Query().Where(orderstatusname.StatusEQ(OrderStatusInReview)).Only(ctx)
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
