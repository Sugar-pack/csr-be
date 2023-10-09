package utils

import (
	"encoding/json"
	"net/http"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
)

func WriteErrorResponse(w http.ResponseWriter, code int32, message string) {
	w.Header().Set("Content-Type", "application/json")
	e := models.SwaggerError{
		Code:    &code,
		Message: &message,
	}
	b, err := json.Marshal(e)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"Unexpected error"}`))
		return
	}
	w.WriteHeader(int(code))
	w.Write(b)
}
