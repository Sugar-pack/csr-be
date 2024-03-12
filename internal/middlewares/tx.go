package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
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
				utils.WriteErrorResponse(w, http.StatusInternalServerError, "Error initiating transaction")
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
