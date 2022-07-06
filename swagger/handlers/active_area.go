package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/active_areas"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

func SetActiveAreaHandler(client *ent.Client, logger *zap.Logger, api *operations.BeAPI) {
	activeAreaRepo := repositories.NewActiveAreaRepository(client)
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

func (area ActiveArea) GetActiveAreasFunc(repository repositories.ActiveAreaRepository) active_areas.GetAllActiveAreasHandlerFunc {
	return func(a active_areas.GetAllActiveAreasParams, access interface{}) middleware.Responder {
		ctx := a.HTTPRequest.Context()
		e, err := repository.AllActiveAreas(ctx)
		if err != nil {
			area.logger.Error("failed to query active areas", zap.Error(err))
			return active_areas.NewGetAllActiveAreasDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		listActiveAreas := models.ListOfActiveAreas{}
		for _, element := range e {
			id := int64(element.ID)
			listActiveAreas = append(listActiveAreas, &models.ActiveArea{ID: &id, Name: &element.Name})
		}
		return active_areas.NewGetAllActiveAreasOK().WithPayload(listActiveAreas)
	}
}
