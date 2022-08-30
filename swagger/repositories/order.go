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
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
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
	Create(ctx context.Context, data *models.OrderCreateRequest, ownerId int) (*ent.Order, error)
	Update(ctx context.Context, id int, data *models.OrderUpdateRequest, ownerId int) (*ent.Order, error)
}

var fieldsToOrderOrders = []string{
	order.FieldID,
	order.FieldRentStart,
}

type orderRepository struct {
	client *ent.Client
}

func NewOrderRepository(client *ent.Client) OrderRepository {
	return &orderRepository{client: client}
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
		return nil, errors.New(fmt.Sprintf("at most %d allowed", maxQuantity))
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
	items, err := r.client.Order.Query().
		Where(order.HasUsersWith(user.ID(ownerId))).
		Order(orderFunc).
		Limit(limit).Offset(offset).
		WithUsers().WithOrderStatus().WithEquipments().
		All(ctx)
	if err != nil {
		return nil, err
	}
	return items, err
}

func (r *orderRepository) OrdersTotal(ctx context.Context, ownerId int) (int, error) {
	return r.client.Order.Query().Where(order.HasUsersWith(user.ID(ownerId))).Count(ctx)
}

func (r *orderRepository) Create(ctx context.Context, data *models.OrderCreateRequest, ownerId int) (*ent.Order, error) {
	// equipment, err := r.client.Equipment.Get(ctx, int(*data.Equipment))
	equipment, err := r.client.Equipment.Query().Where(equipment.ID(int(*data.Equipment))).
		WithKind().WithStatus().WithPetKinds().WithPetSize().WithPhoto().Only(ctx)
	if err != nil {
		return nil, err
	}

	kind, err := equipment.QueryKind().First(ctx)
	if err != nil {
		return nil, err
	}

	rentStart, rentEnd, err := getDates(data.RentStart, data.RentEnd, int(kind.MaxReservationTime))
	if err != nil {
		return nil, err
	}

	owner, err := r.client.User.Get(ctx, ownerId)
	if err != nil {
		return nil, err
	}

	quantity, err := getQuantity(int(*data.Quantity), int(kind.MaxReservationUnits))
	if err != nil {
		return nil, err
	}

	createdOrder, err := r.client.Order.
		Create().
		SetDescription(*data.Description).
		SetQuantity(*quantity).
		SetRentStart(*rentStart).
		SetRentEnd(*rentEnd).
		AddUsers(owner).
		AddEquipments(equipment).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	returnOrder, err := r.client.Order.Query().Where(order.IDEQ(createdOrder.ID)).
		WithUsers().WithOrderStatus().WithEquipments().Only(ctx) // get order with relations
	if err != nil {
		return nil, err
	}

	for _, orderEquipment := range returnOrder.Edges.Equipments {
		if orderEquipment.ID == equipment.ID {
			orderEquipment.Edges = equipment.Edges
		}
	}

	for i := range returnOrder.Edges.OrderStatus { // get order status relations
		statusName, errStatusName := returnOrder.Edges.OrderStatus[i].QueryStatusName().Only(ctx)
		if errStatusName != nil {
			return nil, errStatusName
		}
		returnOrder.Edges.OrderStatus[i].Edges.StatusName = statusName
		statusUser, errStatusUser := returnOrder.Edges.OrderStatus[i].QueryUsers().Only(ctx)
		if errStatusUser != nil {
			return nil, errStatusUser
		}
		returnOrder.Edges.OrderStatus[i].Edges.Users = statusUser
	}

	return returnOrder, nil
}

func (r *orderRepository) Update(ctx context.Context, id int, data *models.OrderUpdateRequest, userId int) (*ent.Order, error) {
	foundOrder, err := r.client.Order.Get(ctx, id)
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

	kind, err := equipment.QueryKind().First(ctx)
	if err != nil {
		return nil, err
	}

	rentStart, rentEnd, err := getDates(data.RentStart, data.RentEnd, int(kind.MaxReservationTime))
	if err != nil {
		return nil, err
	}

	quantity, err := getQuantity(int(*data.Quantity), int(kind.MaxReservationUnits))
	if err != nil {
		return nil, err
	}

	createdOrder, err := r.client.Order.UpdateOne(foundOrder).
		SetRentStart(*rentStart).
		SetRentEnd(*rentEnd).
		SetDescription(*data.Description).
		SetQuantity(*quantity).
		Save(ctx)

	returnOrder, err := r.client.Order.Query().Where(order.IDEQ(createdOrder.ID)). // get order with relations
											WithUsers().WithOrderStatus().WithEquipments().Only(ctx)
	if err != nil {
		return nil, err
	}

	for i := range returnOrder.Edges.OrderStatus { // get order status relations
		statusName, errStatusName := returnOrder.Edges.OrderStatus[i].QueryStatusName().Only(ctx)
		if errStatusName != nil {
			return nil, errStatusName
		}
		returnOrder.Edges.OrderStatus[i].Edges.StatusName = statusName
		statusUser, errStatusUser := returnOrder.Edges.OrderStatus[i].QueryUsers().Only(ctx)
		if errStatusUser != nil {
			return nil, errStatusUser
		}
		returnOrder.Edges.OrderStatus[i].Edges.Users = statusUser
	}

	return returnOrder, nil
}
