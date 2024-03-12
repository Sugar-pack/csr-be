package handlers

import (
	"net/http"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
)

func buildErrorPayload(code int32, msg string, details string) *models.SwaggerError {
	return &models.SwaggerError{
		Code:    &code,
		Message: &msg,
		Details: details, // optional field for raw err messages
	}
}

func buildInternalErrorPayload(msg string, details string) *models.SwaggerError {
	return buildErrorPayload(http.StatusInternalServerError, msg, details)
}

func buildExFailedErrorPayload(msg string, details string) *models.SwaggerError {
	return buildErrorPayload(http.StatusExpectationFailed, msg, details)
}

func buildConflictErrorPayload(msg string, details string) *models.SwaggerError {
	return buildErrorPayload(http.StatusConflict, msg, details)
}

func buildNotFoundErrorPayload(msg string, details string) *models.SwaggerError {
	return buildErrorPayload(http.StatusNotFound, msg, details)
}

func buildForbiddenErrorPayload(msg string, details string) *models.SwaggerError {
	return buildErrorPayload(http.StatusForbidden, msg, details)
}

func buildBadRequestErrorPayload(msg string, details string) *models.SwaggerError {
	return buildErrorPayload(http.StatusBadRequest, msg, details)
}
