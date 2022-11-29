package middlewares

import (
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
	status int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	nrw := &ResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
	return nrw
}

func (r *ResponseWriter) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *ResponseWriter) Status() int {
	return r.status
}
