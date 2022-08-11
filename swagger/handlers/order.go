package handlers

import (
	"errors"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/equipment"
	"math"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

func SetOrderHandler(client *ent.Client, logger *zap.Logger, api *operations.BeAPI) {
	orderRepo := repositories.NewOrderRepository(client)
	ordersHandler := NewOrder(logger)

	api.OrdersGetAllOrdersHandler = ordersHandler.ListOrderFunc(orderRepo)
	api.OrdersCreateOrderHandler = ordersHandler.CreateOrderFunc(orderRepo)
	api.OrdersUpdateOrderHandler = ordersHandler.UpdateOrderFunc(orderRepo)
}

type Order struct {
	logger *zap.Logger
}

func NewOrder(logger *zap.Logger) *Order {
	return &Order{
		logger: logger,
	}
}

func mapOrder(o *ent.Order, log *zap.Logger) (*models.Order, error) {
	if o == nil {
		log.Warn("order is nil")
		return nil, errors.New("order is nil")
	}
	id := int64(o.ID)
	quantity := int64(o.Quantity)
	rentEnd := strfmt.DateTime(o.RentEnd)
	rentStart := strfmt.DateTime(o.RentStart)
	owners := o.Edges.Users
	if owners == nil {
		log.Warn("order has no owners")
		return nil, errors.New("this order has no owners")
	}
	owner := owners[0]

	equipments := o.Edges.Equipments
	if equipments == nil {
		log.Warn("order has no equipments")
		return nil, errors.New("order has no equipments")
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
	var photoURL string
	if equipment.Edges.Photo != nil {
		photoURL = equipment.Edges.Photo.URL
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
			log.Error("failed to map status", zap.Error(err))
			return nil, err
		}
		statusToOrder = mappedStatus
	}

	return &models.Order{
		Description: &o.Description,
		Equipment: &models.EquipmentResponse{
			Description: &equipment.Description,
			Kind:        &kindId,
			Location:    nil,
			Name:        &equipment.Name,
			Photo:       &photoURL,
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

func mapOrdersToResponse(entOrders []*ent.Order, log *zap.Logger) ([]*models.Order, error) {
	modelOrders := make([]*models.Order, len(entOrders))
	for i, o := range entOrders {
		order, err := mapOrder(o, log)
		if err != nil {
			log.Error("failed to map order", zap.Error(err))
			return nil, err
		}
		modelOrders[i] = order
	}

	return modelOrders, nil
}

func (o Order) ListOrderFunc(repository repositories.OrderRepository) orders.GetAllOrdersHandlerFunc {
	return func(p orders.GetAllOrdersParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		limit := utils.GetParamInt(p.Limit, math.MaxInt)
		offset := utils.GetParamInt(p.Offset, 0)
		orderBy := utils.GetParamString(p.OrderBy, utils.AscOrder)
		orderColumn := utils.GetParamString(p.OrderColumn, order.FieldID)

		ownerId, err := authentication.GetUserId(access)
		if err != nil {
			o.logger.Error("get user id failed", zap.Error(err))
			status := http.StatusInternalServerError
			if errors.Is(err, repositories.OrderAccessDenied{}) {
				status = http.StatusForbidden
			}
			return orders.NewGetAllOrdersDefault(status).WithPayload(buildErrorPayload(err))
		}
		total, err := repository.OrdersTotal(ctx, ownerId)
		if err != nil {
			o.logger.Error("Error while getting total of all user's orders", zap.Error(err))
			return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}

		var items []*ent.Order
		if total > 0 {
			items, err = repository.List(ctx, ownerId, limit, offset, orderBy, orderColumn)
			if err != nil {
				o.logger.Error("list items failed", zap.Error(err))
				return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
			}
		}

		mappedOrders, err := mapOrdersToResponse(items, o.logger)
		if err != nil {
			o.logger.Error("map orders to response failed", zap.Error(err))
			return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}
		totalOrders := int64(total)
		listOrders := &models.OrderList{
			Items: mappedOrders,
			Total: &totalOrders,
		}
		return orders.NewGetAllOrdersOK().WithPayload(listOrders)
	}
}

func (o Order) CreateOrderFunc(repository repositories.OrderRepository) orders.CreateOrderHandlerFunc {
	return func(p orders.CreateOrderParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()

		ownerId, err := authentication.GetUserId(access)
		if err != nil {
			o.logger.Error("get user ID failed", zap.Error(err))
			return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		order, err := repository.Create(ctx, p.Data, ownerId)
		if err != nil {
			o.logger.Error("map orders to response failed", zap.Error(err))
			return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		mappedOrder, err := mapOrder(order, o.logger)
		if err != nil {
			o.logger.Error("failed to map order", zap.Error(err))
			return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		return orders.NewCreateOrderCreated().WithPayload(mappedOrder)
	}
}

func (o Order) UpdateOrderFunc(repository repositories.OrderRepository) orders.UpdateOrderHandlerFunc {
	return func(p orders.UpdateOrderParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()

		ownerId, err := authentication.GetUserId(access)
		if err != nil {
			o.logger.Error("get userID failed", zap.Error(err))
			return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		orderID := int(p.OrderID)
		order, err := repository.Update(ctx, orderID, p.Data, ownerId)
		if err != nil {
			o.logger.Error("update order failed", zap.Error(err))
			return orders.NewUpdateOrderDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		mappedOrder, err := mapOrder(order, o.logger)
		if err != nil {
			o.logger.Error("failed to map order", zap.Error(err))
			return orders.NewUpdateOrderDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		return orders.NewUpdateOrderOK().WithPayload(mappedOrder)
	}
}
