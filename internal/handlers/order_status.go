package handlers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/roles"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func SetOrderStatusHandler(logger *zap.Logger, api *operations.BeAPI) (domain.OrderStatusRepository, domain.OrderRepositoryWithFilter, domain.EquipmentStatusRepository) {
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
	return orderStatusRepo, orderFilterRepo, equipmentStatusRepo
}

type OrderStatus struct {
	logger *zap.Logger
}

func NewOrderStatus(logger *zap.Logger) *OrderStatus {
	return &OrderStatus{
		logger: logger,
	}
}

func (h *OrderStatus) OrderStatusesHistory(repository domain.OrderStatusRepository) orders.GetFullOrderHistoryHandlerFunc {
	return func(p orders.GetFullOrderHistoryParams, principal *models.Principal) middleware.Responder {
		h.logger.Info("ListOrderStatus begin")
		ctx := p.HTTPRequest.Context()
		userID := int(principal.ID)
		role := principal.Role
		orderID := int(p.OrderID)

		history, err := repository.StatusHistory(ctx, orderID)
		if err != nil {
			h.logger.Error("ListOrderStatus error", zap.Error(err))
			return orders.NewGetFullOrderHistoryDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		if !canUserAccessOrderHistory(userID, role, history) {
			h.logger.Warn("User have no right to get order history", zap.Any("principal", principal))
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
	var statusName string
	if status.Edges.OrderStatusName == nil {
		orderStatusName, err := status.QueryOrderStatusName().Only(context.Background())
		if err != nil {
			return nil, err
		}
		statusName = orderStatusName.Status
	} else {
		statusName = status.Edges.OrderStatusName.Status
	}
	var userID int64
	var userName string
	if status.Edges.Users == nil {
		user, err := status.QueryUsers().Only(context.Background())
		if err != nil {
			return nil, err
		}
		userID = int64(user.ID)
		userName = user.Login
	} else {
		userID = int64(status.Edges.Users.ID)
		userName = status.Edges.Users.Login
	}
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

func canUserAccessOrderHistory(userID int, role string, history []*ent.OrderStatus) bool {
	if role == roles.Admin || role == roles.Manager || role == roles.Operator {
		return true
	}

	for _, status := range history {
		if status.Edges.Users != nil && status.Edges.Users.ID == userID {
			return true
		}
	}

	return false
}

func (h *OrderStatus) AddNewStatusToOrder(
	orderStatusRepo domain.OrderStatusRepository,
	equipmentStatusRepo domain.EquipmentStatusRepository,
) orders.AddNewOrderStatusHandlerFunc {
	return func(params orders.AddNewOrderStatusParams, principal *models.Principal) middleware.Responder {
		h.logger.Info("AddNewStatusToOrder begin")
		ctx := params.HTTPRequest.Context()
		userID := int(principal.ID)
		userRole := principal.Role

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

		canUserCancelOrder := canUserCancelOrder(userID, currentOrderStatus, *newOrderStatus)
		canRoleChangeStatus := canRoleChangeStatus(userRole, currentOrderStatus, *newOrderStatus)

		if !canUserCancelOrder && !canRoleChangeStatus {
			h.logger.Error("User does not have the right to create an order status", zap.Any("principal", principal))
			return orders.NewAddNewOrderStatusDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "You don't have rights to add a new status"}})
		}

		orderID := currentOrderStatus.Edges.Order.ID
		orderEquipmentStatuses, err := equipmentStatusRepo.GetEquipmentsStatusesByOrder(ctx, orderID)
		if err != nil {
			h.logger.Error("GetEquipmentStatusByOrder error", zap.Error(err))
			return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't get equipment status"))
		}

		err = checkEqStatusRequirements(*newOrderStatus, h.logger, orderEquipmentStatuses)
		if err != nil {
			return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}

		err = orderStatusRepo.UpdateStatus(ctx, userID, *params.Data)
		if err != nil {
			h.logger.Error("Update order status error", zap.Error(err))
			return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Can't update status"))
		}

		switch *newOrderStatus {
		case domain.OrderStatusRejected:
			status := domain.EquipmentStatusAvailable
			model := &models.EquipmentStatus{
				StatusName: &status,
			}
			err = UpdateEqStatuses(ctx, equipmentStatusRepo, orderEquipmentStatuses, model)
			if err != nil {
				h.logger.Error("Update equipment status error", zap.Error(err))
				return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Can't update equipment status"))
			}

		case domain.OrderStatusInProgress:
			status := domain.EquipmentStatusInUse
			model := &models.EquipmentStatus{
				StatusName: &status,
			}
			err = UpdateEqStatuses(ctx, equipmentStatusRepo, orderEquipmentStatuses, model)
			if err != nil {
				h.logger.Error("Update equipment status error", zap.Error(err))
				return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Can't update equipment status"))
			}

		case domain.OrderStatusClosed:
			status := domain.EquipmentStatusAvailable
			model := &models.EquipmentStatus{
				StatusName: &status,
			}

			if currentOrderStatus.Edges.OrderStatusName.Status == domain.OrderStatusInProgress ||
				currentOrderStatus.Edges.OrderStatusName.Status == domain.OrderStatusOverdue {
				if len(orderEquipmentStatuses) != 0 {
					addOneDayToCurrentEndDate := strfmt.DateTime(
						time.Time(orderEquipmentStatuses[0].EndDate).AddDate(0, 0, 1),
					)
					model.EndDate = &addOneDayToCurrentEndDate
				}
			}

			if currentOrderStatus.Edges.OrderStatusName.Status == domain.OrderStatusApproved ||
				currentOrderStatus.Edges.OrderStatusName.Status == domain.OrderStatusBlocked ||
				currentOrderStatus.Edges.OrderStatusName.Status == domain.OrderStatusPrepared &&
					userRole == roles.Manager {
				if len(orderEquipmentStatuses) != 0 {
					addOneDayToCurrentEndDate := strfmt.DateTime(
						time.Time(orderEquipmentStatuses[0].EndDate).AddDate(0, 0, 1),
					)
					model.EndDate = &addOneDayToCurrentEndDate
				}
			}

			err = UpdateEqStatuses(ctx, equipmentStatusRepo, orderEquipmentStatuses, model)
			if err != nil {
				h.logger.Error("Update equipment status error", zap.Error(err))
				return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Can't update equipment status"))
			}
		}

		return orders.NewAddNewOrderStatusOK().WithPayload("all ok")
	}
}

func (h *OrderStatus) GetOrdersByStatus(repository domain.OrderRepositoryWithFilter) orders.GetOrdersByStatusHandlerFunc {
	return func(params orders.GetOrdersByStatusParams, principal *models.Principal) middleware.Responder {
		ctx := params.HTTPRequest.Context()
		limit := utils.GetValueByPointerOrDefaultValue(params.Limit, math.MaxInt)
		offset := utils.GetValueByPointerOrDefaultValue(params.Offset, 0)
		orderBy := utils.GetValueByPointerOrDefaultValue(params.OrderBy, utils.AscOrder)
		orderColumn := utils.GetValueByPointerOrDefaultValue(params.OrderColumn, order.FieldID)

		total, err := repository.OrdersByStatusTotal(ctx, params.Status)
		if err != nil {
			h.logger.Error("GetOrdersByStatus error", zap.Error(err))
			return orders.NewGetOrdersByStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload(fmt.Sprintf("Can't get total count of orders by status. error: %s", err.Error())))
		}

		var ordersByStatus []*ent.Order
		if total > 0 {
			ordersByStatus, err = repository.OrdersByStatus(ctx, params.Status, int(limit), int(offset), orderBy, orderColumn)
			if err != nil {
				h.logger.Error("GetOrdersByStatus error", zap.Error(err))
				return orders.NewGetOrdersByStatusDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload(fmt.Sprintf("Can't get orders by status. error: %s", err.Error())))
			}
		}
		ordersResult := make([]*models.UserOrder, len(ordersByStatus))
		for index, order := range ordersByStatus {
			tmpOrder, errMap := mapUserOrder(order, h.logger)
			if errMap != nil {
				h.logger.Error("GetOrdersByStatus error", zap.Error(errMap))
				return orders.NewGetOrdersByStatusDefault(http.StatusInternalServerError).
					WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't map order"}})
			}
			ordersResult[index] = tmpOrder
		}

		totalOrders := int64(total)
		listOrders := &models.UserOrdersList{
			Items: ordersResult,
			Total: &totalOrders,
		}
		return orders.NewGetOrdersByStatusOK().WithPayload(listOrders)
	}
}

