package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/active_areas"
)

type ActiveArea struct {
	client *ent.Client
	logger *zap.Logger
}

func NewActiveArea(client *ent.Client, logger *zap.Logger) *ActiveArea {
	return &ActiveArea{
		client: client,
		logger: logger,
	}
}

func (area ActiveArea) GetActiveAreasFunc() active_areas.GetAllActiveAreasHandlerFunc {
	return func(a active_areas.GetAllActiveAreasParams) middleware.Responder {
		ctx := a.HTTPRequest.Context()
		e, err := area.client.ActiveArea.Query().Order(ent.Asc("id")).All(ctx)
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
