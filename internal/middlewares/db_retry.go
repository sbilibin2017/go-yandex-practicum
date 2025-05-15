package middlewares

import (
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgconn"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

type HandlerWithContext func(w http.ResponseWriter, r *http.Request) error

func DBRetryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerWithCtx := func(w http.ResponseWriter, r *http.Request) error {
			next.ServeHTTP(w, r)
			return nil
		}
		dbRetryMiddleware(handlerWithCtx).ServeHTTP(w, r)
	})
}

func dbRetryMiddleware(next HandlerWithContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
		var err error

		ctx := r.Context()

		for attempt := 0; attempt <= len(delays); attempt++ {
			err = next(w, r.WithContext(ctx))
			if err == nil || !isRetriableDBError(err) {
				break
			}

			if attempt < len(delays) {
				logger.Log.Warn("Retriable DB error, retrying handler",
					zap.Int("attempt", attempt+1),
					zap.Duration("next_delay", delays[attempt]),
					zap.Error(err),
				)
				select {
				case <-time.After(delays[attempt]):
				case <-ctx.Done():
					http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
					return
				}
			}
		}

		if err != nil {
			logger.Log.Error("Handler failed after retries", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func isRetriableDBError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && len(pgErr.Code) >= 2 && pgErr.Code[:2] == "08" {
		return true
	}
	return false
}
