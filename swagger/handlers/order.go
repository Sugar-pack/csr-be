package handlers

import (
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	equipmentEnt "git.epam.com/epm-lstr/epm-lstr-lc/be/ent/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

func SetOrderHandler(logger *zap.Logger, api *operations.BeAPI) {
	orderRepo := repositories.NewOrderRepository()
	eqStatusRepo := repositories.NewEquipmentStatusRepository()
	equipmentRepo := repositories.NewEquipmentRepository()
	ordersHandler := NewOrder(logger)

	api.OrdersGetAllOrdersHandler = ordersHandler.ListOrderFunc(orderRepo)
	api.OrdersCreateOrderHandler = ordersHandler.CreateOrderFunc(orderRepo, eqStatusRepo, equipmentRepo)
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
					ID:   int64(petKind.ID),
					Name: &petKind.Name,
				}
				petKinds = append(petKinds, j)
			}
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
			ReceiptDate:      &eq.ReceiptDate,
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

	return &models.Order{
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

func (o Order) CreateOrderFunc(orderRepo repositories.OrderRepository,
	eqStatusRepo repositories.EquipmentStatusRepository,
	equipmentRepo repositories.EquipmentRepository) orders.CreateOrderHandlerFunc {
	return func(p orders.CreateOrderParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		ownerId, err := authentication.GetUserId(access)
		if err != nil {
			o.logger.Error("get user ID failed", zap.Error(err))
			return orders.NewCreateOrderDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}
		equipmentFilter := models.EquipmentFilter{Category: *p.Data.Category}
		if p.Data.Subcategory > 0 {
			equipmentFilter.Subcategory = p.Data.Subcategory
		}
		equipmentsByCategory, err := equipmentRepo.EquipmentsByFilter(ctx, equipmentFilter,
			math.MaxInt, 0, utils.DescOrder, equipmentEnt.FieldID)
		if err != nil {
			o.logger.Error("error while getting equipments in category and subcategory", zap.Error(err))
			return orders.NewCreateOrderDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}
		var availableEquipments []int
		for _, eq := range equipmentsByCategory {
			isEquipmentAvailable, err := eqStatusRepo.HasStatusByPeriod(ctx, repositories.EquipmentStatusAvailable, eq.ID,
				time.Time(*p.Data.RentStart), time.Time(*p.Data.RentEnd))
			if err != nil {
				o.logger.Error("error while checking if equipment is available for period", zap.Error(err))
				return orders.NewCreateOrderDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
			}
			if isEquipmentAvailable {
				availableEquipments = append(availableEquipments, eq.ID)
			}
		}
		if len(availableEquipments) < int(*p.Data.Quantity) {
			o.logger.Warn("there is not so much free equipment")
			return orders.NewCreateOrderDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("there is not so much free equipment"))
		}
		equipmentsToOrder := availableEquipments[:int(*p.Data.Quantity)]

		order, err := orderRepo.Create(ctx, p.Data, ownerId, equipmentsToOrder)
		if err != nil {
			o.logger.Error("map orders to response failed", zap.Error(err))
			return orders.NewCreateOrderDefault(http.StatusInternalServerError).WithPayload(buildErrorPayload(err))
		}
		endDate := time.Time(*p.Data.RentEnd).AddDate(0, 0, 1)
		equipmentBookedEndDate := strfmt.DateTime(endDate)
		for _, equipmentID := range equipmentsToOrder {
			eqID := int64(equipmentID)
			_, err = eqStatusRepo.Create(ctx, &models.NewEquipmentStatus{
				EndDate:     &equipmentBookedEndDate,
				EquipmentID: &eqID,
				StartDate:   p.Data.RentStart,
				StatusName:  &repositories.EquipmentStatusBooked,
				OrderID:     int64(order.ID),
			})
			if err != nil {
				o.logger.Error("error while creating equipment status", zap.Error(err))
				return orders.NewGetAllOrdersDefault(http.StatusInternalServerError).
					WithPayload(buildErrorPayload(err))
			}
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
