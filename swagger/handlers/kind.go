package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/kinds"
)

type Kind struct {
	client *ent.Client
	logger *zap.Logger
}

func NewKind(client *ent.Client, logger *zap.Logger) *Kind {
	return &Kind{
		client: client,
		logger: logger,
	}
}

func (c Kind) CreateNewKindFunc() kinds.CreateNewKindHandlerFunc {
	return func(s kinds.CreateNewKindParams) middleware.Responder {
		e, err := c.client.Kind.Create().SetName(*s.Name.Data.Name).Save(s.HTTPRequest.Context())
		if err != nil {
			c.logger.Error("cant create new kind", zap.Error(err))
			return kinds.NewCreateNewKindDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		return kinds.NewCreateNewKindCreated().WithPayload(&models.CreateNewKindResponse{
			Data: &models.Kind{
				ID:   int64(e.ID),
				Name: &e.Name,
			},
		})
	}
}

func (c Kind) GetAllKindsFunc() kinds.GetAllKindsHandlerFunc {
	return func(s kinds.GetAllKindsParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		e, err := c.client.Kind.Query().All(ctx)
		if err != nil {
			c.logger.Error("query all kind error", zap.Error(err))
			return kinds.NewGetAllKindsDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		listOfKinds := models.ListOfKinds{}
		for _, v := range e {
			listOfKinds = append(listOfKinds, &models.Kind{ID: int64(v.ID), Name: &v.Name, MaxReservationTime: v.MaxReservationTime, MaxReservationUnits: v.MaxReservationUnits})
		}
		return kinds.NewGetAllKindsOK().WithPayload(listOfKinds)
	}
}

func (c Kind) GetKindByIDFunc() kinds.GetKindByIDHandlerFunc {
	return func(s kinds.GetKindByIDParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		id, err := strconv.Atoi(s.KindID)
		if err != nil {
			c.logger.Error("failed to convert kindID into string", zap.Error(err))
			return kinds.NewGetKindByIDDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		e, err := c.client.Kind.Get(ctx, id)
		if err != nil {
			c.logger.Error("failed to get kind", zap.Error(err))
			return kinds.NewGetKindByIDDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		return kinds.NewGetKindByIDOK().WithPayload(&models.GetKindByIDResponse{
			Data: &models.Kind{
				ID:                  int64(e.ID),
				Name:                &e.Name,
				MaxReservationTime:  e.MaxReservationTime,
				MaxReservationUnits: e.MaxReservationUnits,
			},
		})
	}
}

func (c Kind) DeleteKindFunc() kinds.DeleteKindHandlerFunc {
	return func(s kinds.DeleteKindParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		id, err := strconv.Atoi(s.KindID)
		if err != nil {
			c.logger.Error("parse KindID into string failed", zap.Error(err))
			return kinds.NewDeleteKindDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		e, err := c.client.Kind.Get(ctx, id)
		if err != nil {
			c.logger.Error("get kind failed", zap.Error(err))
			return kinds.NewDeleteKindDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		err = c.client.Kind.DeleteOneID(e.ID).Exec(ctx)
		if err != nil {
			c.logger.Error("delete kind failed", zap.Error(err))
			return kinds.NewDeleteKindDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		return kinds.NewDeleteKindCreated().WithPayload(&models.DeleteKindResponse{
			Data: &models.Kind{
				ID:                  int64(e.ID),
				Name:                &e.Name,
				MaxReservationTime:  e.MaxReservationTime,
				MaxReservationUnits: e.MaxReservationUnits,
			},
		})
	}
}

func (c Kind) PatchKindFunc() kinds.PatchKindHandlerFunc {
	return func(s kinds.PatchKindParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		id, err := strconv.Atoi(s.KindID)
		if err != nil {
			c.logger.Error("parse KindID into string failed", zap.Error(err))
			return kinds.NewPatchKindDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		e, err := c.client.Kind.Get(ctx, id)
		if err != nil {
			c.logger.Error("get kind failed", zap.Error(err))
			return kinds.NewGetKindByIDDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		maxTime := s.PatchTask.Data.MaxReservationTime
		maxUnits := s.PatchTask.Data.MaxReservationUnits

		if maxTime != 0 {
			e, err = c.client.Kind.UpdateOneID(id).SetMaxReservationTime(maxTime).Save(ctx)
			if err != nil {
				c.logger.Error("update kind failed", zap.Error(err))
				return kinds.NewPatchKindDefault(http.StatusInternalServerError).WithPayload(&models.Error{
					Data: &models.ErrorData{
						Message: err.Error(),
					},
				})
			}
		}
		if maxUnits != 0 {
			e, err = c.client.Kind.UpdateOneID(id).SetMaxReservationUnits(maxUnits).Save(ctx)
			if err != nil {
				c.logger.Error("update kind failed", zap.Error(err))
				return kinds.NewPatchKindDefault(http.StatusInternalServerError).WithPayload(&models.Error{
					Data: &models.ErrorData{
						Message: err.Error(),
					},
				})
			}
		}

		return kinds.NewPatchKindCreated().WithPayload(&models.PatchKindResponse{
			Data: &models.Kind{
				ID:                  int64(id),
				Name:                &e.Name,
				MaxReservationTime:  e.MaxReservationTime,
				MaxReservationUnits: e.MaxReservationUnits,
			},
		})
	}
}
