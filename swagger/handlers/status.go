package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/status"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

type Status struct {
	logger *zap.Logger
}

func NewStatus(logger *zap.Logger) *Status {
	return &Status{
		logger: logger,
	}
}

func (c Status) PostStatusFunc(repository repositories.EquipmentStatusRepository) status.PostStatusHandlerFunc {
	return func(s status.PostStatusParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		name := s.Name.Name
		createdStatus, err := repository.Create(ctx, *name)
		if err != nil {
			c.logger.Error("create status failed", zap.Error(err))
			return status.NewPostStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't create status"))
		}

		return status.NewPostStatusCreated().WithPayload(&models.SuccessStatusOperationResponse{
			Data: mapStatus(createdStatus),
		})
	}
}

func (c Status) GetStatusesFunc(repository repositories.EquipmentStatusRepository) status.GetStatusesHandlerFunc {
	return func(s status.GetStatusesParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		statuses, err := repository.GetAll(ctx)
		if err != nil {
			c.logger.Error("get statuses failed", zap.Error(err))
			return status.NewGetStatusesDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't get statuses"))
		}
		listStatuses := models.ListStatuses{}
		for _, element := range statuses {
			listStatuses = append(listStatuses, mapStatus(element))
		}
		return status.NewGetStatusesOK().WithPayload(listStatuses)
	}
}

func (c Status) GetStatusFunc(repository repositories.EquipmentStatusRepository) status.GetStatusHandlerFunc {
	return func(s status.GetStatusParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		foundStatus, err := repository.Get(ctx, int(s.StatusID))
		if err != nil {
			c.logger.Error("get status failed", zap.Error(err))
			return status.NewGetStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't get status"))
		}

		return status.NewGetStatusOK().WithPayload(&models.SuccessStatusOperationResponse{
			Data: mapStatus(foundStatus),
		})
	}
}

func (c Status) DeleteStatusFunc(repository repositories.EquipmentStatusRepository) status.DeleteStatusHandlerFunc {
	return func(s status.DeleteStatusParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		deletedStatus, err := repository.Delete(ctx, int(s.StatusID))
		if err != nil {
			c.logger.Error("delete status failed", zap.Error(err))
			return status.NewDeleteStatusDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't delete status"))
		}
		return status.NewGetStatusOK().WithPayload(&models.SuccessStatusOperationResponse{
			Data: mapStatus(deletedStatus),
		})
	}
}

func mapStatus(status *ent.Statuses) *models.Status {
	id := int64(status.ID)
	return &models.Status{
		ID:   &id,
		Name: &status.Name,
	}
}