func (h *OrderStatus) GetOrdersByPeriodAndStatus(repository domain.OrderRepositoryWithFilter) orders.GetOrdersByDateAndStatusHandlerFunc {
	return func(params orders.GetOrdersByDateAndStatusParams, _ *models.Principal) middleware.Responder {
		ctx := params.HTTPRequest.Context()
		limit := utils.GetValueByPointerOrDefaultValue(params.Limit, math.MaxInt)
		offset := utils.GetValueByPointerOrDefaultValue(params.Offset, 0)
		orderBy := utils.GetValueByPointerOrDefaultValue(params.OrderBy, utils.AscOrder)
		orderColumn := utils.GetValueByPointerOrDefaultValue(params.OrderColumn, order.FieldID)

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
				int(limit), int(offset), orderBy, orderColumn)
			if err != nil {
				h.logger.Error("GetOrdersByPeriodAndStatus error", zap.Error(err))
				return orders.NewGetOrdersByDateAndStatusDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Can't get orders by period and status"))
			}
		}
		ordersResult := make([]*models.UserOrder, len(ordersByPeriodAndStatus))
		for index, order := range ordersByPeriodAndStatus {
			tmpOrder, errMap := mapUserOrder(order, h.logger)
			if errMap != nil {
				h.logger.Error("GetOrdersByPeriodAndStatus error", zap.Error(errMap))
				return orders.NewGetOrdersByDateAndStatusDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Can't map order"))
			}
			ordersResult[index] = tmpOrder
		}

		totalOrders := int64(total)
		listOrders := &models.UserOrdersList{
			Items: ordersResult,
			Total: &totalOrders,
		}
		return orders.NewGetOrdersByDateAndStatusOK().WithPayload(listOrders)

	}
}

