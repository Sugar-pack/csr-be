package handlers

import "git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"

func buildErrorPayload(err error) *models.Error {
	return &models.Error{
		Data: &models.ErrorData{
			Message: err.Error(),
		},
	}
}
