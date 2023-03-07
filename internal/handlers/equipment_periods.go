package handlers

import (
	"net/http"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	eqPeriods "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/equipment"
	eqStatus "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/equipment_status"
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
		access interface{},
	) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		id := int(s.EquipmentID)

		equipmentStatus, err := eqStatusRepository.GetUnavailableEquipmentStatusByEquipmentID(ctx, id)
		if err != nil {
			_, notFoundError := err.(*ent.NotFoundError)
			if notFoundError {
				c.logger.Error(
					"unable to find unavailable equipment status dates by provided equipment id",
					zap.Error(err),
				)
				return eqStatus.NewCheckEquipmentStatusDefault(http.StatusNotFound).
					WithPayload(
						buildStringPayload(
							"can't find unavailable equipment status dates by provided equipment id",
						),
					)
			}

			c.logger.Error(
				"error during the search for unavailable equipment status dates",
				zap.Error(err),
			)
			return eqStatus.NewCheckEquipmentStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't find unavailable equipment status dates"))
		}

		return eqPeriods.NewGetUnavailabilityPeriodsByEquipmentIDOK().WithPayload(
			&models.EquipmentUnavailabilityPeriodsResponse{
				Data: &models.EquipmentUnavailabilityPeriods{
					EndDate:   (*strfmt.DateTime)(&equipmentStatus.EndDate),
					StartDate: (*strfmt.DateTime)(&equipmentStatus.StartDate),
				},
			})
	}
}
