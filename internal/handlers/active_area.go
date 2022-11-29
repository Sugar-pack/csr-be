package handlers

import (
	"math"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/active_areas"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func SetActiveAreaHandler(logger *zap.Logger, api *operations.BeAPI) {
	activeAreaRepo := repositories.NewActiveAreaRepository()
	activeAreaHandler := NewActiveArea(logger)
	api.ActiveAreasGetAllActiveAreasHandler = activeAreaHandler.GetActiveAreasFunc(activeAreaRepo)
}

type ActiveArea struct {
	logger *zap.Logger
}

func NewActiveArea(logger *zap.Logger) *ActiveArea {
	return &ActiveArea{
		logger: logger,
	}
}

func (area ActiveArea) GetActiveAreasFunc(repository domain.ActiveAreaRepository) active_areas.GetAllActiveAreasHandlerFunc {
	return func(a active_areas.GetAllActiveAreasParams, access interface{}) middleware.Responder {
		ctx := a.HTTPRequest.Context()
		limit := utils.GetValueByPointerOrDefaultValue(a.Limit, math.MaxInt)
		offset := utils.GetValueByPointerOrDefaultValue(a.Offset, 0)
		orderBy := utils.GetValueByPointerOrDefaultValue(a.OrderBy, utils.AscOrder)
		orderColumn := utils.GetValueByPointerOrDefaultValue(a.OrderColumn, order.FieldID)
		total, err := repository.TotalActiveAreas(ctx)
		if err != nil {
			area.logger.Error("failed to query total active areas", zap.Error(err))
			return active_areas.NewGetAllActiveAreasDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}
		var e []*ent.ActiveArea
		if total > 0 {
			e, err = repository.AllActiveAreas(ctx, int(limit), int(offset), orderBy, orderColumn)
			if err != nil {
				area.logger.Error("failed to query active areas", zap.Error(err))
				return active_areas.NewGetAllActiveAreasDefault(http.StatusInternalServerError).
					WithPayload(buildErrorPayload(err))
			}
		}
		totalAreas := int64(total)
		listActiveAreas := &models.ListOfActiveAreas{
			Items: make([]*models.ActiveArea, len(e)),
			Total: &totalAreas,
		}
		for i, element := range e {
			id := int64(element.ID)
			listActiveAreas.Items[i] = &models.ActiveArea{ID: &id, Name: &element.Name}
		}
		return active_areas.NewGetAllActiveAreasOK().WithPayload(listActiveAreas)
	}
}
