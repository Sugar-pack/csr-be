package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/kinds"
	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"
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
			return kinds.NewCreateNewKindDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		id := fmt.Sprintf("%d", e.ID)
		return kinds.NewCreateNewKindCreated().WithPayload(&models.CreateNewKindResponse{
			Data: &models.Kind{
				ID:   &id,
				Name: &e.Name,
			},
		})
	}
}

func (c Kind) GetAllKindsFunc() kinds.GetAllKindsHandlerFunc {
	return func(s kinds.GetAllKindsParams) middleware.Responder {
		e, err := c.client.Kind.Query().All(s.HTTPRequest.Context())
		if err != nil {
			return kinds.NewGetAllKindsDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		listOfKinds := models.ListOfKinds{}
		for _, v := range e {
			id := strconv.Itoa(v.ID)
			listOfKinds = append(listOfKinds, &models.Kind{ID: &id, Name: &v.Name})
		}
		return kinds.NewGetAllKindsOK().WithPayload(listOfKinds)
	}
}

func (c Kind) GetKindByIDFunc() kinds.GetKindByIDHandlerFunc {
	return func(s kinds.GetKindByIDParams) middleware.Responder {
		id, err := strconv.Atoi(s.KindID)
		if err != nil {
			return kinds.NewGetKindByIDDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		e, err := c.client.Kind.Get(s.HTTPRequest.Context(), id)
		if err != nil {
			return kinds.NewGetKindByIDDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		return kinds.NewGetKindByIDOK().WithPayload(&models.GetKindByIDResponse{
			Data: &models.Kind{
				ID:   &s.KindID,
				Name: &e.Name,
			},
		})
	}
}

func (c Kind) DeleteKindFunc() kinds.DeleteKindHandlerFunc {
	return func(s kinds.DeleteKindParams) middleware.Responder {
		id, err := strconv.Atoi(s.KindID)

		e, err := c.client.Kind.Get(s.HTTPRequest.Context(), id)
		if err != nil {
			return kinds.NewDeleteKindDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		err = c.client.Kind.DeleteOneID(e.ID).Exec(s.HTTPRequest.Context())
		if err != nil {
			return kinds.NewDeleteKindDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		return kinds.NewDeleteKindCreated().WithPayload(&models.DeleteKindResponse{
			Data: &models.Kind{
				ID:   &s.KindID,
				Name: &e.Name,
			},
		})
	}
}
