package middlewares

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func TxMiddleware(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if db == nil {
				next.ServeHTTP(w, r)
				return
			}

			tx, err := db.BeginTxx(r.Context(), nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ctx := setTx(r.Context(), tx)
			rr := &responseRecorder{ResponseWriter: w}

			next.ServeHTTP(rr, r.WithContext(ctx))

			if rr.statusCode >= 400 {
				_ = tx.Rollback()
				return
			}

			if err := tx.Commit(); err != nil {
				if !rr.wroteHeader {
					rr.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	wroteHeader bool
	statusCode  int
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.wroteHeader = true
		r.statusCode = code
		r.ResponseWriter.WriteHeader(code)
	}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(b)
}

type txContextKey struct{}

func setTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func GetTx(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txContextKey{}).(*sqlx.Tx)
	return tx
}
