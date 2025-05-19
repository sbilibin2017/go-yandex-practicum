package retrier

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgconn"
)

// WithRetry выполняет функцию fn с попытками повторного вызова при ошибках,
// которые считаются повторяемыми по заданным функциям isRetriableErrorFuncs.
// ctx — контекст для отмены или таймаута.
// attempts — срез длительностей ожидания между попытками (например, [100ms, 200ms, 500ms]).
// fn — функция, которую необходимо выполнить с возможным повтором.
// isRetriableErrorFuncs — переменное число функций, каждая из которых принимает ошибку и возвращает true,
// если ошибка считается повторяемой (требует повторного выполнения).
// Возвращает nil при успешном выполнении fn, либо последнюю ошибку, если попытки исчерпаны.
func WithRetry(
	ctx context.Context,
	attempts []time.Duration,
	fn func(ctx context.Context) error,
	isRetriableErrorFuncs ...func(err error) bool,
) error {
	var lastErr error
	total := len(attempts)

	isRetriableErrorFunc := func(err error) bool {
		if len(isRetriableErrorFuncs) == 0 {
			return false
		}
		for _, f := range isRetriableErrorFuncs {
			if f(err) {
				return true
			}
		}
		return false
	}

	for i := 0; i <= total; i++ {
		err := fn(ctx)
		if err == nil || !isRetriableErrorFunc(err) {
			return err
		}

		lastErr = err

		if i == total {
			break
		}

		select {
		case <-time.After(attempts[i]):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

// IsRetriableDBError определяет, является ли ошибка базы данных повторяемой.
// Проверяет, является ли ошибка ошибкой PostgreSQL с кодом, начинающимся на "08",
// что обычно указывает на ошибки соединения (например, потеря соединения).
// err — ошибка для проверки.
// Возвращает true, если ошибка считается повторяемой, иначе false.
func IsRetriableDBError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && len(pgErr.Code) >= 2 && pgErr.Code[:2] == "08" {
		return true
	}
	return false
}
