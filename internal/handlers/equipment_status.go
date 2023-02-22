package handlers

import (
	"net/http"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	eqStatus "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/equipment_status"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"go.uber.org/zap"
)

// timeNowEquipmentStatus necessary for mock time in tests
var timeNowEquipmentStatus = time.Now

const (
	EQUIPMENT_UNDER_REPAIR_COMMENT_FOR_ORDER = "Equipment under repair"
)

func SetEquipmentStatusHandler(logger *zap.Logger, api *operations.BeAPI) {
	equipmentStatusRepo := repositories.NewEquipmentStatusRepository()
	orderStatusRepo := repositories.NewOrderStatusRepository()

	statusHandler := NewEquipmentStatus(logger)

	api.EquipmentStatusUpdateEquipmentStatusOnUnavailableHandler = statusHandler.PutEquipmentStatusInRepairFunc(equipmentStatusRepo, orderStatusRepo)
	api.EquipmentStatusUpdateEquipmentStatusOnAvailableHandler = statusHandler.DeleteEquipmentStatusFromRepairFunc(equipmentStatusRepo, orderStatusRepo)
	api.EquipmentStatusCheckEquipmentStatusHandler = statusHandler.GetEquipmentStatusCheckDatesFunc(equipmentStatusRepo)
	api.EquipmentStatusUpdateRepairedEquipmentStatusDatesHandler = statusHandler.PatchEquipmentStatusEditDatesFunc(equipmentStatusRepo)
}

type EquipmentStatus struct {
	logger *zap.Logger
}

func NewEquipmentStatus(logger *zap.Logger) *EquipmentStatus {
	return &EquipmentStatus{
		logger: logger,
	}
}

func (c EquipmentStatus) GetEquipmentStatusCheckDatesFunc(
	eqStatusRepository domain.EquipmentStatusRepository) eqStatus.CheckEquipmentStatusHandlerFunc {
	return func(s eqStatus.CheckEquipmentStatusParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		newStatus := s.Name.StatusName

		errResponder := c.checkEquipmentStatusAccessRights(access)
		if errResponder != nil {
			return errResponder
		}

		if !newStatusIsUnavailable(*newStatus) {
			c.logger.Error("Wrong new equipment status, status should be only 'not available'", zap.Any("access", access))
			return orders.NewAddNewOrderStatusDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "Wrong new equipment status, status should be only 'not available'"}})
		}

		data := models.EquipmentStatus{
			EndDate:    s.Name.EndDate,
			StartDate:  s.Name.StartDate,
			StatusName: newStatus,
			ID:         &s.EquipmentstatusID,
		}

		eqStatusResult, err := eqStatusRepository.GetEquipmentStatusByID(
			ctx, int(*data.ID))
		if err != nil {
			c.logger.Error("receiving equipment status by id failed during checking start/end dates", zap.Error(err))
			return eqStatus.NewCheckEquipmentStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't find equipment status by provided id"))
		}

		if !eqStatusResult.EndDate.After(time.Time(*data.StartDate)) &&
			eqStatusResult.StartDate.Before(time.Time(*data.EndDate)) ||
			!eqStatusResult.StartDate.Before(time.Time(*data.EndDate)) {
			return eqStatus.NewCheckEquipmentStatusOK().WithPayload(
				&models.EquipmentStatusRepairConfirmationResponse{})
		}

		orderResult, userResult, err := eqStatusRepository.GetOrderAndUserByEquipmentStatusID(
			ctx, int(*data.ID))
		if err != nil {
			c.logger.Error("receiving order and user data failed", zap.Error(err))
			return eqStatus.NewCheckEquipmentStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't receive order and user data during checking equipment status"))
		}

		if orderResult == nil && userResult == nil {
			return eqStatus.NewCheckEquipmentStatusOK().WithPayload(
				&models.EquipmentStatusRepairConfirmationResponse{})
		}

		orderID := int64(orderResult.ID)
		equipmentID := int64(eqStatusResult.Edges.Equipments.ID)
		return eqStatus.NewCheckEquipmentStatusOK().WithPayload(
			&models.EquipmentStatusRepairConfirmationResponse{
				Data: &models.EquipmentStatusRepairConfirmation{
					EquipmentStatusID: data.ID,
					EndDate:           (*strfmt.DateTime)(&eqStatusResult.EndDate),
					StartDate:         (*strfmt.DateTime)(&eqStatusResult.StartDate),
					StatusName:        &eqStatusResult.Edges.EquipmentStatusName.Name,
					OrderID:           &orderID,
					UserEmail:         &userResult.Email,
					EquipmentID:       &equipmentID,
				},
			})
	}
}

