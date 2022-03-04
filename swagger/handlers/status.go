package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/status"
	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"
)

type Status struct {
	client *ent.Client
	logger *zap.Logger
}

func NewStatus(client *ent.Client, logger *zap.Logger) *Status {
	return &Status{
		client: client,
		logger: logger,
	}
}

func (c Status) PostStatusFunc() status.PostStatusHandlerFunc {
	return func(s status.PostStatusParams) middleware.Responder {
		e, err := c.client.Statuses.Create().SetName(*s.Name.Name).Save(s.HTTPRequest.Context())
		if err != nil {
			return status.NewPostStatusDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		id := fmt.Sprintf("%d", e.ID)
		return status.NewPostStatusCreated().WithPayload(&models.SuccessStatusOperationResponse{
			Data: &models.Status{
				ID:   &id,
				Name: &e.Name,
			},
		})
	}
}

func (c Status) GetStatusesFunc() status.GetStatusesHandlerFunc {
	return func(s status.GetStatusesParams) middleware.Responder {
		e, err := c.client.Statuses.Query().All(s.HTTPRequest.Context())
		if err != nil {
			return status.NewGetStatusesDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		listStatuses := models.ListStatuses{}
		for _, element := range e {
			id := strconv.Itoa(element.ID)
			listStatuses = append(listStatuses, &models.Status{&id, &element.Name})
		}
		return status.NewGetStatusesCreated().WithPayload(listStatuses)
	}
}

func (c Status) GetStatusFunc() status.GetStatusHandlerFunc {
	return func(s status.GetStatusParams) middleware.Responder {
		statusId, err := strconv.Atoi(s.StatusID)
		if err != nil {
			return status.NewGetStatusDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		e, err := c.client.Statuses.Get(s.HTTPRequest.Context(), statusId)
		if err != nil {
			return status.NewGetStatusDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		return status.NewGetStatusCreated().WithPayload(&models.SuccessStatusOperationResponse{
			Data: &models.Status{
				ID:   &s.StatusID,
				Name: &e.Name,
			},
		})
	}
}

func (c Status) DeleteStatusFunc() status.DeleteStatusHandlerFunc {
	return func(s status.DeleteStatusParams) middleware.Responder {
		statusId, err := strconv.Atoi(s.StatusID)
		if err != nil {
			return status.NewDeleteStatusDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		e, err := c.client.Statuses.Get(s.HTTPRequest.Context(), statusId)
		if err != nil {
			return status.NewDeleteStatusDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		deleteErr := c.client.Statuses.DeleteOne(e).Exec(s.HTTPRequest.Context())
		if deleteErr != nil {
			return status.NewDeleteStatusDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: deleteErr.Error(),
				},
			})
		}

		return status.NewGetStatusCreated().WithPayload(&models.SuccessStatusOperationResponse{
			Data: &models.Status{
				ID:   &s.StatusID,
				Name: &e.Name,
			},
		})
	}
}
