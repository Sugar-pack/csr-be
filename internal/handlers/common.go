package handlers

import "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"

func buildErrorPayload(err error) *models.Error {
	return &models.Error{
		Data: &models.ErrorData{
			Message: err.Error(),
		},
	}
}

func buildStringPayload(msg string) *models.Error {
	return &models.Error{
		Data: &models.ErrorData{
			Message: msg,
		},
	}
}