func (c EquipmentStatus) PutEquipmentStatusInRepairFunc(
	eqStatusRepository domain.EquipmentStatusRepository,
	orderStatusRepo domain.OrderStatusRepository) eqStatus.UpdateEquipmentStatusOnUnavailableHandlerFunc {
	return func(s eqStatus.UpdateEquipmentStatusOnUnavailableParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		newStatus := s.Name.StatusName

		errResponder := c.checkEquipmentStatusAccessRights(access)
		if errResponder != nil {
			return errResponder
		}

		if !newStatusIsUnavailable(*newStatus) {
			c.logger.Error("Wrong new equipment status, status should be only 'not available'", zap.Any("access", access))
			return orders.NewAddNewOrderStatusDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "Wrong new equipment status, status should be only 'not available'"}})
		}

		reduceOneDayFromCurrentStartDate := strfmt.DateTime(
			time.Time(*s.Name.StartDate).AddDate(0, 0, -1),
		)

		addOneDayToCurrentEndDate := strfmt.DateTime(
			time.Time(*s.Name.EndDate).AddDate(0, 0, 1),
		)

		data := models.EquipmentStatus{
			StartDate:  &reduceOneDayFromCurrentStartDate,
			EndDate:    &addOneDayToCurrentEndDate,
			StatusName: newStatus,
			ID:         &s.EquipmentstatusID,
		}

		orderResult, userResult, err := eqStatusRepository.GetOrderAndUserByEquipmentStatusID(
			ctx, int(*data.ID))
		if err != nil {
			c.logger.Error("receiving user and order status failed", zap.Error(err))
			return eqStatus.NewUpdateEquipmentStatusOnUnavailableDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't receive order and user for updating equipment status on unavailable"))
		}

		updatedEqStatus, err := eqStatusRepository.Update(ctx, &data)
		if err != nil {
			c.logger.Error("update equipment status failed", zap.Error(err))
			return eqStatus.NewUpdateEquipmentStatusOnUnavailableDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't update equipment status"))
		}

		comment := EQUIPMENT_UNDER_REPAIR_COMMENT_FOR_ORDER
		timeNow := timeNowEquipmentStatus()
		orderID := int64(orderResult.ID)
		model := models.NewOrderStatus{
			Comment:   &comment,
			CreatedAt: (*strfmt.DateTime)(&timeNow),
			OrderID:   &orderID,
			Status:    &domain.OrderStatusRejected,
		}

		err = orderStatusRepo.UpdateStatus(ctx, userResult.ID, model)
		if err != nil {
			c.logger.Error("Update order status error", zap.Error(err))
			return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Can't update order status"))
		}

		eqStatusResult, err := eqStatusRepository.GetEquipmentStatusByID(
			ctx, int(*data.ID))
		if err != nil {
			c.logger.Error("receiving equipment status by id failed during changing status to unavailable", zap.Error(err))
			return eqStatus.NewCheckEquipmentStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't find equipment status by provided id"))
		}

		equipmentStatusID := int64(updatedEqStatus.ID)
		equipmentID := int64(eqStatusResult.Edges.Equipments.ID)
		return eqStatus.NewUpdateEquipmentStatusOnUnavailableOK().WithPayload(
			&models.EquipmentStatusRepairResponse{
				Data: &models.EquipmentStatus{
					ID:          &equipmentStatusID,
					EndDate:     (*strfmt.DateTime)(&updatedEqStatus.EndDate),
					StartDate:   (*strfmt.DateTime)(&updatedEqStatus.StartDate),
					StatusName:  &eqStatusResult.Edges.EquipmentStatusName.Name,
					EquipmentID: &equipmentID,
					CreatedAt:   strfmt.DateTime(eqStatusResult.CreatedAt),
				},
			})
	}
}

func (c EquipmentStatus) DeleteEquipmentStatusFromRepairFunc(
	eqStatusRepository domain.EquipmentStatusRepository,
	orderStatusRepo domain.OrderStatusRepository) eqStatus.UpdateEquipmentStatusOnAvailableHandlerFunc {
	return func(s eqStatus.UpdateEquipmentStatusOnAvailableParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		newStatus := s.Name.StatusName

		errResponder := c.checkEquipmentStatusAccessRights(access)
		if errResponder != nil {
			return errResponder
		}

		if !newStatusIsAvailable(*newStatus) {
			c.logger.Error("Wrong new equipment status, status should be only 'available'", zap.Any("access", access))
			return orders.NewAddNewOrderStatusDefault(http.StatusForbidden).
				WithPayload(&models.Error{Data: &models.ErrorData{Message: "Wrong new equipment status, status should be only 'not available'"}})
		}

		timeNow := timeNowEquipmentStatus()
		addOneDayToCurrentDate := strfmt.DateTime(
			time.Time(timeNow).AddDate(0, 0, 1),
		)
		data := models.EquipmentStatus{
			EndDate:    (*strfmt.DateTime)(&addOneDayToCurrentDate),
			StatusName: newStatus,
			ID:         &s.EquipmentstatusID,
		}

		updatedEqStatus, err := eqStatusRepository.Update(ctx, &data)
		if err != nil {
			c.logger.Error("update equipment on available status failed", zap.Error(err))
			return eqStatus.NewUpdateEquipmentStatusOnAvailableDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't update equipment status on available status"))
		}

		eqStatusResult, err := eqStatusRepository.GetEquipmentStatusByID(
			ctx, int(*data.ID))
		if err != nil {
			c.logger.Error("receiving equipment status by id failed during changing status to available", zap.Error(err))
			return eqStatus.NewCheckEquipmentStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't find equipment status by provided id"))
		}

		equipmentStatusID := int64(updatedEqStatus.ID)
		equipmentID := int64(eqStatusResult.Edges.Equipments.ID)
		return eqStatus.NewUpdateEquipmentStatusOnAvailableOK().WithPayload(
			&models.EquipmentStatusRepairResponse{
				Data: &models.EquipmentStatus{
					ID:          &equipmentStatusID,
					EndDate:     (*strfmt.DateTime)(&updatedEqStatus.EndDate),
					StartDate:   (*strfmt.DateTime)(&updatedEqStatus.StartDate),
					StatusName:  &eqStatusResult.Edges.EquipmentStatusName.Name,
					EquipmentID: &equipmentID,
					CreatedAt:   strfmt.DateTime(eqStatusResult.CreatedAt),
				},
			})
	}
}

