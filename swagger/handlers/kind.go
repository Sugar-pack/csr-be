package handlers

import (
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"math"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/kinds"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

func SetKindHandler(client *ent.Client, logger *zap.Logger, api *operations.BeAPI) {
	kindRepo := repositories.NewKindRepository(client)
	kindHandler := NewKind(logger)

	api.KindsCreateNewKindHandler = kindHandler.CreateNewKindFunc(kindRepo)
	api.KindsGetKindByIDHandler = kindHandler.GetKindByIDFunc(kindRepo)
	api.KindsDeleteKindHandler = kindHandler.DeleteKindFunc(kindRepo)
	api.KindsGetAllKindsHandler = kindHandler.GetAllKindsFunc(kindRepo)
	api.KindsPatchKindHandler = kindHandler.PatchKindFunc(kindRepo)
}

type Kind struct {
	logger *zap.Logger
}

func NewKind(logger *zap.Logger) *Kind {
	return &Kind{
		logger: logger,
	}
}

func (c *Kind) CreateNewKindFunc(repository repositories.KindRepository) kinds.CreateNewKindHandlerFunc {
	return func(s kinds.CreateNewKindParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		createdKind, err := repository.CreateKind(ctx, *s.NewKind)
		if err != nil {
			c.logger.Error("cant create new kind", zap.Error(err))
			return kinds.NewCreateNewKindDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("cant create new kind"))
		}
		return kinds.NewCreateNewKindCreated().WithPayload(&models.CreateNewKindResponse{
			Data: mapKind(createdKind),
		})
	}
}

func (c *Kind) GetAllKindsFunc(repository repositories.KindRepository) kinds.GetAllKindsHandlerFunc {
	return func(s kinds.GetAllKindsParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		limit := utils.GetParamInt(s.Limit, math.MaxInt)
		offset := utils.GetParamInt(s.Offset, 0)
		orderBy := utils.GetParamString(s.OrderBy, utils.AscOrder)
		orderColumn := utils.GetParamString(s.OrderColumn, kind.FieldID)
		total, err := repository.AllKindsTotal(ctx)
		if err != nil {
			c.logger.Error("query total kinds error", zap.Error(err))
			return kinds.NewGetAllKindsDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("cant get total amount of kinds"))
		}
		var allKinds []*ent.Kind
		if total > 0 {
			allKinds, err = repository.AllKinds(ctx, limit, offset, orderBy, orderColumn)
			if err != nil {
				c.logger.Error("query all kind error", zap.Error(err))
				return kinds.NewGetAllKindsDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("cant get all kinds"))
			}
		}
		mappedKinds := make([]*models.Kind, len(allKinds))
		for i, v := range allKinds {
			mappedKinds[i] = mapKind(v)
		}
		totalKinds := int64(total)
		listOfKinds := &models.ListOfKinds{
			Items: mappedKinds,
			Total: &totalKinds,
		}
		return kinds.NewGetAllKindsOK().WithPayload(listOfKinds)
	}
}

func (c *Kind) GetKindByIDFunc(repository repositories.KindRepository) kinds.GetKindByIDHandlerFunc {
	return func(s kinds.GetKindByIDParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		kind, err := repository.KindByID(ctx, int(s.KindID))
		if err != nil {
			c.logger.Error("failed to get kind", zap.Error(err))
			return kinds.NewGetKindByIDDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("failed to get kind"))
		}
		return kinds.NewGetKindByIDOK().WithPayload(&models.GetKindByIDResponse{
			Data: mapKind(kind),
		})
	}
}

func (c *Kind) DeleteKindFunc(repository repositories.KindRepository) kinds.DeleteKindHandlerFunc {
	return func(s kinds.DeleteKindParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		err := repository.DeleteKindByID(ctx, int(s.KindID))
		if err != nil {
			c.logger.Error("delete kind failed", zap.Error(err))
			return kinds.NewDeleteKindDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("delete kind failed"))
		}
		return kinds.NewDeleteKindOK().WithPayload("kind deleted")
	}
}

func (c *Kind) PatchKindFunc(repository repositories.KindRepository) kinds.PatchKindHandlerFunc {
	return func(s kinds.PatchKindParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		updatedKind, err := repository.UpdateKind(ctx, int(s.KindID), *s.PatchKind)
		if err != nil {
			c.logger.Error("cant update kind", zap.Error(err))
			return kinds.NewPatchKindDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("cant update kind"))
		}

		return kinds.NewPatchKindOK().WithPayload(&models.PatchKindResponse{
			Data: mapKind(updatedKind),
		})
	}
}

func mapKind(kind *ent.Kind) *models.Kind {
	return &models.Kind{
		ID:                  int64(kind.ID),
		Name:                &kind.Name,
		MaxReservationTime:  kind.MaxReservationTime,
		MaxReservationUnits: kind.MaxReservationUnits,
	}
}
