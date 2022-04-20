package repositories

import (
	"context"
	"errors"
	"fmt"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"github.com/go-openapi/strfmt"
	"time"
)

type OrderAccessDenied struct {
	Err error
}

func (r OrderAccessDenied) Error() string {
	return r.Err.Error()
}

type OrderRepository interface {
	List(ctx context.Context, ownerId int) ([]*ent.Order, int, error)
	Create(ctx context.Context, data *models.OrderCreateRequest, ownerId int) (*ent.Order, error)
	Update(ctx context.Context, id int, data *models.OrderUpdateRequest, ownerId int) (*ent.Order, error)
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

func (r *orderRepository) List(ctx context.Context, ownerId int) (items []*ent.Order, total int, err error) {
	items, err = r.client.Order.Query().Where(order.HasUsersWith(user.ID(ownerId))).Order(ent.Desc("id")).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	total, err = r.client.Order.Query().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return
}

func (r *orderRepository) Create(ctx context.Context, data *models.OrderCreateRequest, ownerId int) (order *ent.Order, err error) {
	equipment, err := r.client.Equipment.Get(ctx, int(*data.Equipment))
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

	order, err = r.client.Order.
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

	return
}

func (r *orderRepository) Update(ctx context.Context, id int, data *models.OrderUpdateRequest, userId int) (order *ent.Order, err error) {
	order, err = r.client.Order.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	owner, err := order.QueryUsers().First(ctx)
	if err != nil {
		return nil, err
	}

	if owner.ID != userId {
		return nil, OrderAccessDenied{Err: errors.New("permission denied")}
	}

	equipment, err := order.QueryEquipments().First(ctx)
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

	order, err = r.client.Order.UpdateOne(order).
		SetRentStart(*rentStart).
		SetRentEnd(*rentEnd).
		SetDescription(*data.Description).
		SetQuantity(*quantity).
		Save(ctx)

	return
}