func (c EquipmentStatus) PatchEquipmentStatusEditDatesFunc(
	eqStatusRepository domain.EquipmentStatusRepository,
) eqStatus.UpdateRepairedEquipmentStatusDatesHandlerFunc {
	return func(s eqStatus.UpdateRepairedEquipmentStatusDatesParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()

		errResponder := c.checkEquipmentStatusAccessRights(access)
		if errResponder != nil {
			return errResponder
		}

		existEqStatus, err := eqStatusRepository.GetEquipmentStatusByID(
			ctx, int(s.EquipmentstatusID))
		if err != nil {
			c.logger.Error("receiving equipment status by id failed during editing dates", zap.Error(err))
			return eqStatus.NewCheckEquipmentStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't find equipment status by provided id"))
		}

		if !time.Time(s.Name.StartDate).IsZero() {
			existEqStatus.StartDate = time.Time(s.Name.StartDate).AddDate(0, 0, -1)
		}

		if !time.Time(s.Name.EndDate).IsZero() {
			existEqStatus.EndDate = time.Time(s.Name.EndDate).AddDate(0, 0, 1)
		}

		data := models.EquipmentStatus{
			StartDate: (*strfmt.DateTime)(&existEqStatus.StartDate),
			EndDate:   (*strfmt.DateTime)(&existEqStatus.EndDate),
			ID:        &s.EquipmentstatusID,
		}

		updatedEqStatus, err := eqStatusRepository.Update(ctx, &data)
		if err != nil {
			c.logger.Error("update equipment on available status failed during editing dates", zap.Error(err))
			return eqStatus.NewUpdateRepairedEquipmentStatusDatesDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't update equipment status on available status during editing dates"))
		}

		equipmentStatusID := int64(updatedEqStatus.ID)
		equipmentID := int64(existEqStatus.Edges.Equipments.ID)
		return eqStatus.NewUpdateRepairedEquipmentStatusDatesOK().WithPayload(
			&models.EquipmentStatusRepairResponse{
				Data: &models.EquipmentStatus{
					ID:          &equipmentStatusID,
					EndDate:     (*strfmt.DateTime)(&updatedEqStatus.EndDate),
					StartDate:   (*strfmt.DateTime)(&updatedEqStatus.StartDate),
					StatusName:  &existEqStatus.Edges.EquipmentStatusName.Name,
					EquipmentID: &equipmentID,
					CreatedAt:   strfmt.DateTime(existEqStatus.CreatedAt),
				},
			})
	}
}

func (c EquipmentStatus) checkEquipmentStatusAccessRights(
	access interface{}) middleware.Responder {
	ok, err := equipmentStatusAccessRights(access)
	if err != nil {
		c.logger.Error("Error while getting authorization", zap.Error(err))
		return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
			WithPayload(&models.Error{Data: &models.ErrorData{Message: "Can't get authorization"}})

	}

	if !ok {
		c.logger.Error("User have no right to change equipment status",
			zap.Any("access", access))
		return orders.NewAddNewOrderStatusDefault(http.StatusForbidden).
			WithPayload(&models.Error{Data: &models.ErrorData{
				Message: "You don't have rights to change equipment status"}},
			)
	}

	return nil
}

func equipmentStatusAccessRights(access interface{}) (bool, error) {
	isManager, err := authentication.IsManager(access)
	if err != nil {
		return false, err
	}

	return isManager, nil
}

func newStatusIsUnavailable(status string) bool {
	return status == domain.EquipmentStatusNotAvailable
}

func newStatusIsAvailable(status string) bool {
	return status == domain.EquipmentStatusAvailable
}
