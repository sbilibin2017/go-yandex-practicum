package middlewares

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgconn" // импортируем твой пакет с WithRetry
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
		err := withRetry(r.Context(), delays, func(ctx context.Context) error {
			return next(w, r.WithContext(ctx))
		})
		if err != nil {
			logger.Log.Error("Handler failed after retries", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func withRetry(ctx context.Context, delays []time.Duration, fn func(ctx context.Context) error) error {
	var err error
	for attempt := 0; ; attempt++ {
		err = fn(ctx)
		if err == nil || !isRetriableDBError(err) {
			return err
		}

		if attempt >= len(delays) {
			break
		}

		select {
		case <-time.After(delays[attempt]):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
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
