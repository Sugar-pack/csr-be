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

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func SetOrderHandler(logger *zap.Logger, api *operations.BeAPI) {
	orderRepo := repositories.NewOrderRepository()
	eqStatusRepo := repositories.NewEquipmentStatusRepository()
	equipmentRepo := repositories.NewEquipmentRepository()
	ordersHandler := NewOrder(logger)

	api.OrdersGetUserOrdersHandler = ordersHandler.ListUserOrdersFunc(orderRepo)
	api.OrdersCreateOrderHandler = ordersHandler.CreateOrderFunc(orderRepo, eqStatusRepo, equipmentRepo)
	api.OrdersUpdateOrderHandler = ordersHandler.UpdateOrderFunc(orderRepo)
	api.OrdersGetAllOrdersHandler = ordersHandler.ListAllOrdersFunc(orderRepo)
}

type Order struct {
	logger *zap.Logger
}

func NewOrder(logger *zap.Logger) *Order {
	return &Order{
		logger: logger,
	}
}

func mapUserOrder(o *ent.Order, log *zap.Logger) (*models.UserOrder, error) {
	if o == nil {
		log.Warn("order is nil")
		return nil, errors.New("order is nil")
	}
	id := int64(o.ID)
	quantity := int64(o.Quantity)
	rentEnd := strfmt.DateTime(o.RentEnd)
	rentStart := strfmt.DateTime(o.RentStart)
	owner := o.Edges.Users
	equipments := o.Edges.Equipments
	if equipments == nil {
		log.Warn("order has no equipments")
		return nil, errors.New("order has no equipments")
	}
	orderEquipments := make([]*models.EquipmentResponse, len(equipments))
	for i, eq := range equipments {
		var statusId int64
		var categoryId int64
		if eq.Edges.Category != nil {
			categoryId = int64(eq.Edges.Category.ID)
		}
		var subcategoryID int64
		if eq.Edges.Subcategory != nil {
			subcategoryID = int64(eq.Edges.Subcategory.ID)
		}
		if eq.Edges.CurrentStatus != nil {
			statusId = int64(eq.Edges.CurrentStatus.ID)
		}
		var photoID string
		if eq.Edges.Photo != nil {
			photoID = eq.Edges.Photo.ID
		}

		var psID int64
		eqID := int64(eq.ID)
		if eq.Edges.PetSize != nil {
			psID = int64(eq.Edges.PetSize.ID)
		}

		var petKinds []*models.PetKind
		if eq.Edges.PetKinds != nil {
			for _, petKind := range eq.Edges.PetKinds {
				j := &models.PetKind{
					Name: &petKind.Name,
				}
				petKinds = append(petKinds, j)
			}
		}

		var eqReceiptDate int64
		if eq.ReceiptDate != "" {
			eqReceiptTime, err := time.Parse(utils.TimeFormat, eq.ReceiptDate)
			if err != nil {
				log.Error("error during parsing date string")
				return nil, err
			}
			eqReceiptDate = eqReceiptTime.Unix()
		}

		orderEquipments[i] = &models.EquipmentResponse{
			TermsOfUse:       &eq.TermsOfUse,
			CompensationCost: &eq.CompensationCost,
			Condition:        eq.Condition,
			Description:      &eq.Description,
			ID:               &eqID,
			InventoryNumber:  &eq.InventoryNumber,
			Category:         &categoryId,
			Subcategory:      subcategoryID,
			Location:         nil,
			Name:             &eq.Name,
			PhotoID:          &photoID,
			PetSize:          &psID,
			PetKinds:         petKinds,
			ReceiptDate:      &eqReceiptDate,
			Supplier:         &eq.Supplier,
			TechnicalIssues:  &eq.TechIssue,
			Title:            &eq.Title,
			Status:           &statusId,
		}
	}
	ownerId := int64(owner.ID)
	ownerName := owner.Login

	var statusToOrder *models.OrderStatus
	allStatuses := o.Edges.OrderStatus
	if len(allStatuses) != 0 {
		lastStatus := allStatuses[0]
		for _, s := range allStatuses {
			if s.CurrentDate.After(lastStatus.CurrentDate) {
				lastStatus = s
			}
		}
		mappedStatus, err := MapStatus(id, lastStatus)
		if err != nil {
			log.Error("failed to map status", zap.Error(err))
			return nil, err
		}
		statusToOrder = mappedStatus
	}

	return &models.UserOrder{
		Description: &o.Description,
		Equipments:  orderEquipments,
		ID:          &id,
		Quantity:    &quantity,
		RentEnd:     &rentEnd,
		RentStart:   &rentStart,
		User: &models.UserEmbeddable{
			ID:   &ownerId,
			Name: &ownerName,
		},
		LastStatus: statusToOrder,
		IsFirst:    &o.IsFirst,
	}, nil
}

