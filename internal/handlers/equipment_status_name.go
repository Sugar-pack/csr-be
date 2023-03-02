package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	eqStatusName "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/equipment_status_name"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func SetEquipmentStatusNameHandler(logger *zap.Logger, api *operations.BeAPI) {
	equipmentStatusNameRepo := repositories.NewEquipmentStatusNameRepository()
	statusHandler := NewEquipmentStatusName(logger)

	api.EquipmentStatusNamePostEquipmentStatusNameHandler = statusHandler.PostEquipmentStatusNameFunc(equipmentStatusNameRepo)
	api.EquipmentStatusNameListEquipmentStatusNamesHandler = statusHandler.ListEquipmentStatusNamesFunc(equipmentStatusNameRepo)
	api.EquipmentStatusNameGetEquipmentStatusNameHandler = statusHandler.GetEquipmentStatusNameFunc(equipmentStatusNameRepo)
	api.EquipmentStatusNameDeleteEquipmentStatusNameHandler = statusHandler.DeleteEquipmentStatusNameFunc(equipmentStatusNameRepo)
}

type EquipmentStatusName struct {
	logger *zap.Logger
}

func NewEquipmentStatusName(logger *zap.Logger) *EquipmentStatusName {
	return &EquipmentStatusName{
		logger: logger,
	}
}

func (c EquipmentStatusName) PostEquipmentStatusNameFunc(repository domain.EquipmentStatusNameRepository) eqStatusName.PostEquipmentStatusNameHandlerFunc {
	return func(s eqStatusName.PostEquipmentStatusNameParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		name := s.Name.Name
		createdStatus, err := repository.Create(ctx, *name)
		if err != nil {
			c.logger.Error("create status failed", zap.Error(err))
			return eqStatusName.NewPostEquipmentStatusNameDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't create status"))
		}

		return eqStatusName.NewPostEquipmentStatusNameCreated().WithPayload(&models.SuccessEquipmentStatusNameOperationResponse{
			Data: mapEquipmentStatusName(createdStatus),
		})
	}
}

func (c EquipmentStatusName) ListEquipmentStatusNamesFunc(repository domain.EquipmentStatusNameRepository) eqStatusName.ListEquipmentStatusNamesHandlerFunc {
	return func(s eqStatusName.ListEquipmentStatusNamesParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		statuses, err := repository.GetAll(ctx)
		if err != nil {
			c.logger.Error("get statuses failed", zap.Error(err))
			return eqStatusName.NewListEquipmentStatusNamesDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't get statuses"))
		}
		listStatuses := models.ListEquipmentStatusNames{}
		for _, element := range statuses {
			listStatuses = append(listStatuses, mapEquipmentStatusName(element))
		}
		return eqStatusName.NewListEquipmentStatusNamesOK().WithPayload(listStatuses)
	}
}

func (c EquipmentStatusName) GetEquipmentStatusNameFunc(repository domain.EquipmentStatusNameRepository) eqStatusName.GetEquipmentStatusNameHandlerFunc {
	return func(s eqStatusName.GetEquipmentStatusNameParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		foundStatus, err := repository.Get(ctx, int(s.StatusID))
		if err != nil {
			c.logger.Error("get status failed", zap.Error(err))
			return eqStatusName.NewGetEquipmentStatusNameDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't get status"))
		}

		return eqStatusName.NewGetEquipmentStatusNameOK().WithPayload(&models.SuccessEquipmentStatusNameOperationResponse{
			Data: mapEquipmentStatusName(foundStatus),
		})
	}
}

func (c EquipmentStatusName) DeleteEquipmentStatusNameFunc(repository domain.EquipmentStatusNameRepository) eqStatusName.DeleteEquipmentStatusNameHandlerFunc {
	return func(s eqStatusName.DeleteEquipmentStatusNameParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		deletedStatus, err := repository.Delete(ctx, int(s.StatusID))
		if err != nil {
			c.logger.Error("delete status failed", zap.Error(err))
			return eqStatusName.NewDeleteEquipmentStatusNameDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("can't delete status"))
		}
		return eqStatusName.NewDeleteEquipmentStatusNameOK().WithPayload(
			&models.SuccessEquipmentStatusNameOperationResponse{
				Data: mapEquipmentStatusName(deletedStatus),
			},
		)
	}
}

func mapEquipmentStatusName(status *ent.EquipmentStatusName) *models.EquipmentStatusName {
	id := int64(status.ID)
	return &models.EquipmentStatusName{
		ID:   id,
		Name: &status.Name,
	}
}
