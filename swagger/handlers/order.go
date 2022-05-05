package handlers

import (
	"errors"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

type Order struct {
	client *ent.Client
	logger *zap.Logger
}

func NewOrder(client *ent.Client, logger *zap.Logger) *Order {
	return &Order{
		client: client,
		logger: logger,
	}
}

func mapOrder(o *ent.Order) (*models.Order, error) {
	id := int64(o.ID)
	quantity := int64(o.Quantity)
	rentEnd := strfmt.DateTime(o.RentEnd)
	rentStart := strfmt.DateTime(o.RentStart)
	if o == nil {
		return nil, errors.New("order is nil")
	}
	owners := o.Edges.Users
	if owners == nil {
		return nil, errors.New("this order has no owners")
	}
	owner := owners[0]

	equipments := o.Edges.Equipments
	if equipments == nil {
		return nil, errors.New("this order has no equipments")
	}
	equipment := equipments[0]
	ownerId := int64(owner.ID)
	ownerName := owner.Login
	var kindId int64
	if equipment.Edges.Kind != nil {
		kindId = int64(equipment.Edges.Kind.ID)
	}
	var statusId int64
	if equipment.Edges.Status != nil {
		statusId = int64(equipment.Edges.Status.ID)
	}

	allStatuses := o.Edges.OrderStatus
	var statusToOrder *models.OrderStatus
	if len(allStatuses) != 0 {
		lastStatus := allStatuses[0]
		for _, s := range allStatuses {
			if s.CurrentDate.After(lastStatus.CurrentDate) {
				lastStatus = s
			}
		}
		mappedStatus, err := MapStatus(lastStatus)
		if err != nil {
			return nil, err
		}
		statusToOrder = mappedStatus
	}

	return &models.Order{
		Description: &o.Description,
		Equipment: &models.Equipment{
			Description: &equipment.Description,
			Kind:        &kindId,
			Location:    nil,
			Name:        &equipment.Name,
			Photo:       nil,
			RateDay:     &equipment.RateDay,
			RateHour:    &equipment.RateHour,
			Sku:         &equipment.Sku,
			Status:      &statusId,
		},
		ID:        &id,
		Quantity:  &quantity,
		RentEnd:   &rentEnd,
		RentStart: &rentStart,
		User: &models.UserEmbeddable{
			ID:   &ownerId,
			Name: &ownerName,
		},
		LastStatus: statusToOrder,
	}, nil
}

func mapOrdersToResponse(entOrders []*ent.Order) ([]*models.Order, error) {
	modelOrders := make([]*models.Order, len(entOrders))
	for i, o := range entOrders {
		order, err := mapOrder(o)
		if err != nil {
			return nil, err
		}
		modelOrders[i] = order
	}

	return modelOrders, nil
}

func (o Order) ListOrderFunc(repository repositories.OrderRepository) orders.GetAllOrdersHandlerFunc {
	return func(p orders.GetAllOrdersParams, access interface{}) middleware.Responder {
		ownerId, err := authentication.GetUserId(access)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, repositories.OrderAccessDenied{}) {
				status = http.StatusForbidden
			}
			return orders.NewGetAllOrdersDefault(status).WithPayload(buildErrorPayload(err))
		}
		ctx := p.HTTPRequest.Context()
		items, total, err := repository.List(ctx, ownerId)
		if err != nil {
			return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		mappedOrders, err := mapOrdersToResponse(items)
		if err != nil {
			return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		return orders.NewGetAllOrdersOK().WithPayload(&orders.GetAllOrdersOKBody{Data: &orders.GetAllOrdersOKBodyData{
			Items: mappedOrders,
			Total: int64(total),
		}})
	}
}

func (o Order) CreateOrderFunc(repository repositories.OrderRepository) orders.CreateOrderHandlerFunc {
	return func(p orders.CreateOrderParams, access interface{}) middleware.Responder {
		ownerId, err := authentication.GetUserId(access)
		if err != nil {
			return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}
		ctx := p.HTTPRequest.Context()
		order, err := repository.Create(ctx, p.Data, ownerId)
		if err != nil {
			return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		mappedOrder, err := mapOrder(order)
		if err != nil {
			return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		return orders.NewCreateOrderCreated().WithPayload(mappedOrder)
	}
}

func (o Order) UpdateOrderFunc(repository repositories.OrderRepository) orders.UpdateOrderHandlerFunc {
	return func(p orders.UpdateOrderParams, access interface{}) middleware.Responder {
		ownerId, err := authentication.GetUserId(access)
		if err != nil {
			return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}
		ctx := p.HTTPRequest.Context()
		id := int(p.OrderID)
		order, err := repository.Update(ctx, id, p.Data, ownerId)
		if err != nil {
			return orders.NewUpdateOrderDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		mappedOrder, err := mapOrder(order)
		if err != nil {
			return orders.NewUpdateOrderDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		return orders.NewUpdateOrderOK().WithPayload(mappedOrder)
	}
}
