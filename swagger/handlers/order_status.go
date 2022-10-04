package handlers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

func SetOrderStatusHandler(logger *zap.Logger, api *operations.BeAPI) {
	orderStatusRepo := repositories.NewOrderStatusRepository()
	orderFilterRepo := repositories.NewOrderFilter()
	orderStatusNameRepo := repositories.NewOrderStatusNameRepository()
	equipmentStatusRepo := repositories.NewEquipmentStatusRepository()
	orderStatusHandler := NewOrderStatus(logger)

	api.OrdersGetOrdersByStatusHandler = orderStatusHandler.GetOrdersByStatus(orderFilterRepo)
	api.OrdersGetOrdersByDateAndStatusHandler = orderStatusHandler.GetOrdersByPeriodAndStatus(orderFilterRepo)
	api.OrdersAddNewOrderStatusHandler = orderStatusHandler.AddNewStatusToOrder(orderStatusRepo, equipmentStatusRepo)
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
			tmpStatus, mapErr := MapStatus(int64(orderID), status)
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

func MapStatus(orderID int64, status *ent.OrderStatus) (*models.OrderStatus, error) {
	if status == nil {
		return nil, errors.New("status is nil")
	}
	createdAt := strfmt.DateTime(status.CurrentDate)
	statusID := int64(status.ID)
	if status.Edges.OrderStatusName == nil {
		return nil, errors.New("status name is nil")
	}
	statusName := status.Edges.OrderStatusName.Status
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
		OrderID:   &orderID,
	}
	return &tmpStatus, nil
}

func rightForHistory(access interface{}, history []*ent.OrderStatus) bool {
	if orderStatusAccessRights(access) {
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

func (h *OrderStatus) AddNewStatusToOrder(orderStatusRepo repositories.OrderStatusRepository,
	equipmentStatusRepo repositories.EquipmentStatusRepository) orders.AddNewOrderStatusHandlerFunc {
	return func(params orders.AddNewOrderStatusParams, access interface{}) middleware.Responder {
		h.logger.Info("AddNewStatusToOrder begin")
		ctx := params.HTTPRequest.Context()
		if !orderStatusAccessRights(access) {
			h.logger.Error("User have no right to create order status", zap.Any("access", access))
			return orders.NewAddNewOrderStatusDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "You don't have rights to add new status"}})
		}

		newOrderStatus := params.Data.Status
		if newOrderStatus == nil {
			h.logger.Error("Status is empty")
			return orders.NewAddNewOrderStatusDefault(http.StatusBadRequest).
				WithPayload(buildStringPayload("Status is empty"))
		}
		currentOrderStatus, err := orderStatusRepo.GetOrderCurrentStatus(ctx, int(*params.Data.OrderID))
		if err != nil {
			h.logger.Error("getting order current status failed", zap.Error(err))
			return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Can't get order current status"))
		}

		haveRight := rightForStatusCreation(access, currentOrderStatus.Edges.OrderStatusName.Status, *newOrderStatus)
		if !haveRight {
			h.logger.Error("User have no right to create order status", zap.Any("access", access))
			return orders.NewAddNewOrderStatusDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "You don't have rights to add new status"}})
		}
		userID, err := authentication.GetUserId(access)
		if err != nil {
			h.logger.Error("AddNewStatusToOrder error", zap.Error(err))
			return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Can't get user id"))
		}

		orderID := currentOrderStatus.Edges.Order.ID
		orderEquipmentStatuses, err := equipmentStatusRepo.GetEquipmentsStatusesByOrder(ctx, orderID)
		if err != nil {
			h.logger.Error("GetEquipmentStatusByOrder error", zap.Error(err))
			return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't get equipment status"))
		}

		err = orderStatusRepo.UpdateStatus(ctx, userID, *params.Data)
		if err != nil {
			h.logger.Error("Update order status error", zap.Error(err))
			return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Can't update status"))
		}

		switch *newOrderStatus {
		case repositories.OrderStatusRejected:
			for _, eqStatus := range orderEquipmentStatuses {
				eqStatusID := int64(eqStatus.ID)
				_, err = equipmentStatusRepo.Update(ctx, &models.EquipmentStatus{
					StatusName: &repositories.EquipmentStatusAvailable,
					ID:         &eqStatusID,
				})
				if err != nil {
					h.logger.Error("Update equipment status error", zap.Error(err))
					return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
						WithPayload(buildStringPayload("Can't update equipment status"))
				}
			}
		}

		return orders.NewAddNewOrderStatusOK().WithPayload("all ok")
	}
}

func (h *OrderStatus) GetOrdersByStatus(repository repositories.OrderRepositoryWithFilter) orders.GetOrdersByStatusHandlerFunc {
	return func(params orders.GetOrdersByStatusParams, access interface{}) middleware.Responder {
		ctx := params.HTTPRequest.Context()
		limit := utils.GetParamInt(params.Limit, math.MaxInt)
		offset := utils.GetParamInt(params.Offset, 0)
		orderBy := utils.GetParamString(params.OrderBy, utils.AscOrder)
		orderColumn := utils.GetParamString(params.OrderColumn, order.FieldID)

		haveRight := orderStatusAccessRights(access)
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

func (h *OrderStatus) GetOrdersByPeriodAndStatus(repository repositories.OrderRepositoryWithFilter) orders.GetOrdersByDateAndStatusHandlerFunc {
	return func(params orders.GetOrdersByDateAndStatusParams, access interface{}) middleware.Responder {
		ctx := params.HTTPRequest.Context()
		limit := utils.GetParamInt(params.Limit, math.MaxInt)
		offset := utils.GetParamInt(params.Offset, 0)
		orderBy := utils.GetParamString(params.OrderBy, utils.AscOrder)
		orderColumn := utils.GetParamString(params.OrderColumn, order.FieldID)

		haveRight := orderStatusAccessRights(access)
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

func (h *OrderStatus) GetAllStatusNames(repository repositories.OrderStatusNameRepository) orders.GetAllStatusNamesHandlerFunc {
	return func(params orders.GetAllStatusNamesParams, access interface{}) middleware.Responder {
		h.logger.Info("GetAllStatusNames begin")
		ctx := params.HTTPRequest.Context()
		statusNames, err := repository.ListOfOrderStatusNames(ctx)
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

func MapOrderStatusName(status *ent.OrderStatusName) (*models.OrderStatusName, error) {
	if status == nil {
		return nil, errors.New("status is nil")
	}
	id := int64(status.ID)
	return &models.OrderStatusName{
		ID:   &id,
		Name: &status.Status,
	}, nil
}

func orderStatusAccessRights(access interface{}) bool {
	isAdmin, _ := authentication.IsAdmin(access)
	isManager, _ := authentication.IsManager(access)
	isOperator, _ := authentication.IsOperator(access)
	return isAdmin || isManager || isOperator
}

func rightForStatusCreation(access interface{}, currentStatus, newStatus string) bool {
	isAdmin, _ := authentication.IsAdmin(access)
	isManager, _ := authentication.IsManager(access)
	isOperator, _ := authentication.IsOperator(access)
	switch currentStatus {
	case repositories.OrderStatusInReview:
		if newStatus == repositories.OrderStatusApproved || newStatus == repositories.OrderStatusRejected {
			return isManager
		}
		return false
	default:
		return isAdmin || isManager || isOperator
	}
}
