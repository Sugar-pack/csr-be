package handlers

import (
	"errors"
	"fmt"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"math"
	"net/http"
	"time"

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

func SetOrderStatusHandler(client *ent.Client, logger *zap.Logger, api *operations.BeAPI) {
	orderStatusRepo := repositories.NewOrderStatusRepository(client)
	orderFilterRepo := repositories.NewOrderFilter(client)
	orderStatusNameRepo := repositories.NewStatusNameRepository(client)
	orderStatusHandler := NewOrderStatus(logger)

	api.OrdersGetOrdersByStatusHandler = orderStatusHandler.GetOrdersByStatus(orderFilterRepo)
	api.OrdersGetOrdersByDateAndStatusHandler = orderStatusHandler.GetOrdersByPeriodAndStatus(orderFilterRepo)
	api.OrdersAddNewOrderStatusHandler = orderStatusHandler.AddNewStatusToOrder(orderStatusRepo)
	api.OrdersGetFullOrderHistoryHandler = orderStatusHandler.OrderStatusesHistory(orderStatusRepo)
	api.OrdersGetAllStatusNamesHandler = orderStatusHandler.GetAllStatusNames(orderStatusNameRepo)
}

type OrderStatus struct {
	logger *zap.Logger
}

func NewOrderStatus(logger *zap.Logger) *OrderStatus {
	return &OrderStatus{
		logger: logger,
	}
}

func (h *OrderStatus) OrderStatusesHistory(repository repositories.OrderStatusRepository) orders.GetFullOrderHistoryHandlerFunc {
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
	userName := status.Edges.Users.Login
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

func (h *OrderStatus) AddNewStatusToOrder(repository repositories.OrderStatusRepository) orders.AddNewOrderStatusHandlerFunc {
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

		return orders.NewAddNewOrderStatusOK().WithPayload("all ok")
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
	return false
}

func (h *OrderStatus) GetOrdersByStatus(repository repositories.OrderRepositoryWithFilter) orders.GetOrdersByStatusHandlerFunc {
	return func(params orders.GetOrdersByStatusParams, access interface{}) middleware.Responder {
		ctx := params.HTTPRequest.Context()
		limit := utils.GetParamInt(params.Limit, math.MaxInt)
		offset := utils.GetParamInt(params.Offset, 0)
		orderBy := utils.GetParamString(params.OrderBy, utils.AscOrder)
		orderColumn := utils.GetParamString(params.OrderColumn, order.FieldID)

		haveRight := hasSearchRight(access)
		if !haveRight {
			h.logger.Warn("User have no right to get orders by status", zap.Any("access", access))
			return orders.NewGetOrdersByStatusDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "You don't have rights to get orders by status"}})
		}
		total, err := repository.OrdersByStatusTotal(ctx, params.Status)
		if err != nil {
			h.logger.Error("GetOrdersByStatus error", zap.Error(err))
			return orders.NewGetOrdersByStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload(fmt.Sprintf("Can't get total count of orders by status. error: %s", err.Error())))
		}

		var ordersByStatus []*ent.Order
		if total > 0 {
			ordersByStatus, err = repository.OrdersByStatus(ctx, params.Status, limit, offset, orderBy, orderColumn)
			if err != nil {
				h.logger.Error("GetOrdersByStatus error", zap.Error(err))
				return orders.NewGetOrdersByStatusDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload(fmt.Sprintf("Can't get orders by status. error: %s", err.Error())))
			}
		}
		ordersResult := make([]*models.Order, len(ordersByStatus))
		for index, order := range ordersByStatus {
			tmpOrder, errMap := mapOrder(order, h.logger)
			if errMap != nil {
				h.logger.Error("GetOrdersByStatus error", zap.Error(errMap))
				return orders.NewGetOrdersByStatusDefault(http.StatusInternalServerError).
					WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't map order"}})
			}
			ordersResult[index] = tmpOrder
		}

		totalOrders := int64(total)
		listOrders := &models.OrderList{
			Items: ordersResult,
			Total: &totalOrders,
		}
		return orders.NewGetOrdersByStatusOK().WithPayload(listOrders)
	}
}

