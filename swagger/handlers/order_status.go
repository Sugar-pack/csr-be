package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

type OrderStatus struct {
	client *ent.Client
	logger *zap.Logger
}

func NewOrderStatus(client *ent.Client, logger *zap.Logger) *OrderStatus {
	return &OrderStatus{
		client: client,
		logger: logger,
	}
}

func (h OrderStatus) OrderStatusesHistory(repository repositories.OrderStatusRepository) orders.GetFullOrderHistoryHandlerFunc {
	return func(p orders.GetFullOrderHistoryParams, access interface{}) middleware.Responder {
		h.logger.Info("ListOrderStatus begin")
		ctx := p.HTTPRequest.Context()
		orderID := int(p.OrderID)

		history, err := repository.StatusHistory(ctx, orderID)
		if err != nil {
			h.logger.Error("ListOrderStatus error", zap.Error(err))
			return orders.NewGetFullOrderHistoryDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}
		haveRight := rightForHistory(access, history)
		if !haveRight {
			h.logger.Warn("User have no right to get order history", zap.Any("access", access))
			return orders.NewGetFullOrderHistoryDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "You don't have rights to see this order"}})
		}
		if len(history) == 0 {
			h.logger.Info("No order with such id", zap.Int("order_id", orderID))
			return orders.NewGetFullOrderHistoryNotFound().WithPayload("No order with such id")
		}
		result := make([]*models.OrderStatus, len(history))
		for index, status := range history {
			tmpStatus, mapErr := MapStatus(status)
			if mapErr != nil {
				h.logger.Error("ListOrderStatus error", zap.Error(mapErr))

				return orders.NewGetFullOrderHistoryDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(mapErr))
			}
			result[index] = tmpStatus
		}
		h.logger.Info("ListOrderStatus end")

		return orders.NewGetFullOrderHistoryOK().WithPayload(result)
	}
}

func MapStatus(status *ent.OrderStatus) (*models.OrderStatus, error) {
	if status == nil {
		return nil, errors.New("status is nil")
	}
	createdAt := strfmt.DateTime(status.CurrentDate)
	statusID := int64(status.ID)
	if status.Edges.StatusName == nil {
		return nil, errors.New("status name is nil")
	}
	statusName := status.Edges.StatusName.Status
	if status.Edges.Users == nil {
		return nil, errors.New("user is nil")
	}
	userID := int64(status.Edges.Users.ID)
	userName := status.Edges.Users.Name
	user := models.UserEmbeddable{
		ID:   &userID,
		Name: &userName,
	}
	tmpStatus := models.OrderStatus{
		ChangedBy: &user,
		Comment:   &status.Comment,
		CreatedAt: &createdAt,
		ID:        &statusID,
		Status:    &statusName,
	}
	return &tmpStatus, nil
}

func rightForHistory(access interface{}, history []*ent.OrderStatus) bool {
	isAdmin, err := authentication.IsAdmin(access)
	if err != nil {
		return false
	}
	if isAdmin {
		return true
	}
	ownerID, err := authentication.GetUserId(access)
	if err != nil {
		return false
	}
	for _, status := range history {
		if status.Edges.Users == nil {
			return false
		}
		if status.Edges.Users.ID == ownerID {
			return true
		}
	}
	return false
}

func (h OrderStatus) AddNewStatusToOrder(repository repositories.OrderStatusRepository) orders.AddNewOrderStatusHandlerFunc {
	return func(params orders.AddNewOrderStatusParams, access interface{}) middleware.Responder {
		h.logger.Info("AddNewStatusToOrder begin")
		ctx := params.HTTPRequest.Context()
		if params.Data == nil {
			h.logger.Warn("No data in AddNewStatusToOrder request")
			return orders.NewAddNewOrderStatusDefault(http.StatusBadRequest).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "Data is empty"}})
		}
		orderStatus := *params.Data
		status := params.Data.Status
		haveRight := rightForStatusCreation(access, status)
		if !haveRight {
			h.logger.Warn("User have no right to create order status", zap.Any("access", access))
			return orders.NewAddNewOrderStatusDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "You don't have rights to add new status"}})
		}
		userID, err := authentication.GetUserId(access)
		if err != nil {
			h.logger.Error("AddNewStatusToOrder error", zap.Error(err))
			return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't get user id"}})
		}
		err = repository.UpdateStatus(ctx, userID, orderStatus)
		if err != nil {
			h.logger.Error("AddNewStatusToOrder error", zap.Error(err))
			return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't update status"}})
		}
		h.logger.Info("AddNewStatusToOrder end")

		return orders.NewAddNewOrderStatusOK().WithPayload("all ok") // TODO: think about return values
	}
}

