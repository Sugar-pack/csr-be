package handlers

import (
	"net/http"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	eqPeriods "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/equipment"
	eqStatus "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/equipment_status"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/messages"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"go.uber.org/zap"
)

func SetEquipmentPeriodsHandler(logger *zap.Logger, api *operations.BeAPI) {
	equipmentStatusRepo := repositories.NewEquipmentStatusRepository()
	equipmentPeriodsHandler := NewEquipmentPeriods(logger)

	api.EquipmentGetUnavailabilityPeriodsByEquipmentIDHandler = equipmentPeriodsHandler.
		GetEquipmentUnavailableDatesFunc(equipmentStatusRepo)
}

type EquipmentPeriods struct {
	logger *zap.Logger
}

func NewEquipmentPeriods(logger *zap.Logger) *EquipmentPeriods {
	return &EquipmentPeriods{
		logger: logger,
	}
}

func (c EquipmentPeriods) GetEquipmentUnavailableDatesFunc(
	eqStatusRepository domain.EquipmentStatusRepository,
) eqPeriods.GetUnavailabilityPeriodsByEquipmentIDHandlerFunc {
	return func(
		s eqPeriods.GetUnavailabilityPeriodsByEquipmentIDParams,
		_ *models.Principal,
	) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		id := int(s.EquipmentID)

		equipmentStatuses, err := eqStatusRepository.GetUnavailableEquipmentStatusByEquipmentID(ctx, id)
		if err != nil {
			c.logger.Error(messages.ErrGetUnavailableEqStatus, zap.Error(err))
			return eqStatus.NewCheckEquipmentStatusDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrGetUnavailableEqStatus, err.Error()))
		}

		var result []*models.EquipmentUnavailabilityPeriods

		for _, value := range equipmentStatuses {
			var unavPeriod models.EquipmentUnavailabilityPeriods
			unavPeriod.StartDate = (*strfmt.DateTime)(&value.StartDate)
			unavPeriod.EndDate = (*strfmt.DateTime)(&value.EndDate)

			result = append(result, &unavPeriod)
		}

		return eqPeriods.NewGetUnavailabilityPeriodsByEquipmentIDOK().WithPayload(
			&models.EquipmentUnavailabilityPeriodsResponse{Items: result},
		)
	}
}
