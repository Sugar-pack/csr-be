package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
)

type contextKey string

const TxContextKey contextKey = "tx"

func TxFromContext(ctx context.Context) (*ent.Tx, error) {
	v, ok := ctx.Value(TxContextKey).(*ent.Tx)
	if !ok {
		return nil, fmt.Errorf("transaction not found")
	}
	return v, nil
}

func Tx(cln *ent.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := NewResponseWriter(w)
			var (
				tx  *ent.Tx
				s   int
				err error
			)
			ctx := r.Context()
			tx, err = cln.Tx(ctx)
			if err != nil {
				writeResponse(w, http.StatusInternalServerError, &models.Error{
					Data: &models.ErrorData{
						// TODO: add correlation id
						CorrelationID: "",
						Message:       "Error initiating transaction",
					},
				})
				return
			}
			defer func() {
				if p := recover(); p != nil {
					tx.Rollback()
					panic(p)
				} else if err != nil {
					tx.Rollback()
				}
			}()
			ctx = context.WithValue(ctx, TxContextKey, tx)
			next.ServeHTTP(rw, r.WithContext(ctx))
			s = rw.Status()
			if s >= http.StatusOK && s <= http.StatusMultipleChoices {
				err = tx.Commit()
			} else {
				err = fmt.Errorf("non successful status code: %d", s)
			}
		})
	}
}

func writeResponse(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"data":{"message":"Internal server error"}}`))
		return
	}
	w.WriteHeader(code)
	w.Write(b)
}