func hasSearchRight(access interface{}) bool {
	isAdmin, err := authentication.IsAdmin(access)
	if err != nil {
		return false
	}
	return isAdmin
}

func (h *OrderStatus) GetOrdersByPeriodAndStatus(repository repositories.OrderRepositoryWithFilter) orders.GetOrdersByDateAndStatusHandlerFunc {
	return func(params orders.GetOrdersByDateAndStatusParams, access interface{}) middleware.Responder {
		ctx := params.HTTPRequest.Context()
		limit := utils.GetParamInt(params.Limit, math.MaxInt)
		offset := utils.GetParamInt(params.Offset, 0)
		orderBy := utils.GetParamString(params.OrderBy, utils.AscOrder)
		orderColumn := utils.GetParamString(params.OrderColumn, order.FieldID)

		haveRight := hasSearchRight(access)
		if !haveRight {
			h.logger.Warn("User have no right to get orders by period and status", zap.Any("access", access))
			return orders.NewGetOrdersByDateAndStatusDefault(http.StatusForbidden).
				WithPayload(buildStringPayload("You don't have rights to get orders by period and status"))
		}
		total, err := repository.OrdersByPeriodAndStatusTotal(ctx,
			time.Time(params.FromDate), time.Time(params.ToDate), params.StatusName)
		if err != nil {
			h.logger.Error("GetOrdersByPeriodAndStatus error", zap.Error(err))
			return orders.NewGetOrdersByDateAndStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Can't get total amount of orders by period and status"))
		}

		var ordersByPeriodAndStatus []*ent.Order
		if total > 0 {
			ordersByPeriodAndStatus, err = repository.OrdersByPeriodAndStatus(ctx,
				time.Time(params.FromDate), time.Time(params.ToDate), params.StatusName,
				limit, offset, orderBy, orderColumn)
			if err != nil {
				h.logger.Error("GetOrdersByPeriodAndStatus error", zap.Error(err))
				return orders.NewGetOrdersByDateAndStatusDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Can't get orders by period and status"))
			}
		}
		ordersResult := make([]*models.Order, len(ordersByPeriodAndStatus))
		for index, order := range ordersByPeriodAndStatus {
			tmpOrder, errMap := mapOrder(order, h.logger)
			if errMap != nil {
				h.logger.Error("GetOrdersByPeriodAndStatus error", zap.Error(errMap))
				return orders.NewGetOrdersByDateAndStatusDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Can't map order"))
			}
			ordersResult[index] = tmpOrder
		}

		totalOrders := int64(total)
		listOrders := &models.OrderList{
			Items: ordersResult,
			Total: &totalOrders,
		}
		return orders.NewGetOrdersByDateAndStatusOK().WithPayload(listOrders)

	}
}

func (h *OrderStatus) GetAllStatusNames(repository repositories.StatusNameRepository) orders.GetAllStatusNamesHandlerFunc {
	return func(params orders.GetAllStatusNamesParams, access interface{}) middleware.Responder {
		h.logger.Info("GetAllStatusNames begin")
		ctx := params.HTTPRequest.Context()
		statusNames, err := repository.ListOfStatuses(ctx)
		if err != nil {
			h.logger.Error("GetAllStatusNames error", zap.Error(err))
			return orders.NewGetAllStatusNamesDefault(http.StatusInternalServerError).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't get all status names"}})
		}
		statusNamesResult := make([]*models.OrderStatusName, len(statusNames))
		for index, statusName := range statusNames {
			tmpStatusName, errMap := MapOrderStatusName(statusName)
			if errMap != nil {
				h.logger.Error("Cant map ent order status name to model order status name", zap.Error(errMap))
				return orders.NewGetAllStatusNamesDefault(http.StatusInternalServerError).
					WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't map order status name"}})
			}
			statusNamesResult[index] = tmpStatusName
		}
		h.logger.Info("GetAllStatusNames end")
		return orders.NewGetAllStatusNamesOK().WithPayload(statusNamesResult)
	}
}

func MapOrderStatusName(status *ent.StatusName) (*models.OrderStatusName, error) {
	if status == nil {
		return nil, errors.New("status is nil")
	}
	id := int64(status.ID)
	return &models.OrderStatusName{
		ID:   &id,
		Name: &status.Status,
	}, nil
}