func (h *OrderStatus) GetAllStatusNames(repository domain.OrderStatusNameRepository) orders.GetAllStatusNamesHandlerFunc {
	return func(params orders.GetAllStatusNamesParams, principal *models.Principal) middleware.Responder {
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

func checkEqStatusRequirements(newStatus string, logger *zap.Logger, eqStatuses []*ent.EquipmentStatus) error {
	switch newStatus {
	case domain.OrderStatusPrepared, domain.OrderStatusInProgress:
		return CheckEqStatuses(logger, eqStatuses, domain.EquipmentStatusBooked)
	}
	return nil
}

func CheckEqStatuses(logger *zap.Logger, eqStatuses []*ent.EquipmentStatus, status string) error {
	errEqIds := make([]int, 0, len(eqStatuses))
	for _, eqStatus := range eqStatuses {
		if eqStatus.Edges.EquipmentStatusName.Name != status {
			errEqIds = append(errEqIds, eqStatus.ID)
		}
	}
	if len(errEqIds) > 0 {
		logger.Error("Equipment statuses don't match expected", zap.String("status", status))
		return fmt.Errorf("equipment IDs don't have correspondent status: %v", errEqIds)
	}
	return nil
}

func UpdateEqStatuses(ctx context.Context, equipmentStatusRepo domain.EquipmentStatusRepository, eqStatuses []*ent.EquipmentStatus, equipmentStatusModel *models.EquipmentStatus) error {
	for _, eqStatus := range eqStatuses {
		eqStatusID := int64(eqStatus.ID)
		if _, err := equipmentStatusRepo.Update(ctx, &models.EquipmentStatus{
			StatusName: equipmentStatusModel.StatusName,
			ID:         &eqStatusID,
			StartDate:  equipmentStatusModel.StartDate,
			EndDate:    equipmentStatusModel.EndDate,
		}); err != nil {
			return err
		}
	}
	return nil
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

func canRoleChangeStatus(role string, currentStatus *ent.OrderStatus, newStatus string) bool {
	switch currentStatus.Edges.OrderStatusName.Status {
	case domain.OrderStatusInReview:
		return (newStatus == domain.OrderStatusApproved || newStatus == domain.OrderStatusRejected) &&
			role == roles.Manager

	case domain.OrderStatusApproved:
		switch newStatus {
		case domain.OrderStatusPrepared:
			return role == roles.Operator || role == roles.Admin
		case domain.OrderStatusClosed:
			return role == roles.Manager
		}

	case domain.OrderStatusPrepared:
		switch newStatus {
		case domain.OrderStatusInProgress:
			return role == roles.Operator
		case domain.OrderStatusClosed:
			return role == roles.Manager || role == roles.Operator
		}

	case domain.OrderStatusInProgress, domain.OrderStatusOverdue:
		if newStatus == domain.OrderStatusClosed {
			return role == roles.Manager || role == roles.Operator
		}

	case domain.OrderStatusBlocked:
		if newStatus == domain.OrderStatusClosed {
			return role == roles.Manager
		}

	}

	return role == roles.Admin ||
		role == roles.Manager ||
		role == roles.Operator
}

func canUserCancelOrder(userID int, currentStatus *ent.OrderStatus, newStatus string) bool {
	if userID != currentStatus.Edges.Order.Edges.Users.ID {
		return false
	}

	allowedStatus := map[string]bool{
		domain.OrderStatusApproved: true,
		domain.OrderStatusInReview: true,
		domain.OrderStatusPrepared: true,
	}

	return newStatus == domain.OrderStatusClosed &&
		allowedStatus[currentStatus.Edges.OrderStatusName.Status]
}