func mapUserOrdersToResponse(entOrders []*ent.Order, log *zap.Logger) ([]*models.UserOrder, error) {
	modelOrders := make([]*models.UserOrder, len(entOrders))
	for i, o := range entOrders {
		order, err := mapUserOrder(o, log)
		if err != nil {
			log.Error("failed to map order", zap.Error(err))
			return nil, err
		}
		modelOrders[i] = order
	}

	return modelOrders, nil
}

func mapOrdersToResponse(entOrders []*ent.Order, log *zap.Logger) ([]*models.Order, error) {
	modelOrders := make([]*models.Order, len(entOrders))
	for i, o := range entOrders {
		uo, err := mapUserOrder(o, log)
		if err != nil {
			log.Error("failed to map order", zap.Error(err))
			return nil, err
		}
		user := mapUserInfoWoRole(o.Edges.Users)
		mo := &models.Order{
			Description: uo.Description,
			Equipments:  uo.Equipments,
			ID:          uo.ID,
			IsFirst:     uo.IsFirst,
			LastStatus:  uo.LastStatus,
			Quantity:    uo.Quantity,
			RentEnd:     uo.RentEnd,
			RentStart:   uo.RentStart,
			User:        user,
		}
		modelOrders[i] = mo
	}
	return modelOrders, nil
}

func (o Order) ListUserOrdersFunc(repository domain.OrderRepository) orders.GetUserOrdersHandlerFunc {
	return func(p orders.GetUserOrdersParams, principal *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		userID := int(principal.ID)
		limit := utils.GetValueByPointerOrDefaultValue(p.Limit, math.MaxInt)
		offset := utils.GetValueByPointerOrDefaultValue(p.Offset, 0)
		orderBy := utils.GetValueByPointerOrDefaultValue(p.OrderBy, utils.AscOrder)
		orderColumn := utils.GetValueByPointerOrDefaultValue(p.OrderColumn, order.FieldID)

		orderFilter := domain.OrderFilter{
			Filter: domain.Filter{
				Limit:       int(limit),
				Offset:      int(offset),
				OrderBy:     orderBy,
				OrderColumn: orderColumn,
			},
			Status: p.Status,
		}
		if p.Status != nil {
			_, ok := domain.AllOrderStatuses[*p.Status]
			if !ok {
				return orders.NewGetUserOrdersDefault(http.StatusBadRequest).
					WithPayload(buildStringPayload(fmt.Sprintf("Invalid order status '%v'", *p.Status)))
			}
		}

		total, err := repository.OrdersTotal(ctx, &userID)
		if err != nil {
			o.logger.Error("Error while getting total of all user's orders", zap.Error(err))
			return orders.NewGetUserOrdersDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}

		var items []*ent.Order
		if total > 0 {
			items, err = repository.List(ctx, &userID, orderFilter)
			if err != nil {
				o.logger.Error("list items failed", zap.Error(err))
				return orders.NewGetUserOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
			}
		}

		mappedOrders, err := mapUserOrdersToResponse(items, o.logger)
		if err != nil {
			o.logger.Error("map orders to response failed", zap.Error(err))
			return orders.NewGetUserOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}
		totalOrders := int64(total)
		listOrders := &models.UserOrdersList{
			Items: mappedOrders,
			Total: &totalOrders,
		}
		return orders.NewGetUserOrdersOK().WithPayload(listOrders)
	}
}

