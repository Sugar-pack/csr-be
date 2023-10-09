package handlers

import (
	"net/http"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	eqStatus "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/equipment_status"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/messages"
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
	return func(s eqStatus.CheckEquipmentStatusParams, principal *models.Principal) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		newStatus := s.Name.StatusName

		if !newStatusIsUnavailable(*newStatus) {
			c.logger.Error(messages.ErrWrongEqStatus, zap.Any("principal", principal))
			return orders.NewAddNewOrderStatusDefault(http.StatusBadRequest).
				WithPayload(buildBadRequestErrorPayload(messages.ErrWrongEqStatus, ""))
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
				WithPayload(buildInternalErrorPayload(messages.ErrGetEqStatusByID, err.Error()))
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
			c.logger.Error(messages.ErrOrderAndUserByEqStatusID, zap.Error(err))
			return eqStatus.NewCheckEquipmentStatusDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrOrderAndUserByEqStatusID, err.Error()))
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
	return func(s eqStatus.UpdateEquipmentStatusOnUnavailableParams, principal *models.Principal) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		newStatus := s.Name.StatusName

		if !newStatusIsUnavailable(*newStatus) {
			c.logger.Error(messages.ErrWrongEqStatus, zap.Any("principal", principal))
			return orders.NewAddNewOrderStatusDefault(http.StatusBadRequest).
				WithPayload(buildBadRequestErrorPayload(messages.ErrWrongEqStatus, ""))
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
			c.logger.Error(messages.ErrOrderAndUserByEqStatusID, zap.Error(err))
			return eqStatus.NewUpdateEquipmentStatusOnUnavailableDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrOrderAndUserByEqStatusID, err.Error()))
		}

		updatedEqStatus, err := eqStatusRepository.Update(ctx, &data)
		if err != nil {
			c.logger.Error(messages.ErrUpdateEqStatus, zap.Error(err))
			return eqStatus.NewUpdateEquipmentStatusOnUnavailableDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrUpdateEqStatus, err.Error()))
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
			c.logger.Error(messages.ErrUpdateOrderStatus, zap.Error(err))
			return orders.NewAddNewOrderStatusDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrUpdateOrderStatus, err.Error()))
		}

		eqStatusResult, err := eqStatusRepository.GetEquipmentStatusByID(
			ctx, int(*data.ID))
		if err != nil {
			c.logger.Error("receiving equipment status by id failed during changing status to unavailable", zap.Error(err))
			return eqStatus.NewCheckEquipmentStatusDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrGetEqStatusByID, err.Error()))
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
	return func(s eqStatus.UpdateEquipmentStatusOnAvailableParams, principal *models.Principal) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		newStatus := s.Name.StatusName

		if !newStatusIsAvailable(*newStatus) {
			c.logger.Error(messages.ErrWrongEqStatus, zap.Any("principal", principal))
			return orders.NewAddNewOrderStatusDefault(http.StatusBadGateway).
				WithPayload(buildBadRequestErrorPayload(messages.ErrWrongEqStatus, ""))
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
				WithPayload(buildInternalErrorPayload(messages.ErrUpdateEqStatus, err.Error()))
		}

		eqStatusResult, err := eqStatusRepository.GetEquipmentStatusByID(
			ctx, int(*data.ID))
		if err != nil {
			c.logger.Error("receiving equipment status by id failed during changing status to available", zap.Error(err))
			return eqStatus.NewCheckEquipmentStatusDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrGetEqStatusByID, err.Error()))
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
	return func(s eqStatus.UpdateRepairedEquipmentStatusDatesParams, principal *models.Principal) middleware.Responder {
		ctx := s.HTTPRequest.Context()

		existEqStatus, err := eqStatusRepository.GetEquipmentStatusByID(
			ctx, int(s.EquipmentstatusID))
		if err != nil {
			c.logger.Error("receiving equipment status by id failed during editing dates", zap.Error(err))
			return eqStatus.NewCheckEquipmentStatusDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrGetEqStatusByID, err.Error()))
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
				WithPayload(buildInternalErrorPayload(messages.ErrUpdateEqStatus, err.Error()))
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

func newStatusIsUnavailable(status string) bool {
	return status == domain.EquipmentStatusNotAvailable
}

func newStatusIsAvailable(status string) bool {
	return status == domain.EquipmentStatusAvailable
}