func rightForStatusCreation(access interface{}, status *string) bool {
	isAdmin, err := authentication.IsAdmin(access)
	if err != nil {
		return false
	}
	if isAdmin {
		return true
	}
	if status == nil {
		return false
	}
	// TODO: check other options
	return false
}

func (h OrderStatus) GetOrdersByStatus(repository repositories.OrderRepositoryWithStatusFilter) orders.GetOrdersByStatusHandlerFunc {
	return func(params orders.GetOrdersByStatusParams, access interface{}) middleware.Responder {
		h.logger.Info("GetOrdersByStatus begin")
		ctx := params.HTTPRequest.Context()
		haveRight := hasSearchRight(access)
		if !haveRight {
			h.logger.Warn("User have no right to get orders by status", zap.Any("access", access))
			return orders.NewGetOrdersByStatusDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "You don't have rights to get orders by status"}})
		}
		ordersByStatus, err := repository.OrdersByStatus(ctx, params.Status)
		if err != nil {
			h.logger.Error("GetOrdersByStatus error", zap.Error(err))
			return orders.NewGetOrdersByStatusDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: fmt.Sprintf("Can't get orders by status. error: %s", err.Error())}})
		}
		if ordersByStatus == nil {
			h.logger.Warn("No orders by status", zap.Any("access", access))
			return orders.NewGetOrdersByStatusNotFound().WithPayload("no orders by status found")
		}
		ordersResult := make([]*models.Order, len(ordersByStatus))
		for index, order := range ordersByStatus {
			tmpOrder, errMap := mapOrder(ctx, &order)
			if errMap != nil {
				h.logger.Error("GetOrdersByStatus error", zap.Error(errMap))
				return orders.NewGetOrdersByStatusDefault(http.StatusInternalServerError).
					WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't map order"}})
			}
			ordersResult[index] = tmpOrder
		}
		h.logger.Info("GetOrdersByStatus end")
		return orders.NewGetOrdersByStatusOK().WithPayload(ordersResult)
	}
}

func hasSearchRight(access interface{}) bool {
	isAdmin, err := authentication.IsAdmin(access)
	if err != nil {
		return false
	}
	return isAdmin //TODO: discuss other options
}

func (h OrderStatus) GetOrdersByPeriodAndStatus(repository repositories.OrderRepositoryWithStatusFilter) orders.GetOrdersByDateAndStatusHandlerFunc {
	return func(params orders.GetOrdersByDateAndStatusParams, access interface{}) middleware.Responder {
		h.logger.Info("GetOrdersByPeriodAndStatus begin")
		ctx := params.HTTPRequest.Context()
		haveRight := hasSearchRight(access) // TODO: discuss right management
		if !haveRight {
			h.logger.Warn("User have no right to get orders by period and status", zap.Any("access", access))
			return orders.NewGetOrdersByDateAndStatusDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "You don't have rights to get orders by period and status"}})
		}
		ordersByPeriodAndStatus, err := repository.OrdersByPeriodAndStatus(ctx, time.Time(params.FromDate), time.Time(params.ToDate), params.StatusName)
		if err != nil {
			h.logger.Error("GetOrdersByPeriodAndStatus error", zap.Error(err))
			return orders.NewGetOrdersByDateAndStatusDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't get orders by period and status"}})
		}
		ordersResult := make([]*models.Order, len(ordersByPeriodAndStatus))
		for index, order := range ordersByPeriodAndStatus {
			tmpOrder, errMap := mapOrder(ctx, &order)
			if errMap != nil {
				h.logger.Error("GetOrdersByPeriodAndStatus error", zap.Error(errMap))
				return orders.NewGetOrdersByDateAndStatusDefault(http.StatusInternalServerError).
					WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't map order"}})
			}
			ordersResult[index] = tmpOrder
		}
		h.logger.Info("GetOrdersByPeriodAndStatus end")
		return orders.NewGetOrdersByDateAndStatusOK().WithPayload(ordersResult)

	}
}