func (o Order) ListAllOrdersFunc(repository domain.OrderRepository) orders.GetAllOrdersHandlerFunc {
	return func(p orders.GetAllOrdersParams, _ *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		limit := utils.GetValueByPointerOrDefaultValue(p.Limit, math.MaxInt)
		offset := utils.GetValueByPointerOrDefaultValue(p.Offset, 0)
		orderBy := utils.GetValueByPointerOrDefaultValue(p.OrderBy, utils.AscOrder)
		orderColumn := utils.GetValueByPointerOrDefaultValue(p.OrderColumn, order.FieldID)

		orderFilter := domain.OrderFilter{
			Filter: domain.Filter{
				Limit:       int(limit),
				Offset:      int(offset),
				OrderBy:     orderBy,
				OrderColumn: orderColumn,
			},
		}

		if p.Status != nil {
			_, ok := domain.AllOrderStatuses[*p.Status]
			if !ok {
				return orders.NewGetAllOrdersDefault(http.StatusBadRequest).
					WithPayload(buildStringPayload(fmt.Sprintf("Invalid order status '%v'", *p.Status)))
			}
			orderFilter.Status = p.Status
		}

		if p.EquipmentID != nil {
			eid := int(*p.EquipmentID)
			orderFilter.EquipmentID = &eid
		}

		total, err := repository.OrdersTotal(ctx, nil)
		if err != nil {
			o.logger.Error("Error while getting total of all orders", zap.Error(err))
			return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}

		var items []*ent.Order
		if total > 0 {
			items, err = repository.List(ctx, nil, orderFilter)
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
		listOrders := &models.OrdersList{
			Items: mappedOrders,
			Total: &totalOrders,
		}
		return orders.NewGetAllOrdersOK().WithPayload(listOrders)
	}
}

func (o Order) CreateOrderFunc(
	orderRepo domain.OrderRepository,
	eqStatusRepo domain.EquipmentStatusRepository,
	equipmentRepo domain.EquipmentRepository,
) orders.CreateOrderHandlerFunc {
	return func(p orders.CreateOrderParams, principal *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		userID := int(principal.ID)

		id := int(*p.Data.EquipmentID)
		isEquipmentAvailable, err := eqStatusRepo.HasStatusByPeriod(ctx, domain.EquipmentStatusAvailable, id,
			time.Time(*p.Data.RentStart), time.Time(*p.Data.RentEnd))
		if err != nil {
			o.logger.Error("error while checking if equipment is available for period", zap.Error(err))
			return orders.NewCreateOrderDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		if !isEquipmentAvailable {
			o.logger.Warn("equipment is not free")
			return orders.NewCreateOrderDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("requested equipment is not free"))
		}

		order, err := orderRepo.Create(ctx, p.Data, userID, []int{id})
		if err != nil {
			o.logger.Error("map orders to response failed", zap.Error(err))
			return orders.NewCreateOrderDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		equipmentBookedStartDate := strfmt.DateTime(time.Time(*p.Data.RentStart).AddDate(0, 0, -1))
		equipmentBookedEndDate := strfmt.DateTime(time.Time(*p.Data.RentEnd).AddDate(0, 0, 1))
		eqID := int64(id)
		if _, err = eqStatusRepo.Create(ctx, &models.NewEquipmentStatus{
			EquipmentID: &eqID,
			StartDate:   &equipmentBookedStartDate,
			EndDate:     &equipmentBookedEndDate,
			StatusName:  &domain.EquipmentStatusBooked,
			OrderID:     int64(order.ID),
		}); err != nil {
			o.logger.Error("error while creating equipment status", zap.Error(err))
			return orders.NewGetUserOrdersDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}

		mappedOrder, err := mapUserOrder(order, o.logger)
		if err != nil {
			o.logger.Error("failed to map order", zap.Error(err))
			return orders.NewGetUserOrdersDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		return orders.NewCreateOrderCreated().WithPayload(mappedOrder)
	}
}

func (o Order) UpdateOrderFunc(repository domain.OrderRepository) orders.UpdateOrderHandlerFunc {
	return func(p orders.UpdateOrderParams, principal *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		userID := int(principal.ID)
		orderID := int(p.OrderID)

		order, err := repository.Update(ctx, orderID, p.Data, userID)
		if err != nil {
			o.logger.Error("update order failed", zap.Error(err))
			return orders.NewUpdateOrderDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		mappedOrder, err := mapUserOrder(order, o.logger)
		if err != nil {
			o.logger.Error("failed to map order", zap.Error(err))
			return orders.NewUpdateOrderDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}

		return orders.NewUpdateOrderOK().WithPayload(mappedOrder)
	}
}
